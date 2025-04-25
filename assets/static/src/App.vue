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
        <Button
          as="a"
          size="icon"
          variant="ghost"
          href="https://github.com/gabe565/linx-server"
          target="_blank"
          class="rounded-full"
        >
          <GitHubIcon />
          <span class="sr-only">Source code on GitHub</span>
        </Button>
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
import { useConfigStore } from "@/stores/config";
import { useColorMode } from "@vueuse/core";
import { useRouter } from "vue-router";
import GitHubIcon from "~icons/simple-icons/github";

const config = useConfigStore();

const routes = useRouter()
  .getRoutes()
  .filter((route) => route.meta?.navigation);

useColorMode({ disableTransition: false });
</script>
