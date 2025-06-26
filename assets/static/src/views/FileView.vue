<template>
  <DisplayPage v-if="!isCustom" :filename="filename" />
  <CustomPage v-else :filename="filename" />
</template>

<script setup>
import { computed } from "vue";
import CustomPage from "@/components/display/CustomPage.vue";
import DisplayPage from "@/components/display/DisplayPage.vue";
import { useConfigStore } from "@/stores/config.js";

const props = defineProps({
  filename: { type: String, required: true },
});

document.title = props.filename + " · " + useConfigStore().site.site_name;

const config = useConfigStore();
const isCustom = computed(() => config.site?.custom_pages?.includes(props.filename));
</script>
