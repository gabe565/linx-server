<template>
  <div v-if="formatted" class="prose" v-html="formatted" />
</template>

<script setup>
import DOMPurify from "dompurify";
import { marked } from "marked";
import { computed } from "vue";

const props = defineProps({
  content: { type: String, required: true },
});

const formatted = computed(() => DOMPurify.sanitize(marked.parse(props.content)));
</script>
