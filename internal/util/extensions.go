package util

import (
	"path/filepath"
	"strings"

	"gabe565.com/linx-server/internal/backends"
)

func InferLang(fileName string, meta backends.Metadata) string {
	fileExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(fileName), "."))
	if lang, found := extensionToLang[fileExt]; found {
		return lang
	}

	prettyName := fileName
	if meta.OriginalName != "" {
		originalExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(meta.OriginalName), "."))
		if lang, found := extensionToLang[originalExt]; found {
			return lang
		}

		prettyName = meta.OriginalName
	}

	lowerName := strings.ToLower(prettyName)
	switch {
	case strings.HasPrefix(lowerName, "dockerfile"), strings.HasPrefix(lowerName, "containerfile"):
		return "dockerfile"
	case strings.HasPrefix(lowerName, "makefile"):
		return "makefile"
	case strings.HasPrefix(meta.Mimetype, "text/"):
		return "text"
	default:
		return ""
	}
}

//nolint:gochecknoglobals
var extensionToLang = map[string]string{
	"ahk":           "autohotkey",
	"apache":        "apache",
	"applescript":   "applescript",
	"bas":           "basic",
	"bash":          "bash",
	"bat":           "dos",
	"c":             "cpp",
	"c++":           "cpp",
	"cc":            "cpp",
	"cjs":           "javascript",
	"clj":           "clojure",
	"cmake":         "cmake",
	"coffee":        "coffeescript",
	"containerfile": "dockerfile",
	"cp":            "cpp",
	"cpp":           "cpp",
	"cs":            "csharp",
	"css":           "css",
	"cts":           "typescript",
	"cxx":           "cpp",
	"d":             "d",
	"dart":          "dart",
	"diff":          "diff",
	"dockerfile":    "dockerfile",
	"elm":           "elm",
	"erl":           "erlang",
	"for":           "fortran",
	"go":            "go",
	"gql":           "graphql",
	"gradle":        "gradle",
	"graphql":       "graphql",
	"h":             "cpp",
	"htm":           "xml",
	"html":          "xml",
	"ini":           "ini",
	"java":          "java",
	"js":            "javascript",
	"json":          "json",
	"jsp":           "xml",
	"jsx":           "javascript",
	"kt":            "kotlin",
	"less":          "less",
	"lisp":          "lisp",
	"lua":           "lua",
	"m":             "objectivec",
	"mjs":           "javascript",
	"mts":           "typescript",
	"nginx":         "nginx",
	"nix":           "nix",
	"ocaml":         "ocaml",
	"php":           "php",
	"pl":            "perl",
	"proto":         "protobuf",
	"ps1":           "powershell",
	"py":            "python",
	"r":             "r",
	"rb":            "ruby",
	"rs":            "rust",
	"scala":         "scala",
	"scm":           "scheme",
	"scpt":          "applescript",
	"scss":          "scss",
	"sh":            "bash",
	"sql":           "sql",
	"swift":         "swift",
	"tcl":           "tcl",
	"tex":           "latex",
	"toml":          "ini",
	"ts":            "typescript",
	"tsx":           "typescript",
	"vue":           "vue",
	"xml":           "xml",
	"yaml":          "yaml",
	"yml":           "yaml",
	"zsh":           "bash",
}
