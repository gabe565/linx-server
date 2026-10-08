<template>
  <Tooltip>
    <TooltipTrigger as-child>
      <Toggle
        variant="outline"
        :model-value="model"
        @update:model-value="(v) => (model = !!v)"
        :data-state="model ? 'on' : 'off'"
        v-bind="$attrs"
      >
        <BurnIcon class="text-2xl" />
        <span class="sr-only">Burn after reading</span>
      </Toggle>
    </TooltipTrigger>
    <TooltipContent side="bottom">
      <div class="text-sm">Burn after reading</div>
      <div class="text-xs text-muted-foreground">Delete after it is viewed once.</div>
    </TooltipContent>
  </Tooltip>
</template>

<script setup lang="ts">
import { Toggle } from "@/components/ui/toggle";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import BurnIcon from "~icons/material-symbols/local-fire-department-rounded";

defineOptions({ inheritAttrs: false });

const model = defineModel<boolean>({ default: false });
</script>
