package util

import "strings"

func ExtensionToHlLang(filename, extension string) string {
	hlExt, exists := extensionToHl[extension]
	if !exists {
		hlExt = "text"
		if strings.HasPrefix(strings.ToLower(filename), "dockerfile") {
			return "dockerfile"
		}
	}
	return hlExt
}

func SupportedBinExtension(extension string) bool {
	_, exists := extensionToHl[extension]
	return exists
}

//nolint:gochecknoglobals
var extensionToHl = map[string]string{
	"ahk":         "autohotkey",
	"apache":      "apache",
	"applescript": "applescript",
	"bas":         "basic",
	"bash":        "bash",
	"bat":         "dos",
	"c":           "cpp",
	"clj":         "clojure",
	"cmake":       "cmake",
	"coffee":      "coffeescript",
	"cpp":         "cpp",
	"cs":          "csharp",
	"css":         "css",
	"d":           "d",
	"dart":        "dart",
	"diff":        "diff",
	"dockerfile":  "dockerfile",
	"elm":         "elm",
	"erl":         "erlang",
	"for":         "fortran",
	"go":          "go",
	"h":           "cpp",
	"htm":         "xml",
	"html":        "xml",
	"ini":         "ini",
	"java":        "java",
	"js":          "javascript",
	"json":        "json",
	"jsp":         "java",
	"kt":          "kotlin",
	"less":        "less",
	"lisp":        "lisp",
	"lua":         "lua",
	"m":           "objectivec",
	"nginx":       "nginx",
	"ocaml":       "ocaml",
	"php":         "php",
	"pl":          "perl",
	"proto":       "protobuf",
	"ps":          "powershell",
	"py":          "python",
	"rb":          "ruby",
	"rs":          "rust",
	"scala":       "scala",
	"scm":         "scheme",
	"scpt":        "applescript",
	"scss":        "scss",
	"sh":          "bash",
	"sql":         "sql",
	"tcl":         "tcl",
	"tex":         "latex",
	"toml":        "ini",
	"ts":          "typescript",
	"vue":         "vue",
	"xml":         "xml",
	"yaml":        "yaml",
	"yml":         "yaml",
	"zsh":         "bash",
}
