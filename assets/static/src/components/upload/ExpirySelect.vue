<template>
  <Select>
    <SelectTrigger v-bind="$attrs">
      <SelectValue placeholder="Select expiration" />
    </SelectTrigger>
    <SelectContent>
      <SelectItem v-for="opt in options" :key="opt.value" :value="opt.value">
        {{ opt.name }}
      </SelectItem>
    </SelectContent>
  </Select>
</template>

<script setup lang="ts">
import type { PropType } from "vue";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select/index.js";

defineProps({
  options: {
    type: Array as PropType<Array<{ value: string; name: string }>>,
    required: true,
  },
});
</script>
