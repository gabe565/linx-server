<script setup>
import { AccordionRoot, useForwardPropsEmits } from "reka-ui";

const props = defineProps({
  collapsible: { type: Boolean, required: false },
  disabled: { type: Boolean, required: false },
  dir: { type: String, required: false },
  orientation: { type: String, required: false },
  unmountOnHide: { type: Boolean, required: false },
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  type: { type: String, required: false },
  modelValue: { type: null, required: false },
  defaultValue: { type: null, required: false },
});
const emits = defineEmits(["update:modelValue"]);

const forwarded = useForwardPropsEmits(props, emits);
</script>

<template>
  <AccordionRoot data-slot="accordion" v-bind="forwarded">
    <slot />
  </AccordionRoot>
</template>
