import tailwindcss from "@tailwindcss/vite";
import vue from "@vitejs/plugin-vue";
import path from "node:path";
import Icons from "unplugin-icons/vite";
import { defineConfig } from "vite";

export default defineConfig({
  base: "./",
  build: {
    manifest: true,
    rollupOptions: {
      input: {
        main: "src/main.js",
        fouc: "src/fouc.js",
      },
    },
  },
  plugins: [vue(), tailwindcss(), Icons()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
});
