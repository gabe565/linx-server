import { ApiPath } from "@/config/api.js";
import { useConfigStore } from "@/stores/config.js";
import axios from "axios";
import { defineStore } from "pinia";
import { ref } from "vue";
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
        toast.error("Failed to copy.", {
          description: err,
        });
        throw err;
      }
    };

    const uploadFile = async ({
      file,
      expiry,
      randomFilename = false,
      password = "",
      saveOriginalName = true,
    }) => {
      const upload = ref({ progress: 0 });
      const id = uploadID++;
      inProgress.value[id] = upload;

      const form = new FormData();
      form.append("file", file);
      form.append("expires", expiry);
      form.append("access_key", password);
      form.append("randomize", randomFilename);

      try {
        const res = await axios.post(ApiPath(`/upload`), form, {
          headers: {
            Accept: "application/json",
            "Linx-Api-Key": config.apiKey,
          },
          validateStatus: (s) => s === 200,
          onUploadProgress({ progress: newVal }) {
            upload.value.progress = newVal * 100;
          },
        });

        if (saveOriginalName) {
          res.data.original_name = file.name;
        }

        toast.success("File uploaded", {
          description: res.data.filename,
          action: {
            label: "Copy",
            onClick: async () => await copy(res.data),
          },
        });
        uploads.value.unshift(res.data);
        if (!timeout) removeExpired();
        return res.data;
      } catch (err) {
        toast.error("Upload failed", { description: err.response?.data?.error || err.message });
        throw err;
      } finally {
        delete inProgress.value[id];
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
