import { ApiPath } from "@/config/api.js";
import axios from "axios";
import { defineStore } from "pinia";
import { ref } from "vue";
import { toast } from "vue-sonner";

export const useConfigStore = defineStore(
  "config",
  () => {
    const site = ref({ site_name: "Linx", expiration_times: [] });
    const apiKey = ref("");
    const expiry = ref("");
    const filename = ref("");
    const extension = ref("txt");
    const randomFilename = ref(true);
    const password = ref("");
    const content = ref("");

    const loadConfig = async () => {
      try {
        const res = await axios.get(ApiPath("/api/config"), {
          validateStatus: (s) => s === 200,
        });
        site.value = res.data;
        const times = site.value.expiration_times;
        expiry.value = times[times.length - 1].value;
      } catch (err) {
        toast.error("Failed to load config", { description: err.message });
        throw err;
      }
    };
    loadConfig();

    return { site, apiKey, expiry, filename, extension, randomFilename, password, content };
  },
  {
    persist: {
      pick: ["site", "apiKey"],
    },
  },
);
