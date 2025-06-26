<template>
  <DialogContent>
    <DialogHeader class="min-w-0">
      <DialogTitle class="wrap-break-word pr-4">{{
        item.original_name || item.filename
      }}</DialogTitle>
      <DialogDescription v-if="item.expiry > 0" class="tabular-nums">
        <UseTimeAgo
          v-slot="{ timeAgo }"
          :time="new Date(item.expiry * 1000)"
          :show-second="true"
          update-interval="1000"
        >
          expires {{ timeAgo }}
        </UseTimeAgo>
      </DialogDescription>
    </DialogHeader>

    <Table
      class="[&>tr]:flex [&>tr]:sm:table-row [&>tr]:flex-col [&_td:first-child]:font-semibold *:last:border-b-0"
    >
      <TableRow>
        <TableCell>Path</TableCell>
        <TableCell>
          <RouterLink :to="`/${item.filename}`" class="link">
            {{ item.filename }}
          </RouterLink>
        </TableCell>
      </TableRow>
      <TableRow v-if="item.access_key">
        <TableCell>Password</TableCell>
        <TableCell>
          <PasswordViewInput :model-value="item.access_key" />
        </TableCell>
      </TableRow>
      <TableRow v-if="item.delete_key">
        <TableCell>Delete Key</TableCell>
        <TableCell>
          <PasswordViewInput :model-value="item.delete_key" />
        </TableCell>
      </TableRow>
      <TableRow v-if="item.uploaded">
        <TableCell>Uploaded</TableCell>
        <TableCell>{{ new Date(item.uploaded).toLocaleString() }}</TableCell>
      </TableRow>
      <TableRow v-if="item.expiry > 0">
        <TableCell>Expires</TableCell>
        <TableCell>{{ new Date(item.expiry * 1000).toLocaleString() }}</TableCell>
      </TableRow>
      <TableRow v-if="item.size">
        <TableCell>Size</TableCell>
        <TableCell>{{ formatBytes(item.size) }}</TableCell>
      </TableRow>
      <TableRow v-if="item.mimetype">
        <TableCell>MIME Type</TableCell>
        <TableCell>{{ item.mimetype }}</TableCell>
      </TableRow>
    </Table>
  </DialogContent>
</template>

<script setup>
import { UseTimeAgo } from "@vueuse/components";
import {
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog/index.js";
import { Table, TableCell, TableRow } from "@/components/ui/table/index.js";
import PasswordViewInput from "@/components/upload/PasswordViewInput.vue";
import { formatBytes } from "@/util/bytes.js";

defineProps({
  item: { type: Object, required: true },
});
</script>
