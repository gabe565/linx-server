<template>
  <div class="container flex flex-col justify-center gap-6 max-w-2xl mx-auto" v-bind="$attrs">
    <Card>
      <CardHeader>
        <CardTitle>Upload</CardTitle>
      </CardHeader>

      <CardContent class="flex flex-col gap-4">
        <div class="flex flex-col sm:flex-row items-center justify-between gap-4">
          <Label v-if="!config.site?.force_random">
            <Switch v-model="config.randomFilename" />
            Random filename
          </Label>
          <PasswordInput v-model="config.password" class="sm:flex-1" />
          <ExpirySelect
            v-model="config.expiry"
            :options="config.site?.expiration_times"
            class="w-full sm:w-40"
          />
        </div>
        <DropZone @upload="doUpload" :max-file-size="config.site?.max_size" />
      </CardContent>
    </Card>

    <UploadList v-model:show-auth="showAuth" />
  </div>

  <AuthDialog v-if="config.site?.auth" v-model="showAuth" @submit="doUpload(retryFile)" />
</template>

<script setup>
import { ref } from "vue";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card/index.js";
import { Label } from "@/components/ui/label/index.js";
import { Switch } from "@/components/ui/switch/index.js";
import AuthDialog from "@/components/upload/AuthDialog.vue";
import DropZone from "@/components/upload/DropZone.vue";
import ExpirySelect from "@/components/upload/ExpirySelect.vue";
import PasswordInput from "@/components/upload/PasswordInput.vue";
import UploadList from "@/components/upload/UploadList.vue";
import { useConfigStore } from "@/stores/config";
import { useUploadStore } from "@/stores/upload";

const config = useConfigStore();
const uploads = useUploadStore();
const showAuth = ref(false);
let retryFile;

const doUpload = async (file) => {
  if (!file) return;
  retryFile = null;
  try {
    await uploads.uploadFile({
      file,
      randomFilename: config.randomFilename,
      expiry: config.expiry,
      password: config.password,
    });
  } catch (err) {
    console.error(err);
    if (err.response?.status === 401) {
      retryFile = file;
      showAuth.value = true;
    }
  }
};
</script>
