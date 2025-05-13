import { ApiPath } from "@/config/api.js";
import { useConfigStore } from "@/stores/config.js";
import { useWakeLock } from "@vueuse/core";
import axios from "axios";
import { defineStore } from "pinia";
import { reactive, ref } from "vue";
import { toast } from "vue-sonner";

const config = useConfigStore();

let uploadID = 0;

export const useUploadStore = defineStore(
  "uploads",
  () => {
    const uploads = ref([]);
    const inProgress = ref({});
    let timeout;

    const copy = async (item) => {
      try {
        const url = document.location.origin + "/" + item.filename;
        await navigator.clipboard.writeText(url);
        toast.success("Copied to clipboard.", {
          description: url,
        });
      } catch (err) {
        console.error(err);
        toast.error("Failed to copy.", {
          description: err,
        });
      }
    };

    const wakelock = reactive(useWakeLock());

    const uploadFile = async ({
      file,
      expiry,
      randomFilename = false,
      password,
      saveOriginalName = true,
    }) => {
      const controller = new AbortController();
      const upload = ref({ original_name: file.name, progress: { progress: 0 }, controller });
      const id = uploadID++;
      if (Object.keys(inProgress.value).length === 0) {
        wakelock.request("screen");
      }
      inProgress.value[id] = upload;

      const form = new FormData();
      form.append("file", file);
      form.append("expires", expiry);
      if (password) {
        form.append("access_key", password);
      }
      form.append("randomize", randomFilename);

      try {
        const res = await axios.post(ApiPath(`/upload`), form, {
          headers: {
            Accept: "application/json",
            "Linx-Api-Key": config.apiKey,
          },
          signal: controller.signal,
          validateStatus: (s) => s === 200,
          onUploadProgress(state) {
            upload.value.progress = state;
          },
        });

        if (saveOriginalName) {
          res.data.original_name = file.name;
        }
        res.data.uploaded = new Date();

        toast.success("File uploaded", {
          description: res.data.original_name || res.data.filename,
          action: {
            label: "Copy",
            onClick: async () => await copy(res.data),
          },
        });
        uploads.value.unshift(res.data);
        removeExpired();
        return res.data;
      } catch (err) {
        let description = err.response?.data?.error || err.message;
        if (description === "canceled") {
          description = "Canceled by user";
        }
        toast.error("Upload failed", { description });
        throw err;
      } finally {
        delete inProgress.value[id];
        if (Object.keys(inProgress.value).length === 0) {
          wakelock.release();
        }
      }
    };

    const deleteItem = async (upload) => {
      try {
        await axios.delete(ApiPath(`/${upload.filename}`), {
          validateStatus: (s) => s === 200 || s === 404,
          headers: {
            Accept: "application/json",
            "Linx-Api-Key": config.apiKey,
            "Linx-Delete-Key": upload.delete_key,
          },
        });
        uploads.value = uploads.value.filter((u) => u.filename !== upload.filename);
        toast.success("File deleted", { description: upload.filename });
      } catch (err) {
        toast.error("Delete failed", { description: err.response?.data?.error || err.message });
        throw err;
      }
    };

    const removeExpired = () => {
      const now = Math.floor(Date.now() / 1000);
      uploads.value = uploads.value.filter(
        (upload) => upload.expiry === "0" || upload.expiry > now,
      );
      const closest = Math.min(
        ...uploads.value.map((upload) => upload.expiry).filter((e) => e > 0),
      );
      const nextRun = (closest - now) * 1000;
      if (!Number.isFinite(nextRun)) return;
      clearTimeout(timeout);
      timeout = setTimeout(removeExpired, nextRun);
    };

    return { uploads, inProgress, uploadFile, deleteItem, removeExpired, copy };
  },
  {
    persist: {
      pick: ["uploads"],
      afterHydrate(ctx) {
        ctx.store.removeExpired();
      },
    },
  },
);
