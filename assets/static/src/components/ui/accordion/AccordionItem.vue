<script setup>
import { cn } from "@/lib/utils";
import { AccordionItem, useForwardProps } from "reka-ui";
import { computed } from "vue";

const props = defineProps({
  disabled: { type: Boolean, required: false },
  value: { type: String, required: true },
  unmountOnHide: { type: Boolean, required: false },
  asChild: { type: Boolean, required: false },
  as: { type: null, required: false },
  class: { type: null, required: false },
});

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props;

  return delegated;
});

const forwardedProps = useForwardProps(delegatedProps);
</script>

<template>
  <AccordionItem
    data-slot="accordion-item"
    v-bind="forwardedProps"
    :class="cn('border-b last:border-b-0', props.class)"
  >
    <slot />
  </AccordionItem>
</template>
