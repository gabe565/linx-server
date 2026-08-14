<template>
  <div v-if="formatted" class="prose" v-html="formatted" />
</template>

<script setup lang="ts">
import DOMPurify from "dompurify";
import { marked } from "marked";
import { computed } from "vue";

const props = defineProps({
  content: { type: String, required: true },
});

const formatted = computed(() => {
  const parsed = marked.parse(props.content) as string;
  const root = DOMPurify.sanitize(parsed, { RETURN_DOM: true }) as HTMLElement;
  // Wrapped so a wide table scrolls instead of squeezing columns to min-content.
  for (const table of root.querySelectorAll("table")) {
    const wrap = document.createElement("div");
    wrap.className = "table-wrap";
    table.replaceWith(wrap);
    wrap.append(table);
  }
  return root.innerHTML;
});
</script>
