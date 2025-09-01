<template>
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

    <MarkdownViewer
      v-else-if="!!state.content && state.mode === Modes.MARKDOWN"
      class="max-w-none"
      :content="state.content"
    />

    <pre v-else-if="state.mode === Modes.ARCHIVE" class="overflow-x-scroll max-h-[600px]">{{
      state.meta.archive_files.join("\n")
    }}</pre>

    <CSVViewer
      v-else-if="!!state.content && state.mode === Modes.CSV"
      class="space-y-4"
      :content="state.content"
    />

    <HighlightJS
      v-else-if="!!state.content && state.mode === Modes.TEXT"
      class="p-4"
      :class="[wrap ? 'whitespace-pre-wrap wrap-break-word' : 'overflow-x-scroll']"
      :language="state.meta.language"
      :code="state.content"
    />
  </CardContent>
</template>

<script setup lang="ts">
import { defineAsyncComponent } from "vue";
import HighlightJS from "@/components/HighlightJS.ts";
import Modes from "@/components/display/fileModes.js";
import { CardContent } from "@/components/ui/card/index.js";

const MarkdownViewer = defineAsyncComponent(() => import("@/components/MarkdownViewer.vue"));

const CSVViewer = defineAsyncComponent(() => import("@/components/CSVViewer.vue"));

defineProps({
  state: { type: Object, required: true },
  wrap: { type: Boolean, default: false },
});
</script>
