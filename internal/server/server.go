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
)

func Setup() (*chi.Mux, error) {
	// make directories if needed
	err := os.MkdirAll(config.Default.FilesPath, 0o755)
	if err != nil {
		return nil, fmt.Errorf("could not create files directory: %w", err)
	}

	err = os.MkdirAll(config.Default.MetaPath, 0o700)
	if err != nil {
		return nil, fmt.Errorf("could not create metadata directory: %w", err)
	}

	switch config.Default.SiteURL.Path {
	case "", "/":
		config.Default.SiteURL.Path = "/"
	default:
		config.Default.SiteURL.Path = "/" + strings.Trim(config.Default.SiteURL.Path, "/") + "/"
	}
	config.Default.SelifPath = strings.Trim(config.Default.SelifPath, "/") + "/"

	if config.Default.S3Bucket != "" {
		config.StorageBackend, err = s3.NewS3Backend(context.Background(), config.Default.S3Bucket, config.Default.S3Region, config.Default.S3Endpoint, config.Default.S3ForcePathStyle)
		if err != nil {
			return nil, fmt.Errorf("failed to create S3 backend: %w", err)
		}
	} else {
		config.StorageBackend = localfs.NewLocalfsBackend(config.Default.MetaPath, config.Default.FilesPath)
		if config.Default.CleanupEvery.Duration > 0 {
			go cleanup.PeriodicCleanup(config.Default.CleanupEvery.Duration, config.Default.FilesPath, config.Default.MetaPath, config.Default.NoLogs)
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
	if config.Default.RealIP {
		r.Use(middleware.RealIP)
	}

	if !config.Default.NoLogs {
		r.Use(middleware.Logger)
	}

	r.Use(middleware.Recoverer)
	r.Use(ContentSecurityPolicy(CSPOptions{
		Policy:         config.Default.ContentSecurityPolicy,
		ReferrerPolicy: config.Default.ReferrerPolicy,
		Frame:          config.Default.XFrameOptions,
	}))
	r.Use(headers.AddHeaders(config.Default.AddHeaders))

	if config.Default.AuthFile != "" {
		r.Use(apikeys.NewAPIKeysMiddleware(apikeys.AuthOptions{
			AuthFile:      config.Default.AuthFile,
			UnauthMethods: []string{http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace},
			BasicAuth:     config.Default.BasicAuth,
			SiteName:      config.Default.SiteName,
			SitePath:      config.Default.SiteURL.Path,
		}))
	}

	if config.Default.AuthFile == "" || config.Default.BasicAuth {
		r.Get("/", handlers.Index)
		r.Get("/paste", handlers.Paste)
	} else {
		r.Get("/", http.RedirectHandler(path.Join(config.Default.SiteURL.Path, "API"), http.StatusSeeOther).ServeHTTP)
		r.Get("/paste", http.RedirectHandler(path.Join(config.Default.SiteURL.Path, "API/"), http.StatusSeeOther).ServeHTTP)
	}

	r.Get("/api", handlers.APIDoc)
	r.Get("/API", http.RedirectHandler(path.Join(config.Default.SiteURL.Path, "api"), http.StatusPermanentRedirect).ServeHTTP)

	if config.Default.RemoteUploads {
		r.Get("/upload", upload.Remote)

		if config.Default.RemoteAuthFile != "" {
			config.RemoteAuthKeys = apikeys.ReadAuthKeys(config.Default.RemoteAuthFile)
		}
	}

	r.Post("/upload", upload.POSTHandler)
	r.Put("/upload", upload.PUTHandler)
	r.Put("/upload/{name}", upload.PUTHandler)

	r.Delete("/{name}", handlers.Delete)

	r.Get("/{name}", handlers.FileAccessHandler)
	r.Post("/{name}", handlers.FileAccessHandler)
	r.Get(path.Join("/", config.Default.SelifPath, "{name}"), handlers.FileServeHandler)
	r.Get("/torrent/{name}", torrent.FileTorrentHandler)
	r.Get("/{name}/torrent", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/torrent/"+chi.URLParam(r, "name"), http.StatusMovedPermanently)
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
