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
              <CardTitle class="min-w-0 wrap-break-word">{{ item.original_name }}</CardTitle>
              <CardDescription>
                <Skeleton class="h-2.5 mt-2 mb-0.5 w-1/4" />
              </CardDescription>
              <CardAction>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
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
              v-model="item.progress"
              class="absolute bottom-0 left-0 rounded-none"
            />
          </Card>

          <Card v-else class="py-4">
            <CardHeader class="px-4">
              <CardTitle class="min-w-0">
                <RouterLink
                  :to="`/${item.filename}`"
                  class="wrap-break-word text-blue-600 dark:text-blue-400 hover:underline"
                >
                  {{ item.filename }}
                  <span
                    v-if="item.original_name && item.filename !== item.original_name"
                    class="text-sm"
                  >
                    ({{ item.original_name }})
                  </span>
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

              <CardAction>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger>
                      <Button
                        variant="secondary"
                        size="icon"
                        @click.prevent="upload.copy(item)"
                        class="rounded-r-none"
                      >
                        <span class="sr-only">Copy</span>
                        <CopyIcon />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>Copy Link</TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger>
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
import { Progress } from "@/components/ui/progress/index.js";
import { Skeleton } from "@/components/ui/skeleton/index.js";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip/index.js";
import { useUploadStore } from "@/stores/upload.js";
import { UseTimeAgo } from "@vueuse/components";
import { computed } from "vue";
import CloseIcon from "~icons/material-symbols/close-rounded";
import CopyIcon from "~icons/material-symbols/content-copy-rounded";
import DeleteIcon from "~icons/material-symbols/delete-rounded";

const upload = useUploadStore();

const emit = defineEmits(["delete", "error"]);

const items = computed(() => {
  return Object.values(upload.inProgress).concat(upload.uploads);
});

const deleteItem = async (item) => {
  try {
    await upload.deleteItem(item);
  } catch (err) {
    emit("error", err);
    throw err;
  }
};
</script>
