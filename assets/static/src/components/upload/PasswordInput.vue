<script setup lang="ts">
import type { HTMLAttributes } from "vue";
import { computed, ref } from "vue";
import { Button } from "@/components/ui/button/index.js";
import { Input } from "@/components/ui/input/index.js";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip/index.js";
import { cn } from "@/lib/utils";
import VisibilityOffIcon from "~icons/material-symbols/visibility-off-rounded";
import VisibilityIcon from "~icons/material-symbols/visibility-rounded";

const props = defineProps<{
  class?: HTMLAttributes["class"];
}>();

const model = defineModel<string>();

const show = ref(false);
const text = computed(() => (show.value ? "Hide password" : "Show password"));
</script>

<template>
  <div :class="cn('relative', props.class)">
    <Input
      :type="show ? 'text' : 'password'"
      v-model="model"
      placeholder="Password"
      aria-label="Password"
      class="pr-10"
    />
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger as-child>
          <Button
            variant="ghost"
            size="icon"
            @click="show = !show"
            type="button"
            class="absolute right-0 top-0 h-full px-3 text-muted-foreground hover:text-foreground hover:bg-transparent"
          >
            <VisibilityOffIcon v-if="show" />
            <VisibilityIcon v-else />
            <span class="sr-only">{{ text }}</span>
          </Button>
        </TooltipTrigger>
        <TooltipContent side="bottom">{{ text }}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  </div>
</template>
