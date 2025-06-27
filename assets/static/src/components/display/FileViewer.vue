<template>
  <CardContent class="flex flex-col justify-center">
    <div
      v-if="state.mode === Modes.IMAGE"
      class="mx-auto"
      :class="{ 'w-full h-full flex flex-col': state.meta.mimetype === 'image/svg+xml' }"
    >
      <img :src="state.meta.direct_url" alt="" class="max-w-full h-auto rounded-md" />
    </div>

    <div v-else-if="state.mode === Modes.AUDIO">
      <audio controls preload="metadata" class="w-full rounded-md">
        <source :src="state.meta.direct_url" :type="state.meta.mimetype" />
        Your browser doesn’t support playing {{ state.meta.mimetype }}.
      </audio>
    </div>

    <div v-else-if="state.mode === Modes.VIDEO">
      <video controls preload="metadata" class="w-full rounded-md">
        <source :src="state.meta.direct_url" :type="state.meta.mimetype" />
        Your browser doesn’t support playing {{ state.meta.mimetype }}.
      </video>
    </div>

    <div v-else-if="state.mode === Modes.PDF">
      <object
        class="w-full h-[800px] rounded-md"
        :data="state.meta.direct_url"
        :type="state.meta.mimetype"
      >
        Your web browser does not support displaying PDF files.
      </object>
    </div>

    <div
      v-else-if="!!state.formatted && state.mode === Modes.MARKDOWN"
      class="prose max-w-none"
      v-html="state.formatted"
    />

    <pre v-else-if="state.mode === Modes.ARCHIVE" class="overflow-x-scroll max-h-[600px]">{{
      state.meta.archive_files.join("\n")
    }}</pre>

    <div v-else-if="!!state.formatted && state.mode === Modes.CSV" class="space-y-4">
      <Table>
        <TableRow v-for="(row, key) in state.formatted?.data?.slice(0, csvRows)" :key="key">
          <TableCell v-for="(cell, ckey) in row" :key="ckey">{{ cell }}</TableCell>
        </TableRow>
      </Table>
      <div class="flex justify-between">
        Showing {{ Math.min(csvRows, state.formatted?.data?.length) }} of
        {{ state.formatted?.data?.length }} rows
        <Button v-if="csvRows < state.formatted?.data?.length" @click="csvRows += 250"
          >Show more</Button
        >
      </div>
    </div>

    <HighlightJS
      v-else-if="!!state.content && state.mode === Modes.TEXT"
      class="p-4"
      :class="[wrap ? 'whitespace-pre-wrap wrap-break-word' : 'overflow-x-scroll']"
      :language="state.meta.language"
      :code="state.content"
    />
  </CardContent>
</template>

<script setup>
import { ref } from "vue";
import HighlightJS from "@/components/HighlightJS.js";
import Modes from "@/components/display/fileModes.js";
import { Button } from "@/components/ui/button/index.js";
import { CardContent } from "@/components/ui/card/index.js";
import { Table, TableCell, TableRow } from "@/components/ui/table/index.js";

defineProps({
  state: { type: Object, required: true },
  wrap: { type: Boolean, default: false },
});

const csvRows = ref(250);
</script>
