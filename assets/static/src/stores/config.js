import { defineStore } from "pinia";
import { ref } from "vue";

export const useConfigStore = defineStore(
  "config",
  () => {
    const site = ref(window.config);
    const apiKey = ref();
    const expiry = ref(site.value.expiration_times[site.value.expiration_times.length - 1].value);
    const filename = ref("");
    const extension = ref("txt");
    const randomFilename = ref(true);
    const password = ref("");
    const content = ref("");

    return { site, apiKey, expiry, filename, extension, randomFilename, password, content };
  },
  {
    persist: {
      pick: ["apiKey"],
    },
  },
);
