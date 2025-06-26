<template>
  <div v-if="isLoading" class="animate-in fade-in duration-1000 flex flex-col items-center">
    <SpinnerIcon class="text-4xl" />
  </div>

  <Card v-else class="container max-w-3xl mx-auto">
    <CardContent>
      <div v-html="state" class="prose max-w-none" />
    </CardContent>
  </Card>
</template>

<script setup>
import { useAsyncState } from "@vueuse/core";
import axios from "axios";
import { toast } from "vue-sonner";
import { Card, CardContent } from "@/components/ui/card/index.js";
import { ApiPath } from "@/config/api.js";
import SpinnerIcon from "~icons/svg-spinners/ring-resize";

const props = defineProps({
  filename: { type: String, required: true },
});

const { state, isLoading } = useAsyncState(async () => {
  try {
    const res = await axios.get(ApiPath(`/api/custom_page/${props.filename}`), {
      validateStatus: (s) => s === 200,
    });

    const markdown = (await import("@/util/markdown.js")).default;
    const content = markdown(res.data);
    return content;
  } catch (err) {
    console.error(err);
    toast.error("Failed to load page", {
      description: err.message,
    });
  }
});
</script>
