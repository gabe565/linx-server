<template>
  <div class="flex flex-col min-h-screen min-h-svh bg-background text-foreground">
    <header class="grid grid-cols-3 border-b bg-surface px-4 py-2 items-center">
      <div>
        <Button variant="link" class="p-0" as-child>
          <RouterLink to="/">
            <h1 class="text-2xl font-semibold">{{ config.site.site_name }}</h1>
          </RouterLink>
        </Button>
      </div>

      <NavigationMenu class="justify-self-center">
        <NavigationMenuList>
          <NavigationMenuItem v-for="route in routes" :key="route.name">
            <NavigationMenuLink as-child>
              <RouterLink :to="route.path">{{ route.name }}</RouterLink>
            </NavigationMenuLink>
          </NavigationMenuItem>
        </NavigationMenuList>
      </NavigationMenu>

      <div class="justify-self-end">
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger as-child :key="mode">
              <Button variant="ghost" @click="mode = nextMode" class="rounded-full">
                <component :is="modeIcon" />
                <span class="sr-only">Change to {{ nextMode }} mode</span>
              </Button>
            </TooltipTrigger>
            <TooltipContent>Change to {{ nextMode }} mode</TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger as-child>
              <Button
                as="a"
                variant="ghost"
                href="https://github.com/gabe565/linx-server"
                target="_blank"
                class="rounded-full"
              >
                <GitHubIcon />
                <span class="sr-only">View source on GitHub</span>
              </Button>
            </TooltipTrigger>
            <TooltipContent>View source on GitHub</TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>
    </header>

    <main class="flex-1 content-center p-6">
      <router-view />
    </main>

    <Toaster />
  </div>
</template>

<script setup>
import { Button } from "@/components/ui/button";
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
} from "@/components/ui/navigation-menu";
import { Toaster } from "@/components/ui/sonner";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip/index.js";
import { useConfigStore } from "@/stores/config";
import { useColorMode } from "@vueuse/core";
import { computed } from "vue";
import { useRouter } from "vue-router";
import DarkIcon from "~icons/material-symbols/brightness-2-rounded";
import LightIcon from "~icons/material-symbols/brightness-5-rounded";
import AutoIcon from "~icons/material-symbols/brightness-auto-rounded";
import GitHubIcon from "~icons/simple-icons/github";

const config = useConfigStore();

const router = useRouter();
const routes = computed(() =>
  router
    .getRoutes()
    .filter((route) => route.meta?.navigation)
    .concat(config.site?.custom_pages?.map((v) => ({ name: v, path: `/${v}` })) || []),
);

const mode = useColorMode({ disableTransition: false, emitAuto: true });

const nextMode = computed(() => {
  if (mode.value === "auto") return "dark";
  if (mode.value === "dark") return "light";
  return "auto";
});

const modeIcon = computed(() => {
  if (mode.value === "auto") return AutoIcon;
  if (mode.value === "light") return LightIcon;
  return DarkIcon;
});
</script>
