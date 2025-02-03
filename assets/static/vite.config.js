import { defineConfig } from "vite";

export default defineConfig({
  server: {
    cors: {
      origin: "http://localhost:8080",
    },
  },
  build: {
    manifest: "manifest.json",
    rollupOptions: {
      input: "src/main.js",
    },
  },
});
