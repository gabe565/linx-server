<template>
  <div
    class="w-full h-50 border-2 border-dashed border-gray-500 text-gray-500 dark:border-gray-400 dark:text-gray-400 opacity-75 hover:opacity-100 rounded-lg flex flex-col items-center justify-center text-center cursor-pointer transition"
    :class="{ 'opacity-100': isOverDropZone }"
    @click.prevent="triggerFileInput"
    v-bind="$attrs"
  >
    <div class="flex flex-col place-content-center items-center flex-1 gap-4">
      <UploadIcon class="text-4xl" />
      <Label for="dropzone" class="flex-wrap justify-center gap-1 cursor-pointer">
        <span class="font-bold">Click to upload</span> or drag and drop
      </Label>
    </div>
    <span v-if="maxFileSize" class="text-sm opacity-60 pb-1">
      Upload up to {{ formatBytes(maxFileSize) }}
    </span>
  </div>

  <input
    id="dropzone"
    ref="fileInput"
    type="file"
    multiple
    @change="onFileChange"
    class="sr-only"
  />
</template>

<script setup lang="ts">
import { useDropZone, useEventListener } from "@vueuse/core";
import { useTemplateRef } from "vue";
import { Label } from "@/components/ui/label/index.js";
import { formatBytes } from "@/util/bytes.ts";
import UploadIcon from "~icons/material-symbols/upload-rounded";

defineProps({
  maxFileSize: { type: Number, required: false },
});

const fileInput = useTemplateRef("fileInput");
const emit = defineEmits(["upload"]);

const triggerFileInput = () => {
  fileInput.value?.click();
};

useEventListener(window, "paste", (e: ClipboardEvent) => {
  if (!e.clipboardData) return;
  for (const file of e.clipboardData.files) {
    emit("upload", file);
  }
});

const onFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement | null;
  const files = target?.files;
  if (!files) return;
  for (const file of files) {
    emit("upload", file);
  }
  if (fileInput.value) fileInput.value.value = "";
};

const { isOverDropZone } = useDropZone(document, {
  onDrop(files) {
    if (!files) return;
    for (const file of files) {
      emit("upload", file);
    }
  },
  multiple: true,
});
</script>
