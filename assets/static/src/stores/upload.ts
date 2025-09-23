import { useEventListener, useWakeLock } from "@vueuse/core";
import axios, { type AxiosProgressEvent, isAxiosError } from "axios";
import { defineStore } from "pinia";
import { reactive, ref } from "vue";
import { toast } from "vue-sonner";
import { ApiPath } from "@/config/api.ts";
import { useConfigStore } from "@/stores/config.ts";

let uploadID = 0;

export type InProgressItem = {
  original_name: string;
  progress: AxiosProgressEvent;
  controller: AbortController;
};

export type UploadedItem = {
  url: string;
  filename: string;
  original_name?: string;
  delete_key: string;
  access_key?: string;
  expiry: number;
  uploaded?: Date;
  size: number;
  mimetype?: string;
};

const MaxDelay = 2 ** 31 - 1;

export const useUploadStore = defineStore(
  "uploads",
  () => {
    const config = useConfigStore();

    const version = ref(0);
    const uploads = ref<UploadedItem[]>([]);
    const inProgress = reactive<Record<number, InProgressItem>>({});
    let timeout: ReturnType<typeof setTimeout> | undefined;

    const copy = async (item: UploadedItem) => {
      try {
        const url = document.location.origin + "/" + item.filename;
        await navigator.clipboard.writeText(url);
        toast.success("Copied to clipboard.", {
          description: url,
        });
      } catch (err) {
        console.error(err);
        toast.error("Failed to copy.", {
          description: err instanceof Error ? err.message : String(err),
        });
        throw err;
      }
    };

    const wakelock = reactive(useWakeLock());

    useEventListener(window, "beforeunload", (e) => {
      if (Object.keys(inProgress).length !== 0) {
        e.preventDefault();
      }
    });

    const uploadFile = async ({
      file,
      expiry,
      randomFilename = false,
      password,
      saveOriginalName = true,
    }: {
      file: File;
      expiry: number | string;
      randomFilename?: boolean;
      password?: string;
      saveOriginalName?: boolean;
    }) => {
      const controller = new AbortController();
      const upload: InProgressItem = {
        original_name: file.name,
        progress: { progress: 0 } as AxiosProgressEvent,
        controller,
      };
      const id = uploadID++;
      if (Object.keys(inProgress).length === 0) {
        wakelock.request("screen");
      }
      inProgress[id] = upload;

      const form = new FormData();
      form.append("size", String(file.size));
      form.append("expires", String(expiry));
      if (password) form.append("access_key", password);
      form.append("randomize", randomFilename.toString());
      // This field must be last since it is streamed
      form.append("file", file);

      try {
        if (config.site.auth) {
          await axios.post(ApiPath("/api/auth"), null, {
            headers: {
              "Linx-Api-Key": encodeURIComponent(config.apiKey),
            },
            validateStatus: (s) => s === 200,
          });
        }

        const res = await axios.post(ApiPath(`/upload`), form, {
          headers: {
            Accept: "application/json",
            "Linx-Api-Key": encodeURIComponent(config.apiKey),
          },
          signal: controller.signal,
          validateStatus: (s) => s === 200,
          onUploadProgress(state) {
            if (inProgress[id]) inProgress[id].progress = state;
          },
        });

        if (saveOriginalName) {
          res.data.original_name = file.name;
        }
        res.data.uploaded = new Date();
        res.data.expiry = Number(res.data.expiry ?? 0);
        res.data.size = Number(res.data.size ?? 0);

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
        let description = err instanceof Error ? err.message : String(err);
        if (isAxiosError(err) && err.response?.data?.error) description = err.response.data.error;
        if (description === "canceled") description = "Canceled by user";
        toast.error("Upload failed", { description });
        throw err;
      } finally {
        delete inProgress[id];
        if (Object.keys(inProgress).length === 0) {
          wakelock.release();
        }
      }
    };

    const deleteItem = async (upload: UploadedItem) => {
      try {
        await axios.delete(ApiPath(`/${upload.filename}`), {
          validateStatus: (s) => s === 200 || s === 404,
          headers: {
            Accept: "application/json",
            "Linx-Api-Key": encodeURIComponent(config.apiKey),
            "Linx-Delete-Key": encodeURIComponent(upload.delete_key ?? ""),
          },
        });
        uploads.value = uploads.value.filter((u) => u.filename !== upload.filename);
        toast.success("File deleted", { description: upload.original_name || upload.filename });
      } catch (err) {
        let description = err instanceof Error ? err.message : String(err);
        if (isAxiosError(err) && err.response?.data?.error) description = err.response.data.error;
        toast.error("Delete failed", { description });
        throw err;
      }
    };

    const removeExpired = () => {
      clearTimeout(timeout);
      const now = Math.floor(Date.now() / 1000);
      let closest = Infinity;
      uploads.value = uploads.value.filter((u) => {
        if (u.expiry === 0) return true;
        if (u.expiry <= now) return false;
        if (u.expiry < closest) closest = u.expiry;
        return true;
      });
      if (!Number.isFinite(closest)) return;
      const nextRun = Math.min(MaxDelay, Math.max(500, (closest - now) * 1000));
      timeout = setTimeout(removeExpired, nextRun);
    };

    return { version, uploads, inProgress, uploadFile, deleteItem, removeExpired, copy };
  },
  {
    persist: {
      pick: ["version", "uploads"],
      afterHydrate(ctx) {
        if (ctx.store.$state.version === 0 && ctx.store.$state.uploads?.length) {
          ctx.store.$state.uploads = ctx.store.$state.uploads.map((u: any) => ({
            ...u,
            expiry: Number(u.expiry),
            size: Number(u.size),
          }));
          ctx.store.version = 1;
        }

        ctx.store.removeExpired();
      },
    },
  },
);
