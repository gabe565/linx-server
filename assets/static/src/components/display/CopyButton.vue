<template>
  <Tooltip>
    <TooltipTrigger as-child>
      <Button variant="outline" @click="copy" v-bind="$attrs">
        <CopyIcon class="text-2xl" />
        <span class="sr-only">Copy raw</span>
      </Button>
    </TooltipTrigger>
    <TooltipContent side="bottom">Copy raw</TooltipContent>
  </Tooltip>
</template>

<script setup lang="ts">
import { toast } from "vue-sonner";
import { Button } from "@/components/ui/button/index.js";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
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
