<script setup>
import { Input } from "@/components/ui/input/index.js";

const model = defineModel();
</script>

<template>
  <Input type="password" v-model="model" placeholder="Password" aria-label="Password" />
</template>
