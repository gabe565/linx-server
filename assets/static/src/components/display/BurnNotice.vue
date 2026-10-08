<template>
  <Card>
    <CardHeader>
      <CardTitle class="flex items-center gap-2">
        <BurnIcon class="text-2xl text-destructive shrink-0" />
        <span class="wrap-break-word min-w-0">{{ meta.original_name || meta.filename }}</span>
      </CardTitle>
      <CardDescription>
        This file will be permanently deleted as soon as it is
        {{ preview ? "viewed" : "downloaded" }}. Once you continue, nobody will be able to open this
        link again.
      </CardDescription>
    </CardHeader>
    <CardFooter class="flex flex-col-reverse sm:flex-row justify-end gap-2">
      <Button variant="outline" class="w-full sm:w-auto" @click="copyLink">
        <CopyIcon class="text-2xl" />
        Copy link
      </Button>
      <Button v-if="preview" class="w-full sm:w-auto" :disabled="loading" @click="emit('view')">
        <SpinnerIcon v-if="loading" class="text-2xl" />
        <VisibilityIcon v-else class="text-2xl" />
        View and delete
      </Button>
      <Button
        v-else
        as="a"
        class="w-full sm:w-auto"
        :href="meta.download_url"
        :download="meta.original_name || meta.filename"
        @click="emit('download')"
      >
        <DownloadIcon class="text-2xl" />
        Download and delete
        <span class="text-xs opacity-70">({{ formatBytes(meta.size) }})</span>
      </Button>
    </CardFooter>
  </Card>
</template>

<script setup lang="ts">
import { toast } from "vue-sonner";
import { Button } from "@/components/ui/button/index.js";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card/index.js";
import { formatBytes } from "@/util/bytes.ts";
import CopyIcon from "~icons/material-symbols/content-copy-rounded";
import DownloadIcon from "~icons/material-symbols/download-rounded";
import BurnIcon from "~icons/material-symbols/local-fire-department-rounded";
import VisibilityIcon from "~icons/material-symbols/visibility-rounded";
import SpinnerIcon from "~icons/svg-spinners/ring-resize";

defineProps({
  meta: { type: Object, required: true },
  preview: { type: Boolean, default: false },
  loading: { type: Boolean, default: false },
});

const emit = defineEmits<{ view: []; download: [] }>();

const copyLink = async () => {
  try {
    await navigator.clipboard.writeText(location.href);
    toast.success("Copied to clipboard.", { description: location.href });
  } catch (err) {
    console.error(err);
    toast.error("Failed to copy.", {
      description: err instanceof Error ? err.message : String(err),
    });
  }
};
</script>
