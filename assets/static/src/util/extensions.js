import hljs from "highlight.js/lib/core";

export const getExtension = (filename) => filename.split(".").pop();

export const loadLanguage = async (language) => {
  let lang;
  switch (language) {
    case "apache":
      lang = await import("highlight.js/lib/languages/apache");
      break;
    case "applescript":
      lang = await import("highlight.js/lib/languages/applescript");
      break;
    case "autohotkey":
      lang = await import("highlight.js/lib/languages/autohotkey");
      break;
    case "bash":
      lang = await import("highlight.js/lib/languages/bash");
      break;
    case "basic":
      lang = await import("highlight.js/lib/languages/basic");
      break;
    case "clojure":
      lang = await import("highlight.js/lib/languages/clojure");
      break;
    case "cmake":
      lang = await import("highlight.js/lib/languages/cmake");
      break;
    case "coffeescript":
      lang = await import("highlight.js/lib/languages/coffeescript");
      break;
    case "cpp":
      lang = await import("highlight.js/lib/languages/cpp");
      break;
    case "csharp":
      lang = await import("highlight.js/lib/languages/csharp");
      break;
    case "css":
      lang = await import("highlight.js/lib/languages/css");
      break;
    case "d":
      lang = await import("highlight.js/lib/languages/d");
      break;
    case "dart":
      lang = await import("highlight.js/lib/languages/dart");
      break;
    case "diff":
      lang = await import("highlight.js/lib/languages/diff");
      break;
    case "dockerfile":
      lang = await import("highlight.js/lib/languages/dockerfile");
      break;
    case "dos":
      lang = await import("highlight.js/lib/languages/dos");
      break;
    case "elm":
      lang = await import("highlight.js/lib/languages/elm");
      break;
    case "erlang":
      lang = await import("highlight.js/lib/languages/erlang");
      break;
    case "fortran":
      lang = await import("highlight.js/lib/languages/fortran");
      break;
    case "go":
      lang = await import("highlight.js/lib/languages/go");
      break;
    case "ini":
      lang = await import("highlight.js/lib/languages/ini");
      break;
    case "java":
      lang = await import("highlight.js/lib/languages/java");
      break;
    case "javascript":
      lang = await import("highlight.js/lib/languages/javascript");
      break;
    case "json":
      lang = await import("highlight.js/lib/languages/json");
      break;
    case "kotlin":
      lang = await import("highlight.js/lib/languages/kotlin");
      break;
    case "latex":
      lang = await import("highlight.js/lib/languages/latex");
      break;
    case "less":
      lang = await import("highlight.js/lib/languages/less");
      break;
    case "lisp":
      lang = await import("highlight.js/lib/languages/lisp");
      break;
    case "lua":
      lang = await import("highlight.js/lib/languages/lua");
      break;
    case "nginx":
      lang = await import("highlight.js/lib/languages/nginx");
      break;
    case "objectivec":
      lang = await import("highlight.js/lib/languages/objectivec");
      break;
    case "ocaml":
      lang = await import("highlight.js/lib/languages/ocaml");
      break;
    case "perl":
      lang = await import("highlight.js/lib/languages/perl");
      break;
    case "php":
      lang = await import("highlight.js/lib/languages/php");
      break;
    case "powershell":
      lang = await import("highlight.js/lib/languages/powershell");
      break;
    case "protobuf":
      lang = await import("highlight.js/lib/languages/protobuf");
      break;
    case "python":
      lang = await import("highlight.js/lib/languages/python");
      break;
    case "ruby":
      lang = await import("highlight.js/lib/languages/ruby");
      break;
    case "rust":
      lang = await import("highlight.js/lib/languages/rust");
      break;
    case "scala":
      lang = await import("highlight.js/lib/languages/scala");
      break;
    case "scheme":
      lang = await import("highlight.js/lib/languages/scheme");
      break;
    case "scss":
      lang = await import("highlight.js/lib/languages/scss");
      break;
    case "sql":
      lang = await import("highlight.js/lib/languages/sql");
      break;
    case "tcl":
      lang = await import("highlight.js/lib/languages/tcl");
      break;
    case "typescript":
      lang = await import("highlight.js/lib/languages/typescript");
      break;
    case "vue": {
      const loadVue = async () => {
        const lang = await import("highlightjs-vue/dist/highlightjs-vue.esm.js");
        lang.default(hljs);
      };

      await Promise.all([
        loadVue(),
        loadLanguage("javascript"),
        loadLanguage("typescript"),
        loadLanguage("css"),
        loadLanguage("xml"),
      ]);
      return true;
    }
    case "xml":
      lang = await import("highlight.js/lib/languages/xml");
      break;
    case "yaml":
      lang = await import("highlight.js/lib/languages/yaml");
      break;
    default:
      return false;
  }
  hljs.registerLanguage(language, lang.default);
  return true;
};
