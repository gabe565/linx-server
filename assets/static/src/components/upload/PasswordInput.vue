<script setup>
import { Input } from "@/components/ui/input/index.js";

const model = defineModel();
</script>

<template>
  <div class="m-auto">
    <Input
      type="password"
      v-model="model"
      placeholder="File Password"
      class="flex-1 min-w-50"
      aria-label="Password"
    />
  </div>
</template>
