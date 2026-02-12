import { defineStore } from "pinia";
import { ref } from "vue";

export interface ExpirationTime {
  name: string;
  value: string;
}

export interface WindowConfig {
  site_name: string;
  site_path: string;
  max_size: number;
  force_random: boolean;
  auth: boolean;
  expiration_times: ExpirationTime[];
  custom_pages?: string[];
}

declare global {
  interface Window {
    config: WindowConfig;
  }
}

export const useConfigStore = defineStore(
  "config",
  () => {
    const site = ref(window.config);
    const apiKey = ref("");
    const expiry = ref(site.value?.expiration_times?.at(-1)?.value ?? "");
    const filename = ref("");
    const extension = ref("txt");
    const randomFilename = ref(true);
    const password = ref("");
    const content = ref("");

    return { site, apiKey, expiry, filename, extension, randomFilename, password, content };
  },
  {
    persist: {
      pick: ["apiKey", "expiry"],
      afterHydrate(ctx) {
        const options = (ctx.store.site?.expiration_times ?? []) as ExpirationTime[];
        const fallback = options.at(-1)?.value ?? "";
        if (!options.some((opt) => opt.value === ctx.store.expiry)) {
          ctx.store.expiry = fallback;
        }
      },
    },
  },
);
