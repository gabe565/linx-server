<template>
  <CardHeader class="gap-4 sm:flex sm:flex-wrap sm:items-center sm:justify-between">
    <div class="flex flex-wrap items-start gap-4 w-full sm:w-auto sm:flex-1 sm:min-w-0">
      <div class="flex flex-col gap-1 min-w-56 max-w-full flex-1">
        <CardTitle class="wrap-break-word">
          {{ state.meta.original_name || state.meta.filename }}
        </CardTitle>

        <UseTimeAgo
          v-if="expiry"
          v-slot="{ timeAgo }"
          :time="expiry"
          :show-second="true"
          :update-interval="1000"
        >
          <CardDescription class="text-xs tabular-nums">
            {{ expired ? "expired" : "expires" }} {{ timeAgo }}
          </CardDescription>
        </UseTimeAgo>
      </div>

      <ButtonGroup class="shrink-0 max-w-full ml-auto" v-if="isPlainText">
        <EditButton :meta="state.meta" :content="state.content" />
        <CopyButton :content="state.content" />
      </ButtonGroup>
    </div>

    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-end w-full sm:w-auto">
      <div v-if="showWrapSwitch" class="flex items-center space-x-2">
        <Switch id="opt-wrap" v-model="wrap" />
        <Label for="opt-wrap">Wrap</Label>
      </div>

      <DownloadButton
        v-if="state.meta"
        :meta="state.meta"
        :disabled="expired"
        class="w-full sm:w-auto"
      />
    </div>
  </CardHeader>
</template>

<script setup lang="ts">
import Modes from "./fileModes.js";
import { UseTimeAgo } from "@vueuse/components";
import { useTimeoutFn } from "@vueuse/core";
import { computed, ref, watch } from "vue";
import CopyButton from "@/components/display/CopyButton.vue";
import DownloadButton from "@/components/display/DownloadButton.vue";
import EditButton from "@/components/display/EditButton.vue";
import { ButtonGroup } from "@/components/ui/button-group";
import { CardDescription, CardHeader, CardTitle } from "@/components/ui/card/index.js";
import { Label } from "@/components/ui/label/index.js";
import { Switch } from "@/components/ui/switch/index.js";

const props = defineProps({
  state: { type: Object, required: true },
});

const wrap = defineModel<boolean>("wrap");

const expiry = computed(() => {
  const exp = props.state?.meta?.expiry;
  return exp && exp > 0 ? new Date(exp * 1000) : false;
});
const expiryMs = ref();
const expired = ref(false);
const expiryTimeout = useTimeoutFn(() => (expired.value = true), expiryMs, { immediate: false });

watch(
  () => props.state,
  () => {
    expired.value = false;
    if (props.state?.meta?.expiry > 0) {
      expiryMs.value = new Date(props.state?.meta?.expiry * 1000).getTime() - Date.now();
      // https://developer.mozilla.org/docs/Web/API/Window/setTimeout#maximum_delay_value
      if (expiryMs.value < 2 ** 31) {
        expiryTimeout.start();
      }
    }
  },
  { immediate: true },
);

const showWrapSwitch = computed(() => !!props.state?.content && props.state?.mode === Modes.TEXT);
const isPlainText = computed(
  () =>
    !!props.state?.content &&
    (props.state?.mode === Modes.TEXT ||
      props.state?.mode === Modes.MARKDOWN ||
      props.state?.mode === Modes.CSV),
);
</script>
