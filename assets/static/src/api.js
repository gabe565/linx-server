import hljs from "highlight.js/lib/core";
import { loadLanguage } from "./bin.js";

export const initAPI = async () => {
  await loadLanguage("bash");
  await loadLanguage("json");
  hljs.default.highlightAll();
};
