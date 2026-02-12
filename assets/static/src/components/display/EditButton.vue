<template>
  <Button variant="outline" @click="edit">
    <EditIcon class="text-2xl" />
    Edit
  </Button>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";
import { Button } from "@/components/ui/button/index.js";
import { useConfigStore } from "@/stores/config.ts";
import { getExtension } from "@/util/extensions.ts";
import EditIcon from "~icons/material-symbols/edit-rounded";

const props = defineProps({
  meta: { type: Object, required: true },
  content: { type: String, required: true },
});

const config = useConfigStore();
const router = useRouter();

const edit = () => {
  config.extension = getExtension(props.meta.filename);
  config.content = props.content;
  router.push({ path: "/paste" });
};
</script>
