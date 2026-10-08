<template>
  <div v-if="formatted" class="prose" v-html="formatted" />
</template>

<script setup lang="ts">
import { computedAsync } from "@vueuse/core";
import DOMPurify from "dompurify";
import { marked } from "@/util/markdown.ts";

const props = defineProps({
  content: { type: String, required: true },
});

// Async so code blocks can lazy-load their highlight.js languages.
const formatted = computedAsync(async () => {
  const parsed = await marked.parse(props.content);
  const root = DOMPurify.sanitize(parsed, { RETURN_DOM: true }) as HTMLElement;
  // Wrapped so a wide table scrolls instead of squeezing columns to min-content.
  for (const table of root.querySelectorAll("table")) {
    const wrap = document.createElement("div");
    wrap.className = "table-wrap";
    table.replaceWith(wrap);
    wrap.append(table);
  }
  return root.innerHTML;
}, "");
</script>
