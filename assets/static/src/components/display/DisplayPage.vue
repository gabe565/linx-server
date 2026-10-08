<template>
  <div class="container mx-auto" :class="maxWidth">
    <div v-if="isLoading" class="animate-in fade-in duration-1000 flex flex-col items-center">
      <SpinnerIcon class="text-4xl" />
    </div>

    <form v-if="errorStatus === 401" @submit.prevent="() => execute()">
      <Card>
        <CardHeader>
          <CardTitle>Authentication Required</CardTitle>
          <CardDescription>This file is password-protected.</CardDescription>
        </CardHeader>
        <CardContent class="space-y-2">
          <Label for="password" :class="{ 'text-destructive': downloadAttempts > 1 }">
            Password
          </Label>
          <Input
            id="password"
            type="password"
            v-model="accessKey"
            placeholder="Enter password"
            class="flex-1 min-w-50"
            autofocus
          />
          <span
            v-if="downloadAttempts > 1"
            class="font-medium text-destructive text-sm"
            role="alert"
          >
            Invalid password
          </span>
        </CardContent>
        <CardFooter class="flex flex-col items-end">
          <Button type="submit" class="w-full sm:w-auto">Unlock</Button>
        </CardFooter>
      </Card>
    </form>

    <Card v-else-if="errorStatus === 404">
      <CardHeader>
        <CardTitle>Oops! You found a Dead Link</CardTitle>
        <CardDescription>This file has expired or does not exist.</CardDescription>
      </CardHeader>
      <CardContent class="flex flex-col items-center">
        <img :src="DeadLink" alt="Dead Link" class="max-w-80 py-10" />
      </CardContent>
    </Card>

    <Card v-else-if="burnState === 'downloaded'">
      <CardHeader>
        <CardTitle>File Deleted</CardTitle>
        <CardDescription>
          This file has been downloaded and deleted from the server.
        </CardDescription>
      </CardHeader>
    </Card>

    <Card v-else-if="error">
      <CardHeader>
        <CardTitle>Error</CardTitle>
        <CardDescription> An error occurred while loading the file: {{ message }} </CardDescription>
      </CardHeader>
    </Card>

    <BurnNotice
      v-else-if="state.meta && burnState === 'pending'"
      :meta="state.meta"
      :preview="canPreviewBurn"
      :loading="isRevealing"
      @view="reveal"
      @download="onBurnDownload"
    />

    <Card v-else-if="state.meta">
      <FileHeader v-model:wrap="wrap" :state="state" />
      <FileViewer :state="state" :wrap="wrap" />
    </Card>
  </div>
</template>

<script setup lang="ts">
import Modes from "./fileModes.ts";
import { useAsyncState } from "@vueuse/core";
import axios, { isAxiosError } from "axios";
import { computed, onBeforeUnmount, ref } from "vue";
import { toast } from "vue-sonner";
import DeadLink from "@/assets/dead-link.svg";
import BurnNotice from "@/components/display/BurnNotice.vue";
import FileHeader from "@/components/display/FileHeader.vue";
import FileViewer from "@/components/display/FileViewer.vue";
import { Button } from "@/components/ui/button/index.js";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card/index.js";
import { Input } from "@/components/ui/input/index.js";
import { Label } from "@/components/ui/label/index.js";
import { ApiPath } from "@/config/api.ts";
import { useConfigStore } from "@/stores/config.ts";
import { getExtension, loadLanguage } from "@/util/extensions.ts";
import SpinnerIcon from "~icons/svg-spinners/ring-resize";

const props = defineProps({
  filename: { type: String, required: true },
});

const config = useConfigStore();

document.title = props.filename + " · " + config.site.site_name;

const accessKey = ref();
const encAccessKey = computed(() =>
  accessKey.value ? encodeURIComponent(accessKey.value) : undefined,
);
const downloadAttempts = ref(0);
const wrap = ref(true);

type DisplayMeta = Record<string, any> & {
  filename: string;
  mimetype: string;
  size: number;
  original_name?: string;
  language?: string;
  direct_url: string;
  download_url: string;
  archive_files?: string[];
  expiry?: number;
  burn_after_read?: boolean;
};

type DisplayState = {
  meta: DisplayMeta | null;
  mode: symbol | null;
  content: string | null;
};

const MaxBurnPreviewSize = 100 * 1024 * 1024;
const MaxTextSize = 512 * 1024;

const isTextMode = (mode: symbol | null) =>
  mode === Modes.TEXT || mode === Modes.MARKDOWN || mode === Modes.CSV;

const burnState = ref<"pending" | "viewed" | "downloaded" | null>(null);
const isRevealing = ref(false);
const revealError = ref<unknown>();
let objectURL: string | undefined;

