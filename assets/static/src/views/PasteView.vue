<template>
  <form @submit.prevent="doUpload">
    <Card class="container max-w-4xl mx-auto" v-bind="$attrs">
      <CardHeader>
        <CardTitle>Paste</CardTitle>
      </CardHeader>

      <CardContent class="space-y-6">
        <div class="flex flex-wrap flex-col sm:flex-row gap-4 w-full justify-between">
          <div class="flex items-end sm:w-60">
            <Input
              v-model="config.filename"
              placeholder="Filename"
              class="w-3/4 sm:min-w-30"
              aria-label="Filename"
              :disabled="config.overwrite && canOverwriteExisting"
            />
            <span class="p-1 text-gray-500">.</span>
            <Input
              v-model="config.extension"
              placeholder="Ext"
              class="w-1/4 sm:min-w-16"
              aria-label="Extension"
              :disabled="config.overwrite && canOverwriteExisting"
              @focus="$event.target.select()"
            />
          </div>

          <div class="flex flex-col sm:flex-row gap-4 justify-between items-start sm:items-center">
            <Tooltip v-if="canOverwriteExisting">
              <TooltipTrigger as-child>
                <Toggle
                  variant="outline"
                  :model-value="config.overwrite"
                  @update:model-value="(v) => (config.overwrite = !!v)"
                  :data-state="config.overwrite ? 'on' : 'off'"
                >
                  <PublishedChangesIcon class="text-2xl" />
                  <span class="sr-only">Overwrite existing link</span>
                </Toggle>
              </TooltipTrigger>
              <TooltipContent side="bottom">
                <div class="text-sm">Overwrite existing link</div>
                <div class="text-xs text-muted-foreground">
                  Available because you uploaded this file.
                </div>
              </TooltipContent>
            </Tooltip>

            <PasswordInput v-model="config.password" class="w-full sm:w-50 ml-auto" />
            <ExpirySelect
              v-model="config.expiry"
              :options="config.site?.expiration_times"
              class="w-full sm:w-40"
            />
            <Button type="submit">Paste</Button>
          </div>
        </div>

        <Alert v-if="config.editTargetFilename && !canOverwriteExisting">
          <InfoIcon />
          <AlertTitle>
            This file is not in your upload history, so editing will create a new link.
          </AlertTitle>
        </Alert>

        <Textarea
          ref="textarea"
          v-model="config.content"
          placeholder="Paste your text here..."
          class="font-mono h-96"
          autofocus
          autocomplete="off"
          autocorrect="off"
          autocapitalize="off"
          spellcheck="false"
        />
      </CardContent>
    </Card>
  </form>

  <AuthDialog v-if="config.site?.auth" v-model="showAuth" @submit="doUpload" />
</template>

<script setup lang="ts">
import { useDropZone, useEventListener, useMagicKeys } from "@vueuse/core";
import { isAxiosError } from "axios";
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { toast } from "vue-sonner";
import { Alert, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { CardTitle } from "@/components/ui/card/index.js";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Toggle } from "@/components/ui/toggle";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import AuthDialog from "@/components/upload/AuthDialog.vue";
import ExpirySelect from "@/components/upload/ExpirySelect.vue";
import PasswordInput from "@/components/upload/PasswordInput.vue";
import { useConfigStore } from "@/stores/config.ts";
import { useUploadStore } from "@/stores/upload.ts";
import InfoIcon from "~icons/material-symbols/info-rounded";
import PublishedChangesIcon from "~icons/material-symbols/published-with-changes-rounded";

const config = useConfigStore();
const upload = useUploadStore();
const router = useRouter();
const showAuth = ref(false);
const canOverwriteExisting = computed(() => !!config.editTargetFilename && !!config.editDeleteKey);

const doUpload = async () => {
  const file = new File([config.content], config.filename + "." + config.extension);
  try {
    const res =
      config.overwrite && canOverwriteExisting.value
        ? await upload.overwriteFile({
            file,
            filename: config.editTargetFilename,
            deleteKey: config.editDeleteKey,
            expiry: config.expiry,
            password: config.password,
            saveOriginalName: false,
          })
        : await upload.uploadFile({
            file,
            expiry: config.expiry,
            password: config.password,
            saveOriginalName: false,
          });

    if (
      config.overwrite &&
      canOverwriteExisting.value &&
      res.filename !== config.editTargetFilename
    ) {
      toast.warning("Could not overwrite original link.", {
        description: `Created ${res.filename} instead.`,
      });
    }

    config.content = "";
    config.filename = "";
    config.extension = "txt";
    config.editTargetFilename = "";
    config.editDeleteKey = "";
    config.overwrite = false;
    await router.push(`/${res.filename}`);
  } catch (err) {
    console.error(err);
    if (isAxiosError(err) && err.response?.status === 401) {
      showAuth.value = true;
    }
  }
};

const { Ctrl_Enter, Meta_Enter } = useMagicKeys();

const ctrlEnter = Ctrl_Enter ?? ref(false);
const metaEnter = Meta_Enter ?? ref(false);

watch(ctrlEnter, (pressed) => pressed && doUpload());
watch(metaEnter, (pressed) => pressed && doUpload());
watch(canOverwriteExisting, (can) => {
  if (!can) config.overwrite = false;
});

const textarea = ref();
onMounted(() => textarea.value.$el.focus());

const loadFile = async (file: File) => {
  if (file.size > 1024 * 1024) return;
  config.filename = file.name?.split(".").slice(0, -1).join(".") || "";
  config.extension = file.name?.split(".").pop() || "txt";
  config.content = await file.text();
};

useDropZone(document, {
  dataTypes(t) {
    const type = t[0];
    if (!type) return false;
    return type.startsWith("text/") || type === "application/json" || type.endsWith("yaml");
  },
  async onDrop(files) {
    if (!files?.length) return;
    await loadFile(files[0] as File);
  },
  preventDefaultForUnhandled: true,
});

useEventListener(window, "paste", async (e: ClipboardEvent) => {
  if (!e.clipboardData?.files?.length) return;
  await loadFile(e.clipboardData.files[0] as File);
});
</script>
