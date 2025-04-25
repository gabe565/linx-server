<script setup>
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table/index.js";
import { useConfigStore } from "@/stores/config.js";

const config = useConfigStore();
</script>

<template>
  <template v-if="config.site?.auth">
    <h4 class="text-lg font-medium">Required Headers</h4>
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Header</TableHead>
          <TableHead class="w-full">Description</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell><code>Linx-Api-Key: mysecretkey</code></TableCell>
          <TableCell>Auth key</TableCell>
        </TableRow>
      </TableBody>
    </Table>
  </template>

  <h4 class="text-lg font-medium">Optional Headers</h4>
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Header</TableHead>
        <TableHead class="w-full">Description</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      <TableRow>
        <TableCell><code>Linx-Delete-Key: mysecret</code></TableCell>
        <TableCell>Set a custom delete key</TableCell>
      </TableRow>
      <TableRow>
        <TableCell><code>Linx-Access-Key: mysecret</code></TableCell>
        <TableCell>Password-protect file with an access key</TableCell>
      </TableRow>
      <TableRow>
        <TableCell><code>Linx-Expiry: 1h</code></TableCell>
        <TableCell>Set file expiry (Go duration or seconds)</TableCell>
      </TableRow>
      <TableRow>
        <TableCell><code>Accept: application/json</code></TableCell>
        <TableCell>Request JSON response</TableCell>
      </TableRow>
    </TableBody>
  </Table>
</template>
