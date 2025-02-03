import pluginJs from "@eslint/js";
import pluginPrettier from "eslint-plugin-prettier/recommended";
import globals from "globals";

export default [
  { languageOptions: { globals: globals.browser } },
  { ignores: ["dist"] },
  pluginJs.configs.recommended,
  pluginPrettier,
];
