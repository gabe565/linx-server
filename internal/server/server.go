package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"gabe565.com/linx-server/assets"
	"gabe565.com/linx-server/internal/auth/apikeys"
	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/backends/localfs"
	"gabe565.com/linx-server/internal/backends/s3"
	"gabe565.com/linx-server/internal/cleanup"
	"gabe565.com/linx-server/internal/config"
	"gabe565.com/linx-server/internal/custompages"
	"gabe565.com/linx-server/internal/handlers"
	"gabe565.com/linx-server/internal/headers"
	"gabe565.com/linx-server/internal/templates"
	"gabe565.com/linx-server/internal/torrent"
	"gabe565.com/linx-server/internal/upload"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func Setup(ctx context.Context) (*chi.Mux, error) {
	var err error

	switch config.Default.SiteURL.Path {
	case "", "/":
		config.Default.SiteURL.Path = "/"
	default:
		config.Default.SiteURL.Path = "/" + strings.Trim(config.Default.SiteURL.Path, "/") + "/"
	}
	config.Default.SelifPath = strings.Trim(config.Default.SelifPath, "/") + "/"

	if config.Default.S3.Bucket != "" {
		config.StorageBackend, err = s3.New(ctx,
			config.Default.S3.Bucket,
			config.Default.S3.Region,
			config.Default.S3.Endpoint,
			config.Default.S3.ForcePathStyle,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 backend: %w", err)
		}
	} else {
		err := os.MkdirAll(config.Default.FilesPath, 0o755)
		if err != nil {
			return nil, fmt.Errorf("could not create files directory: %w", err)
		}

		err = os.MkdirAll(config.Default.MetaPath, 0o700)
		if err != nil {
			return nil, fmt.Errorf("could not create metadata directory: %w", err)
		}

		config.StorageBackend = localfs.New(config.Default.MetaPath, config.Default.FilesPath)
	}

	if config.Default.CleanupEvery.Duration > 0 {
		if backend, ok := config.StorageBackend.(backends.ListBackend); ok {
			go cleanup.PeriodicCleanup(backend, config.Default.CleanupEvery.Duration, config.Default.NoLogs)
		}
	}

	// Template setup
	config.Templates, err = templates.Load(assets.Templates())
	if err != nil {
		return nil, fmt.Errorf("could not load templates: %w", err)
	}

	config.TimeStarted = time.Now()
	config.TimeStartedStr = strconv.FormatInt(config.TimeStarted.Unix(), 10)

	if err := assets.LoadManifest(); err != nil {
		return nil, fmt.Errorf("failed to load Vite manifest: %w", err)
	}

	// Routing setup
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(func(next http.Handler) http.Handler {
		if config.Default.SiteURL.Path == "/" {
			return middleware.RedirectSlashes(next)
		}
		redirectSlashes := middleware.RedirectSlashes(next)
		fn := func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.URL.Path == config.Default.SiteURL.Path:
				next.ServeHTTP(w, r)
			case r.URL.Path == strings.TrimSuffix(config.Default.SiteURL.Path, "/"):
				http.Redirect(w, r, config.Default.SiteURL.String(), http.StatusPermanentRedirect)
			default:
				redirectSlashes.ServeHTTP(w, r)
			}
		}
		return http.HandlerFunc(fn)
	})
	if config.Default.SiteURL.Path != "/" {
		r.Use(middleware.StripPrefix(strings.TrimSuffix(config.Default.SiteURL.Path, "/")))
	}
	if config.Default.Header.RealIP {
		r.Use(middleware.RealIP)
	}

	r.Use(middleware.Heartbeat("/ping"))

	if !config.Default.NoLogs {
		r.Use(middleware.Logger)
	}

	r.Use(middleware.Recoverer)
	r.Use(middleware.GetHead)
	r.Use(ContentSecurityPolicy(CSPOptions{
		Policy:         config.Default.Header.ContentSecurityPolicy,
		ReferrerPolicy: config.Default.Header.ReferrerPolicy,
		Frame:          config.Default.Header.XFrameOptions,
	}))
	r.Use(headers.AddHeaders(config.Default.Header.AddHeaders))

	if config.Default.Auth.File != "" {
		r.Use(apikeys.NewAPIKeysMiddleware(apikeys.AuthOptions{
			AuthFile:      config.Default.Auth.File,
			UnauthMethods: []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace},
			BasicAuth:     config.Default.Auth.Basic,
			SiteName:      config.Default.SiteName,
			SitePath:      config.Default.SiteURL.Path,
		}))
	}

	if config.Default.Auth.File == "" || config.Default.Auth.Basic {
		r.Get("/", handlers.Index)
		r.Get("/paste", handlers.Paste)
	} else {
		r.Get("/", http.RedirectHandler(path.Join(config.Default.SiteURL.Path, "API"), http.StatusSeeOther).ServeHTTP)
		r.Get("/paste", http.RedirectHandler(path.Join(config.Default.SiteURL.Path, "API/"), http.StatusSeeOther).ServeHTTP)
	}

	r.Get("/api", handlers.APIDoc)
	r.Get("/API",
		http.RedirectHandler(path.Join(config.Default.SiteURL.Path, "api"), http.StatusPermanentRedirect).ServeHTTP,
	)

	r.Group(func(r chi.Router) {
		r.Use(rateLimit(config.Default.Limit.UploadMaxRequests, config.Default.Limit.UploadInterval.Duration))

		r.Post("/upload", upload.POSTHandler)
		r.Put("/upload", upload.PUTHandler)
		r.Put("/upload/{name}", upload.PUTHandler)
		if config.Default.RemoteUploads {
			r.Get("/upload", upload.Remote)

			if config.Default.Auth.RemoteFile != "" {
				config.RemoteAuthKeys = apikeys.ReadAuthKeys(config.Default.Auth.RemoteFile)
			}
		}

		r.Delete("/{name}", handlers.Delete)
	})

	r.Group(func(r chi.Router) {
		r.Use(rateLimit(config.Default.Limit.FileMaxRequests, config.Default.Limit.FileInterval.Duration))

		r.Get("/{name}", handlers.FileAccessHandler)
		r.Post("/{name}", handlers.FileAccessHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(rateLimit(config.Default.Limit.FileMaxRequests, config.Default.Limit.FileInterval.Duration))

		r.Get(path.Join("/", config.Default.SelifPath, "{name}"), handlers.FileServeHandler)
		if !config.Default.NoTorrent {
			r.Get("/torrent/{name}", torrent.FileTorrentHandler)
			r.Get("/{name}/torrent", func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/torrent/"+chi.URLParam(r, "name"), http.StatusMovedPermanently)
			})
		}
	})

	if config.Default.CustomPagesDir != "" {
		custompages.InitializeCustomPages(config.Default.CustomPagesDir)
		for fileName := range custompages.Names {
			r.Get("/"+fileName, handlers.MakeCustomPage(fileName))
		}
	}

	r.NotFound(handlers.AssetHandler)

	return r, nil
}

func rateLimit(requestLimit int, windowLength time.Duration) func(next http.Handler) http.Handler {
	limiter := httprate.NewRateLimiter(requestLimit, windowLength,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
			handlers.Oops(w, r, handlers.RespAUTO, "Too many requests")
		}),
	)
	return limiter.Handler
}
