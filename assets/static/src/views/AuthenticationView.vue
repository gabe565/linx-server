<template>
  <div class="container max-w-5xl mx-auto p-4">
    <Card>
      <CardHeader>
        <CardTitle>Authentication Required</CardTitle>
      </CardHeader>
      <CardContent class="flex flex-col gap-4">
        <p>Uploads require an API key.</p>

        <form class="flex gap-2 m-auto" @submit.prevent="router.back()">
          <Label class="w-full flex flex-wrap">
            API Key
            <Input type="password" v-model="config.apiKey" class="flex-1 min-w-50" />
          </Label>

          <Button type="submit" variant="secondary">Submit</Button>
        </form>
      </CardContent>
    </Card>
  </div>
</template>

<script setup>
import { Button } from "@/components/ui/button/index.js";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card/index.js";
import { Input } from "@/components/ui/input/index.js";
import { Label } from "@/components/ui/label/index.js";
import { useConfigStore } from "@/stores/config.js";
import { useRouter } from "vue-router";

const config = useConfigStore();
const router = useRouter();
</script>
