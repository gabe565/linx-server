<script setup lang="ts">
import { Button } from "@/components/ui/button/index.js";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog/index.js";
import { Input } from "@/components/ui/input/index.js";
import { Label } from "@/components/ui/label/index.js";
import { useConfigStore } from "@/stores/config.ts";

const model = defineModel({ type: Boolean });
const config = useConfigStore();

const emit = defineEmits(["submit"]);

const submit = () => {
  model.value = false;
  emit("submit");
};
</script>

<template>
  <Dialog :open="model" @update:open="model = $event">
    <DialogContent>
      <DialogHeader class="pb-2">
        <DialogTitle>Authentication Required</DialogTitle>
        <DialogDescription>This server does not allow public uploads.</DialogDescription>
      </DialogHeader>

      <form @submit.prevent="submit" id="auth" class="space-y-2">
        <Label for="password">API key</Label>
        <Input
          id="password"
          type="password"
          v-model="config.apiKey"
          placeholder="Enter API key"
          class="flex-1 min-w-50"
          autofocus
        />
      </form>

      <DialogFooter class="flex justify-end">
        <Button type="submit" form="auth">Login</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
