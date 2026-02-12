<template>
  <ButtonGroup>
    <Button
      :as="disabled ? 'button' : 'a'"
      variant="outline"
      :href="`${meta.direct_url}?download`"
      :download="meta.original_name || meta.filename"
      class="flex-1"
      v-bind="$attrs"
      :disabled="disabled"
    >
      <DownloadIcon class="text-2xl" />
      Download <span class="text-xs text-gray-500">({{ formatBytes(meta.size) }})</span>
    </Button>
    <DropdownMenu v-if="meta.torrent_url">
      <DropdownMenuTrigger as-child class="!px-2">
        <Button variant="outline" :disabled="disabled">
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
  </ButtonGroup>
</template>

<script setup lang="ts">
import { ButtonGroup } from "@/components/ui/button-group";
import { Button } from "@/components/ui/button/index.js";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { formatBytes } from "@/util/bytes.ts";
import DownloadIcon from "~icons/material-symbols/download-rounded";
import DownIcon from "~icons/material-symbols/keyboard-arrow-down-rounded";

defineProps({
  meta: { type: Object, required: true },
  disabled: { type: Boolean, required: false },
});
</script>
