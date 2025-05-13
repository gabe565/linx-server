<template>
  <Card v-if="items.length">
    <CardHeader>
      <CardTitle>Your Uploads</CardTitle>
    </CardHeader>

    <CardContent>
      <ul class="flex flex-col gap-3 justify-center justify-items-center">
        <li v-for="(item, key) in items" :key="item.filename || key">
          <Card v-if="'progress' in item" class="relative py-4 overflow-hidden">
            <CardHeader class="px-4">
              <CardTitle class="min-w-0 wrap-break-word animate-pulse">{{
                item.original_name
              }}</CardTitle>
              <CardDescription>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger class="flex gap-3 items-center">
                      <strong> {{ Math.round(item.progress.progress * 100) }}% </strong>
                      <span v-if="item.progress.estimated" class="before:content-['·'] before:pr-3">
                        {{ formatDuration(item.progress.estimated) }}
                      </span>
                    </TooltipTrigger>
                    <TooltipContent class="flex flex-col items-center">
                      <span v-if="item.progress.loaded && item.progress.total">
                        {{ formatBytes(item.progress.loaded, { decimals: 1, hideUnit: true }) }} /
                        {{ formatBytes(item.progress.total, { decimals: 1 }) }}
                      </span>
                      <span v-if="item.progress.rate">
                        {{ formatBitsPerSecond(item.progress.rate) }}
                      </span>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </CardDescription>
              <CardAction>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger as-child>
                      <Button
                        variant="secondary"
                        size="icon"
                        @click.prevent="item.controller.abort()"
                        class="w-18"
                      >
                        <span class="sr-only">Cancel</span>
                        <CloseIcon />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Cancel</TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </CardAction>
            </CardHeader>
            <Progress
              v-if="'progress' in item"
              :model-value="item.progress.progress * 100"
              class="absolute bottom-0 left-0 rounded-none animate-pulse"
            />
          </Card>

          <Card v-else class="py-4">
            <CardHeader class="px-4">
              <CardTitle class="min-w-0">
                <RouterLink :to="`/${item.filename}`" class="wrap-break-word link">
                  {{ item.original_name || item.filename }}
                </RouterLink>
              </CardTitle>

              <CardDescription v-if="item.expiry">
                <UseTimeAgo
                  v-if="item.expiry > 0"
                  v-slot="{ timeAgo }"
                  :time="new Date(item.expiry * 1000)"
                  :show-second="true"
                  update-interval="1000"
                >
                  expires {{ timeAgo }}
                </UseTimeAgo>
              </CardDescription>

              <Dialog>
                <CardAction v-if="smAndLarger">
                  <TooltipProvider>
                    <Tooltip>
                      <DialogTrigger>
                        <TooltipTrigger as-child>
                          <Button variant="secondary" size="icon" class="rounded-r-none">
                            <span class="sr-only">Info</span>
                            <InfoIcon />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>Info</TooltipContent>
                      </DialogTrigger>
                    </Tooltip>
                    <Tooltip>
                      <TooltipTrigger as-child>
                        <Button
                          variant="secondary"
                          size="icon"
                          @click.prevent="upload.copy(item)"
                          class="rounded-none"
                        >
                          <span class="sr-only">Copy</span>
                          <CopyIcon />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>Copy Link</TooltipContent>
                    </Tooltip>
                    <Tooltip>
                      <TooltipTrigger as-child>
                        <Button
                          variant="destructive"
                          size="icon"
                          @click.prevent="deleteItem(item)"
                          class="rounded-l-none"
                        >
                          <span class="sr-only">Delete</span>
                          <DeleteIcon />
                        </Button>
                      </TooltipTrigger>
                      <TooltipContent>Delete</TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </CardAction>

                <CardAction v-else>
                  <DropdownMenu>
                    <DropdownMenuTrigger>
                      <Button variant="ghost" size="icon" class="rounded-full">
                        <MoreIcon />
                      </Button>
                    </DropdownMenuTrigger>

                    <DropdownMenuContent side="left">
                      <DialogTrigger as-child>
                        <DropdownMenuItem>
                          <InfoIcon />
                          Info
                        </DropdownMenuItem>
                      </DialogTrigger>
                      <DropdownMenuItem @click.prevent="upload.copy(item)">
                        <CopyIcon />
                        Copy
                      </DropdownMenuItem>
                      <DropdownMenuItem @click.prevent="deleteItem(item)">
                        <DeleteIcon />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </CardAction>

                <UploadInfo :item="item" />
              </Dialog>
            </CardHeader>
          </Card>
        </li>
      </ul>
    </CardContent>
  </Card>
</template>

<script setup>
import { Button } from "@/components/ui/button/index.js";
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card/index.js";
import { Dialog, DialogTrigger } from "@/components/ui/dialog/index.js";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu/index.js";
import { Progress } from "@/components/ui/progress/index.js";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip/index.js";
import UploadInfo from "@/components/upload/UploadInfo.vue";
import { useUploadStore } from "@/stores/upload.js";
import { formatBitsPerSecond, formatBytes } from "@/util/bytes.js";
import { formatDuration } from "@/util/time.js";
import { UseTimeAgo } from "@vueuse/components";
import { breakpointsTailwind, useBreakpoints } from "@vueuse/core";
import { computed } from "vue";
import MoreIcon from "~icons/ic/round-more-horiz";
import CloseIcon from "~icons/material-symbols/close-rounded";
import CopyIcon from "~icons/material-symbols/content-copy-rounded";
import DeleteIcon from "~icons/material-symbols/delete-rounded";
import InfoIcon from "~icons/material-symbols/info-rounded";

const showAuth = defineModel("showAuth");
const breakpoints = useBreakpoints(breakpointsTailwind);
const smAndLarger = breakpoints.greaterOrEqual("sm");
const upload = useUploadStore();
const items = computed(() => {
  return Object.values(upload.inProgress).concat(upload.uploads);
});

const deleteItem = async (item) => {
  try {
    await upload.deleteItem(item);
  } catch (err) {
    if (err.response?.status === 401) {
      showAuth.value = true;
    }
  }
};
</script>
