import hljs from "highlight.js/lib/core";
import { Marked } from "marked";
import markedAlert from "marked-alert";
import { markedHighlight } from "marked-highlight";
import { loadLanguage } from "@/util/extensions.ts";

// Common fence names that loadLanguage doesn't know
const aliases: Record<string, string> = {
  c: "cpp",
  "c++": "cpp",
  cs: "csharp",
  docker: "dockerfile",
  golang: "go",
  html: "xml",
  js: "javascript",
  jsx: "javascript",
  kt: "kotlin",
  make: "makefile",
  ps1: "powershell",
  py: "python",
  rb: "ruby",
  rs: "rust",
  sh: "bash",
  shell: "bash",
  svg: "xml",
  ts: "typescript",
  tsx: "typescript",
  yml: "yaml",
  zsh: "bash",
};

const resolveLanguage = async (lang: string): Promise<string | undefined> => {
  lang = lang.toLowerCase();
  if (hljs.getLanguage(lang)) return lang;
  const name = aliases[lang] ?? lang;
  if (!hljs.getLanguage(name) && !(await loadLanguage(name))) return;
  return name;
};

export const marked = new Marked().use(markedAlert()).use(
  markedHighlight({
    async: true,
    langPrefix: "hljs language-",
    async highlight(code, lang) {
      if (!lang) return code;
      const language = await resolveLanguage(lang).catch(() => undefined);
      return language ? hljs.highlight(code, { language }).value : code;
    },
  }),
);
