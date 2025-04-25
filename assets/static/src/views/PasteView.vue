<template>
  <form @submit.prevent="doUpload" v-bind="$attrs">
    <Card class="container max-w-4xl mx-auto">
      <CardHeader>
        <CardTitle>Paste</CardTitle>
      </CardHeader>

      <CardContent class="space-y-6">
        <div
          class="flex flex-col sm:flex-row flex-wrap items-center justify-center sm:justify-between gap-4 w-full"
        >
          <div class="flex items-end grow">
            <Input
              v-model="config.filename"
              placeholder="Filename"
              class="w-40"
              aria-label="Filename"
            />
            <span class="p-1 text-gray-500">.</span>
            <Input
              v-model="config.extension"
              placeholder="Ext"
              class="w-16"
              aria-label="Extension"
              @focus="$event.target.select()"
            />
          </div>

          <PasswordInput v-model="config.password" />
          <ExpirySelect v-model="config.expiry" :options="config.site?.expiration_times" />
          <Button type="submit">Paste</Button>
        </div>

        <Textarea v-model="config.content" placeholder="Paste your text here..." class="h-96" />
      </CardContent>
    </Card>
  </form>

  <AuthDialog v-if="config.site?.auth" v-model="showAuth" @submit="doUpload" />
</template>

<script setup>
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
import { useMagicKeys } from "@vueuse/core";
import { ref, watch } from "vue";
import { useRouter } from "vue-router";

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
    if (err.response?.status === 401) {
      showAuth.value = true;
      return;
    }
    throw err;
  }
};

const { Ctrl_Enter, Meta_Enter } = useMagicKeys();

watch(Ctrl_Enter, doUpload);
watch(Meta_Enter, doUpload);
</script>
