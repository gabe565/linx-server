<template>
  <Dialog>
    <DialogTrigger as-child>
      <Button v-bind="$attrs" variant="secondary" size="icon" class="rounded-r-none">
        <span class="sr-only">Info</span>
        <InfoIcon />
      </Button>
    </DialogTrigger>
    <DialogContent>
      <DialogHeader class="min-w-0">
        <DialogTitle class="wrap-break-word pr-4">{{ item.filename }}</DialogTitle>
        <DialogDescription v-if="item.expiry > 0">
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

      <Table class="[&>tr]:flex [&>tr]:sm:table-row [&>tr]:flex-col [&_td:first-child]:font-semibold">
        <TableRow v-if="item.original_name">
          <TableCell>Original Name</TableCell>
          <TableCell>{{ item.original_name }}</TableCell>
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
  </Dialog>
</template>

<script setup>
import { Button } from "@/components/ui/button/index.js";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog/index.js";
import { Table, TableCell, TableRow } from "@/components/ui/table/index.js";
import PasswordViewInput from "@/components/upload/PasswordViewInput.vue";
import { formatBytes } from "@/util/bytes.js";
import { UseTimeAgo } from "@vueuse/components";
import InfoIcon from "~icons/material-symbols/info-rounded";

defineProps({
  item: { type: Object, required: true },
});
</script>