const {
  state,
  isLoading,
  error: loadError,
  execute,
} = useAsyncState<DisplayState>(
  async () => {
    downloadAttempts.value += 1;
    let res;
    try {
      res = await axios.get(ApiPath(`/${props.filename}`), {
        headers: {
          Accept: "application/json",
          "Linx-Access-Key": encAccessKey.value,
        },
        validateStatus: (s) => s === 200,
        withCredentials: true,
      });
    } catch (err) {
      console.error(err);
      if (downloadAttempts.value > 1) {
        const msg = err instanceof Error ? err.message : String(err);
        toast.error("Failed to load file", { description: msg });
      }
      throw err;
    }

    const meta = res.data;

    if (meta.original_name) {
      document.title = meta.original_name + " · " + config.site.site_name;
    }

    let mode: symbol | undefined;
    if (meta.mimetype.startsWith("image/")) {
      mode = Modes.IMAGE;
    } else if (meta.mimetype.startsWith("audio/")) {
      mode = Modes.AUDIO;
    } else if (meta.mimetype.startsWith("video/")) {
      mode = Modes.VIDEO;
    } else if (meta.mimetype === "application/pdf") {
      mode = Modes.PDF;
    } else if (getExtension(meta.filename) === "md") {
      mode = Modes.MARKDOWN;
    } else if (meta.mimetype === "text/csv") {
      mode = Modes.CSV;
    } else if (meta.archive_files) {
      mode = Modes.ARCHIVE;
    } else if (meta.mimetype.startsWith("text/") || !!meta.language) {
      mode = Modes.TEXT;
    }

    if (meta.burn_after_read) {
      burnState.value = "pending";
      return { meta, mode: mode ?? null, content: null };
    }

    let content: string | undefined;
    if (meta.size < MaxTextSize && isTextMode(mode ?? null)) {
      // Presigned URLs encode auth in the query string and live on a different
      // origin; sending the access-key header or credentials cross-origin breaks
      // CORS preflight.
      const sameOrigin = new URL(meta.direct_url, location.href).origin === location.origin;
      try {
        const res = await Promise.all([
          axios.get(meta.direct_url, {
            headers: sameOrigin ? { "Linx-Access-Key": encAccessKey.value } : {},
            responseType: "text",
            validateStatus: (s) => s === 200,
            withCredentials: sameOrigin,
          }),
          loadLanguage(meta.language),
        ]);
        content = res[0].data;
      } catch (err) {
        console.error(err);
        const msg = err instanceof Error ? err.message : String(err);
        toast.error("Failed to download file", { description: msg });
        throw err;
      }
    }

    return {
      meta,
      mode: mode ?? null,
      content: content ?? null,
    };
  },
  { meta: null, mode: null, content: null },
);

const error = computed(() => loadError.value ?? revealError.value);

const canPreviewBurn = computed(
  () => state.value.mode !== null && (state.value.meta?.size ?? 0) <= MaxBurnPreviewSize,
);

const reveal = async () => {
  const { meta, mode } = state.value;
  if (!meta || isRevealing.value) return;
  isRevealing.value = true;
  try {
    const [res] = await Promise.all([
      axios.get<Blob>(meta.direct_url, {
        headers: { "Linx-Access-Key": encAccessKey.value },
        responseType: "blob",
        validateStatus: (s) => s === 200,
        withCredentials: true,
      }),
      loadLanguage(meta.language ?? ""),
    ]);
    const blob = res.data;
    let content: string | null = null;
    if (blob.size < MaxTextSize && isTextMode(mode)) {
      content = await blob.text();
    }
    objectURL = URL.createObjectURL(blob);
    state.value = {
      meta: { ...meta, direct_url: objectURL, download_url: objectURL },
      mode,
      content,
    };
    burnState.value = "viewed";
  } catch (err) {
    console.error(err);
    if (isAxiosError(err) && err.response?.status === 404) {
      revealError.value = err;
    } else {
      const msg = err instanceof Error ? err.message : String(err);
      toast.error("Failed to download file", { description: msg });
    }
  } finally {
    isRevealing.value = false;
  }
};

const onBurnDownload = () => {
  // Let the download start before the link is removed
  setTimeout(() => (burnState.value = "downloaded"));
};

onBeforeUnmount(() => {
  if (objectURL) URL.revokeObjectURL(objectURL);
});

const errorStatus = computed(() => (isAxiosError(error.value) ? error.value.status : undefined));

const maxWidth = computed(() => {
  if (errorStatus.value === 401) return "max-w-lg";
  if (error.value) return "max-w-xl";
  if (burnState.value === "pending" || burnState.value === "downloaded") return "max-w-xl";
  if (!state.value.meta) return "max-w-5xl";
  // Code wants every pixel; a file with no preview is just a download button.
  if (state.value.mode === Modes.TEXT) return "max-w-7xl";
  if (state.value.mode === null) return "max-w-xl";
  return "max-w-5xl";
});

const message = computed(() => {
  const err = error.value;
  let msg = err instanceof Error ? err.message : String(err);
  if (isAxiosError(err) && err.response?.data?.error) msg = err.response.data.error;
  return msg;
});
</script>
