<template>
  <div v-if="formatted" class="prose">
    <div class="table-wrap">
      <table>
        <thead v-if="header">
          <tr>
            <th v-for="(cell, ckey) in header" :key="ckey">{{ cell }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="(row, key) in rows" :key="key">
            <td v-for="(cell, ckey) in row" :key="ckey">{{ cell }}</td>
          </tr>
        </tbody>
      </table>
    </div>
    <div class="flex justify-between">
      Showing {{ rows.length }} of {{ dataRows.length }} rows
      <Button v-if="csvRows < dataRows.length" @click="csvRows += 250"> Show more </Button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { parse } from "papaparse";
import { computed, ref } from "vue";
import { Button } from "@/components/ui/button";

const props = defineProps({
  content: { type: String, required: true },
});

const formatted = computed(() => parse<string[]>(props.content));
const csvRows = ref(250);

// First record is treated as a header, the way GitHub renders CSVs.
const header = computed(() => formatted.value.data[0]);
const dataRows = computed(() => formatted.value.data.slice(1));
const rows = computed(() => dataRows.value.slice(0, csvRows.value));
</script>
