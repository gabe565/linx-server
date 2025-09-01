<template>
  <div class="flex gap-2">
    <Input
      @focus="$event.target.select()"
      @click="$event.target.select()"
      :type="show ? 'text' : 'password'"
      v-model="model"
      readonly
      class="cursor-pointer h-8"
    />
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger as-child>
          <Button variant="secondary" size="sm" @click="show = !show">
            <VisibilityOffIcon v-if="show" />
            <VisibilityIcon v-else />
            <span class="sr-only">{{ text }}</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent side="left">{{ text }}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import { Button } from "@/components/ui/button/index.js";
import { Input } from "@/components/ui/input/index.js";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip/index.js";
import VisibilityOffIcon from "~icons/material-symbols/visibility-off-rounded";
import VisibilityIcon from "~icons/material-symbols/visibility-rounded";

const model = defineModel<string>();

const show = ref(false);
const text = computed(() => (show.value ? "Hide" : "Show"));
</script>
