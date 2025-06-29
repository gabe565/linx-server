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
            />
            <span class="p-1 text-gray-500">.</span>
            <Input
              v-model="config.extension"
              placeholder="Ext"
              class="w-1/4 sm:min-w-16"
              aria-label="Extension"
              @focus="$event.target.select()"
            />
          </div>

          <div class="flex flex-col sm:flex-row gap-4 justify-between">
            <PasswordInput v-model="config.password" class="w-full sm:w-50 ml-auto" />
            <ExpirySelect
              v-model="config.expiry"
              :options="config.site?.expiration_times"
              class="w-full sm:w-40"
            />
            <Button type="submit">Paste</Button>
          </div>
        </div>

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

<script setup>
import { useDropZone, useEventListener, useMagicKeys } from "@vueuse/core";
import { onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { CardTitle } from "@/components/ui/card/index.js";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import AuthDialog from "@/components/upload/AuthDialog.vue";
import ExpirySelect from "@/components/upload/ExpirySelect.vue";
import PasswordInput from "@/components/upload/PasswordInput.vue";
import { useConfigStore } from "@/stores/config.js";
import { useUploadStore } from "@/stores/upload.js";

const config = useConfigStore();
const upload = useUploadStore();
const router = useRouter();
const showAuth = ref(false);

const doUpload = async () => {
  const file = new File([config.content], config.filename + "." + config.extension);
  try {
    const res = await upload.uploadFile({
      file: file,
      expiry: config.expiry,
      password: config.password,
      saveOriginalName: false,
    });
    config.content = "";
    config.extension = "txt";
    await router.push(`/${res.filename}`);
  } catch (err) {
    console.error(err);
    if (err.response?.status === 401) {
      showAuth.value = true;
    }
  }
};

const { Ctrl_Enter, Meta_Enter } = useMagicKeys();

watch(Ctrl_Enter, doUpload);
watch(Meta_Enter, doUpload);

const textarea = ref();
onMounted(() => textarea.value.$el.focus());

const loadFile = async (file) => {
  if (file.size > 1024 * 1024) return;
  config.filename = file.name?.split(".").slice(0, -1).join(".") || "";
  config.extension = file.name?.split(".").pop() || "txt";
  config.content = await file.text();
};

useDropZone(window, {
  dataTypes(t) {
    t = t[0];
    return t.startsWith("text/") || t === "application/json" || t.endsWith("yaml");
  },
  async onDrop(files) {
    if (!files?.length) return;
    await loadFile(files[0]);
  },
  preventDefaultForUnhandled: true,
});

useEventListener(window, "paste", async (e) => {
  if (!e.clipboardData?.files?.length) return;
  await loadFile(e.clipboardData.files[0]);
});
</script>
