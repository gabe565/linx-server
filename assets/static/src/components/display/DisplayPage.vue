<template>
  <div class="container mx-auto" :class="[error?.status === 401 ? 'max-w-lg' : 'max-w-5xl']">
    <div v-if="isLoading" class="animate-in fade-in duration-1000 flex flex-col items-center">
      <SpinnerIcon class="text-4xl" />
    </div>

    <form v-if="error?.status === 401" @submit.prevent="execute">
      <Card>
        <CardHeader>
          <CardTitle>Authentication Required</CardTitle>
        </CardHeader>
        <CardContent class="flex flex-col gap-6">
          <p>{{ filename }} is protected with a password:</p>

          <Label class="w-full flex flex-wrap">
            Password
            <Input
              type="password"
              v-model="accessKey"
              placeholder="Enter password"
              class="flex-1 min-w-50"
              autofocus
            />
          </Label>
        </CardContent>
        <CardFooter class="flex flex-col items-end">
          <Button type="submit" class="w-full sm:w-auto">Unlock</Button>
        </CardFooter>
      </Card>
    </form>

    <Card v-else-if="error?.status === 404">
      <CardHeader>
        <CardTitle>Oops! You found a Dead Link</CardTitle>
        <CardDescription>This file has expired or does not exist.</CardDescription>
      </CardHeader>
      <CardContent class="flex flex-col items-center">
        <img :src="DeadLink" alt="Dead Link" class="max-w-80 py-10" />
      </CardContent>
    </Card>

    <Card v-else-if="error">
      <CardHeader>
        <CardTitle>Error</CardTitle>
        <CardDescription>
          An error occurred while loading the file: {{ error.response?.data?.error || message }}
        </CardDescription>
      </CardHeader>
    </Card>

    <Card v-else-if="state.meta">
      <CardHeader class="flex flex-wrap justify-between items-center gap-4">
        <div class="flex flex-col gap-1 max-w-full">
          <CardTitle class="wrap-break-word">{{
            state.meta.original_name || state.meta.filename
          }}</CardTitle>

          <UseTimeAgo
            v-if="expiry"
            v-slot="{ timeAgo }"
            :time="expiry"
            :show-second="true"
            update-interval="1000"
          >
            <CardDescription class="text-xs tabular-nums">
              {{ expired ? "expired" : "expires" }} {{ timeAgo }}
            </CardDescription>
          </UseTimeAgo>
        </div>

        <div v-if="showWrapSwitch" class="flex items-center space-x-2">
          <Switch id="opt-wrap" v-model="wrap" />
          <Label for="opt-wrap">Wrap</Label>
        </div>

        <div
          class="flex flex-col sm:flex-row gap-4 sm:gap-0 flex-wrap mx-auto sm:mx-0 w-full sm:w-auto"
        >
          <EditButton
            v-if="showEditButton"
            :meta="state.meta"
            :content="state.content"
            class="sm:rounded-r-none"
          />
          <DownloadButton
            v-if="state.meta"
            :meta="state.meta"
            :class="{
              'sm:border-l-0 sm:rounded-l-none': showEditButton,
            }"
            :disabled="expired"
          />
        </div>
      </CardHeader>
      <CardContent class="flex flex-col justify-center">
        <div
          v-if="state.mode === Modes.IMAGE"
          class="mx-auto"
          :class="{ 'w-full h-full flex flex-col': state.meta.mimetype === 'image/svg+xml' }"
        >
          <img :src="state.meta.direct_url" alt="" class="max-w-full h-auto rounded-md" />
        </div>

        <div v-else-if="state.mode === Modes.AUDIO">
          <audio controls preload="metadata" class="w-full rounded-md">
            <source :src="state.meta.direct_url" :type="state.meta.mimetype" />
            Your browser doesn’t support playing {{ state.meta.mimetype }}.
          </audio>
        </div>

        <div v-else-if="state.mode === Modes.VIDEO">
          <video controls preload="metadata" class="w-full rounded-md">
            <source :src="state.meta.direct_url" :type="state.meta.mimetype" />
            Your browser doesn’t support playing {{ state.meta.mimetype }}.
          </video>
        </div>

        <div v-else-if="state.mode === Modes.PDF">
          <object
            class="w-full h-[800px] rounded-md"
            :data="state.meta.direct_url"
            :type="state.meta.mimetype"
          >
            Your web browser does not support displaying PDF files.
          </object>
        </div>

        <div
          v-else-if="!!state.formatted && state.mode === Modes.MARKDOWN"
          class="prose max-w-none"
          v-html="state.formatted"
        />

        <pre v-else-if="state.mode === Modes.ARCHIVE" class="overflow-x-scroll max-h-[600px]">{{
          state.meta.archive_files.join("\n")
        }}</pre>

        <div v-else-if="!!state.formatted && state.mode === Modes.CSV" class="space-y-4">
          <Table>
            <TableRow v-for="(row, key) in state.formatted?.data?.slice(0, csvRows)" :key="key">
              <TableCell v-for="(cell, key) in row" :key="key">{{ cell }}</TableCell>
            </TableRow>
          </Table>
          <div class="flex justify-between">
            Showing {{ Math.min(csvRows, state.formatted?.data?.length) }} of
            {{ state.formatted?.data?.length }} rows
            <Button v-if="csvRows < state.formatted?.data?.length" @click="csvRows += 250"
              >Show more</Button
            >
          </div>
        </div>

        <HighlightJS
          v-else-if="!!state.content && state.mode === Modes.TEXT"
          class="p-4"
          :class="[wrap ? 'whitespace-pre-wrap wrap-break-word' : 'overflow-x-scroll']"
          :language="state.meta.language"
          :code="state.content"
        />
      </CardContent>
    </Card>
  </div>
