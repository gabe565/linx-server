<template>
  <div class="flex items-center">
    <Button
      :as="disabled ? 'button' : 'a'"
      variant="outline"
      size="lg"
      :href="`${meta.direct_url}?download`"
      :download="meta.original_name || meta.filename"
      class="flex-1"
      :class="{ 'rounded-r-none': meta.torrent_url }"
      v-bind="$attrs"
      :disabled="disabled"
    >
      <DownloadIcon class="text-2xl" />
      Download <span class="text-xs text-gray-500">({{ formatBytes(meta.size) }})</span>
    </Button>
    <DropdownMenu v-if="meta.torrent_url">
      <DropdownMenuTrigger as-child class="rounded-l-none border-l-0 !px-2">
        <Button variant="outline" size="lg" :disabled="disabled">
          <DownIcon />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuItem
          as="a"
          :href="meta.torrent_url"
          :download="`${meta.original_name || meta.filename}.torrent`"
          :disabled="disabled"
        >
          Download Torrent
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
</template>

<script setup>
import { Button } from "@/components/ui/button/index.js";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { formatBytes } from "@/util/bytes.js";
import DownloadIcon from "~icons/material-symbols/download-rounded";
import DownIcon from "~icons/material-symbols/keyboard-arrow-down-rounded";

defineProps({
  meta: { type: Object, required: true },
  disabled: { type: Boolean, required: false },
});
</script>
