package apikeys

import (
	"bufio"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"

	"gabe565.com/linx-server/internal/auth/keyhash"
	"gabe565.com/linx-server/internal/util"
)

type AuthOptions struct {
	AuthFile      string
	UnauthMethods []string
	BasicAuth     bool
	SiteName      string
	SitePath      string
}

type Middleware struct {
	successHandler http.Handler
	authKeys       []string
	o              AuthOptions
}

func ReadAuthKeys(authFile string) []string {
	var authKeys []string

	f, err := os.Open(authFile)
	if err != nil {
		slog.Error("Failed to open authfile", "error", err)
		os.Exit(1)
	}
	defer func() {
		_ = f.Close()
	}()

	scanner := bufio.NewScanner(f)
	for i := 0; scanner.Scan(); i++ {
		if len(scanner.Bytes()) == 0 || scanner.Bytes()[0] == '#' {
			continue
		}

		key := scanner.Text()
		if i := strings.Index(key, "#"); i != -1 {
			key = key[:i]
		}
		key = strings.TrimSpace(key)

		if key == "" {
			continue
		}

		if !strings.HasPrefix(key, keyhash.KeyPrefix) {
			key = keyhash.KeyPrefix + key
		}

		if !keyhash.IsValidHash(key) {
			slog.Warn("Skipping invalid key in authfile", "line", i+1)
			continue
		}

		authKeys = append(authKeys, key)
	}

	if err := scanner.Err(); err != nil {
		slog.Error("Scanner error while reading authfile", "error", err)
		os.Exit(1) //nolint:gocritic
	}

	return slices.Clip(authKeys)
}

func (a Middleware) getSitePrefix() string {
	prefix := a.o.SitePath
	if len(prefix) == 0 || prefix[0] != '/' {
		prefix = "/" + prefix
	}
	return prefix
}

func (a Middleware) goodAuthorizationHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Location", a.getSitePrefix())
	w.WriteHeader(http.StatusFound)
}

func (a Middleware) badAuthorizationHandler(w http.ResponseWriter, _ *http.Request) {
	if a.o.BasicAuth {
		rs := ""
		if a.o.SiteName != "" {
			rs = " realm=" + strconv.Quote(a.o.SiteName)
		}
		w.Header().Set("WWW-Authenticate", `Basic`+rs)
	}
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func (a Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var successHandler http.Handler
	prefix := a.getSitePrefix()

	if r.URL.Path == prefix+"auth" {
		successHandler = http.HandlerFunc(a.goodAuthorizationHandler)
	} else {
		successHandler = a.successHandler
	}

	if slices.Contains(a.o.UnauthMethods, r.Method) && r.URL.Path != prefix+"auth" {
		// allow unauthenticated methods
		successHandler.ServeHTTP(w, r)
		return
	}

	key := util.TryPathUnescape(r.Header.Get("Linx-Api-Key"))
	if key == "" && a.o.BasicAuth {
		_, password, ok := r.BasicAuth()
		if ok {
			key = password
		}
	}

	result, err := keyhash.CheckList(a.authKeys, key)
	if err != nil || !result {
		http.HandlerFunc(a.badAuthorizationHandler).ServeHTTP(w, r)
		return
	}

	successHandler.ServeHTTP(w, r)
}

func NewAPIKeysMiddleware(o AuthOptions) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return Middleware{
			successHandler: h,
			authKeys:       ReadAuthKeys(o.AuthFile),
			o:              o,
		}
	}
}