</template>

<script>
const Modes = Object.freeze({
  IMAGE: Symbol("image"),
  AUDIO: Symbol("audio"),
  VIDEO: Symbol("video"),
  PDF: Symbol("pdf"),
  MARKDOWN: Symbol("markdown"),
  CSV: Symbol("csv"),
  ARCHIVE: Symbol("archive"),
  TEXT: Symbol("text"),
});
</script>

<script setup>
import { UseTimeAgo } from "@vueuse/components";
import { useAsyncState, useTimeoutFn } from "@vueuse/core";
import axios from "axios";
import { computed, ref } from "vue";
import { toast } from "vue-sonner";
import DeadLink from "@/assets/dead-link.svg";
import HighlightJS from "@/components/HighlightJS.js";
import DownloadButton from "@/components/display/DownloadButton.vue";
import EditButton from "@/components/display/EditButton.vue";
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
import { Switch } from "@/components/ui/switch/index.js";
import { Table, TableCell, TableRow } from "@/components/ui/table/index.js";
import { ApiPath } from "@/config/api.js";
import { useConfigStore } from "@/stores/config.js";
import { getExtension, loadLanguage } from "@/util/extensions.js";
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
const csvRows = ref(250);

const expiryMs = ref();
const expired = ref(false);
const expiryTimeout = useTimeoutFn(() => (expired.value = true), expiryMs, { immediate: false });

const { state, isLoading, error, execute } = useAsyncState(async () => {
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
      toast.error("Failed to load file", {
        description: err.message,
      });
    }
    throw err;
  }

  const meta = res.data;

  if (meta.original_name) {
    document.title = meta.original_name + " · " + config.site.site_name;
  }

  let mode;
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

  let content, formatted;
  if (
    meta.size < 512 * 1024 &&
    (mode === Modes.TEXT || mode === Modes.MARKDOWN || mode === Modes.CSV)
  ) {
    try {
      const res = await Promise.all([
        axios.get(meta.direct_url, {
          headers: { "Linx-Access-Key": encAccessKey.value },
          responseType: "text",
          validateStatus: (s) => s === 200,
          withCredentials: true,
        }),
        loadLanguage(meta.language),
      ]);
      content = res[0].data;
    } catch (err) {
      console.error(err);
      toast.error("Failed to download file", {
        description: err.message,
      });
      throw err;
    }

    if (mode === Modes.MARKDOWN) {
      try {
        const markdown = (await import("@/util/markdown.js")).default;
        formatted = markdown(content);
      } catch (err) {
        console.error(err);
        toast.error("Failed to format markdown", {
          description: err.message,
        });
      }
    } else if (mode === Modes.CSV) {
      try {
        const papaparse = (await import("papaparse")).default;
        formatted = papaparse.parse(content);
      } catch (err) {
        console.error(err);
        toast.error("Failed to format CSV", {
          description: err.message,
        });
      }
    }
  }

  expiryTimeout.stop();
  expired.value = false;
  if (meta.expiry > 0) {
    expiryMs.value = new Date(meta.expiry * 1000).getTime() - Date.now();
    // https://developer.mozilla.org/docs/Web/API/Window/setTimeout#maximum_delay_value
    if (expiryMs.value < 2 ** 31) {
      expiryTimeout.start();
    }
  }

  return {
    meta,
    mode,
    content,
    formatted,
  };
}, {});

const showWrapSwitch = computed(() => !!state.value.content && state.value.mode === Modes.TEXT);
const showEditButton = computed(
  () =>
    !!state.value.content &&
    (state.value.mode === Modes.TEXT ||
      state.value.mode === Modes.MARKDOWN ||
      state.value.mode === Modes.CSV),
);
const expiry = computed(() =>
  state.value.meta?.expiry > 0 ? new Date(state.value?.meta?.expiry * 1000) : false,
);
</script>
