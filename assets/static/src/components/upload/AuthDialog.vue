<script setup>
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
import { useConfigStore } from "@/stores/config.js";

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
      <DialogHeader>
        <DialogTitle>Authentication Required</DialogTitle>
        <DialogDescription>Please enter an API key.</DialogDescription>
      </DialogHeader>

      <Label>
        API key
        <Input type="password" v-model="config.apiKey" class="flex-1 min-w-50" />
      </Label>

      <DialogFooter @click="submit" class="flex justify-end">
        <Button>Login</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
