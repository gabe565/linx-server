<template>
  <Button variant="outline" @click="copy">
    <CopyIcon class="text-2xl" />
    Copy
  </Button>
</template>

<script setup lang="ts">
import { toast } from "vue-sonner";
import { Button } from "@/components/ui/button/index.js";
import CopyIcon from "~icons/material-symbols/content-copy-rounded";

const props = defineProps({
  content: { type: String, required: true },
});

const copy = async () => {
  try {
    await navigator.clipboard.writeText(props.content);
    toast.success("Copied to clipboard.");
  } catch (err) {
    console.error(err);
    toast.error("Failed to copy.", {
      description: err instanceof Error ? err.message : String(err),
    });
  }
};
</script>
