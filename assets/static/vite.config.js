import autoprefixer from "autoprefixer";
import { defineConfig } from "vite";

export default defineConfig({
  server: {
    cors: {
      origin: "http://localhost:8080",
    },
  },
  build: {
    manifest: true,
    rollupOptions: {
      input: "src/main.js",
    },
  },
  css: {
    postcss: {
      plugins: [autoprefixer()],
    },
  },
});
