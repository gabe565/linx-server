<template>
  <div v-if="formatted">
    <Table>
      <TableRow v-for="(row, key) in formatted?.data?.slice(0, csvRows)" :key="key">
        <TableCell v-for="(cell, ckey) in row" :key="ckey">{{ cell }}</TableCell>
      </TableRow>
    </Table>
    <div class="flex justify-between">
      Showing {{ Math.min(csvRows, formatted?.data?.length) }} of {{ formatted?.data?.length }} rows
      <Button v-if="csvRows < formatted?.data?.length" @click="csvRows += 250"> Show more </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { parse } from "papaparse";
import { computed, ref } from "vue";
import { Button } from "@/components/ui/button";
import { Table, TableCell, TableRow } from "@/components/ui/table";

const props = defineProps({
  content: { type: String, required: true },
});

const formatted = computed(() => parse(props.content));
const csvRows = ref(250);
</script>
