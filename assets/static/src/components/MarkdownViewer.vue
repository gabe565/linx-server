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
  return DOMPurify.sanitize(parsed);
});
</script>
