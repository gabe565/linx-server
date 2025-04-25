import js from "@eslint/js";
import skipFormatting from "@vue/eslint-config-prettier/skip-formatting";
import pluginOxlint from "eslint-plugin-oxlint";
import pluginVue from "eslint-plugin-vue";
import { defineConfig, globalIgnores } from "eslint/config";
import globals from "globals";

export default defineConfig([
  {
    name: "app/files-to-lint",
    files: ["**/*.{js,mjs,jsx,vue}"],
  },

  globalIgnores(["**/dist/**", "**/dist-ssr/**", "**/coverage/**"]),

  {
    languageOptions: {
      globals: {
        ...globals.browser,
        __dirname: true,
      },
    },
  },

  js.configs.recommended,
  ...pluginVue.configs["flat/essential"],
  ...pluginOxlint.configs["flat/recommended"],
  skipFormatting,
  {
    rules: {
      "no-unused-vars": ["error", { varsIgnorePattern: "^_", argsIgnorePattern: "^_" }],
    },
  },
  {
    files: ["src/components/ui/**/*.vue"],
    rules: {
      "vue/multi-word-component-names": "off",
    },
  },
]);
