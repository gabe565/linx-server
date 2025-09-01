<script setup lang="ts">
import hljs from "highlight.js/lib/core";
import langJSON from "highlight.js/lib/languages/json";
import { computed } from "vue";
import HighlightJS from "@/components/HighlightJS.ts";
import { ApiPath } from "@/config/api.ts";
import { AlphaNum, Hex, randomString } from "@/util/random.ts";

hljs.registerLanguage("json", langJSON);

const props = defineProps({
  filename: { type: String, default: randomString(8, AlphaNum) },
  extension: { type: String, default: "jpg" },
  mime: { type: String, default: "image/jpeg" },
  size: { type: Number, default: 1048576 },
  full: { type: Boolean, default: false },
});

const file = computed(() => props.filename + "." + props.extension);
const deleteKey = randomString(30, AlphaNum);
const expiry = Math.floor(Date.now() / 1000) + 60 * 60;
const sha = randomString(64, Hex);

let code = "{";
if (props.full) {
  code += `
  // Public file page
  "url": "${ApiPath(`/${file.value}`)}",`;
}
code += `
  // URL to access the file directly
  "direct_url": "${ApiPath(`/selif/${file.value}`)}",
  // Optionally-generated filename
  "filename": "${file.value}",`;
if (props.full) {
  code += `
  // Optionally-generated deletion key
  "delete_key": "${deleteKey}",
  // Optionally-supplied access key
  "access_key": "",`;
}
code += `
  // Unix timestamp at which the file will expire (0 if no expiry)
  "expiry": ${expiry},
  // Size in bytes of the file
  "size": ${props.size},
  // Inferred mimetype of the file
  "mimetype": "${props.mime}",
  // SHA256 checksum of the file
  "sha256sum": "${sha}"
}`;
</script>

<template>
  <HighlightJS lang="json" :code="code" class="overflow-x-auto p-3 rounded" />
</template>
