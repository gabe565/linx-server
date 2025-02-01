package server

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
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
	err := os.MkdirAll(config.Default.FilesDir, 0o755)
	if err != nil {
		return nil, fmt.Errorf("could not create files directory: %w", err)
	}

	err = os.MkdirAll(config.Default.MetaDir, 0o700)
	if err != nil {
		return nil, fmt.Errorf("could not create metadata directory: %w", err)
	}

	if config.Default.SiteURL != "" {
		config.Default.SiteURL = strings.TrimSuffix(config.Default.SiteURL, "/") + "/"
	}
	parsedURL, err := url.Parse(config.Default.SiteURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse siteurl: %w", err)
	}
	config.Default.SitePath = "/" + strings.Trim(parsedURL.Path, "/")
	config.Default.SelifPath = strings.Trim(config.Default.SelifPath, "/")

	if config.Default.S3Bucket != "" {
		config.StorageBackend = s3.NewS3Backend(config.Default.S3Bucket, config.Default.S3Region, config.Default.S3Endpoint, config.Default.S3ForcePathStyle)
	} else {
		config.StorageBackend = localfs.NewLocalfsBackend(config.Default.MetaDir, config.Default.FilesDir)
		if config.Default.CleanupEvery.Duration > 0 {
			go cleanup.PeriodicCleanup(config.Default.CleanupEvery.Duration, config.Default.FilesDir, config.Default.MetaDir, config.Default.NoLogs)
		}
	}

	// Template setup
	config.Templates, err = templates.Load(assets.Template)
	if err != nil {
		return nil, fmt.Errorf("could not load templates: %w", err)
	}

	config.TimeStarted = time.Now()
	config.TimeStartedStr = strconv.FormatInt(config.TimeStarted.Unix(), 10)

	// Routing setup
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RedirectSlashes)

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
			SitePath:      config.Default.SitePath,
		}))
	}

	r.Route(config.Default.SitePath, func(r chi.Router) {
		if config.Default.AuthFile == "" || config.Default.BasicAuth {
			r.Get("/", handlers.Index)
			r.Get("/paste", handlers.Paste)
		} else {
			r.Get("/", http.RedirectHandler(config.Default.SitePath+"API", http.StatusSeeOther).ServeHTTP)
			r.Get("/paste", http.RedirectHandler(config.Default.SitePath+"API/", http.StatusSeeOther).ServeHTTP)
		}

		r.Get("/API", handlers.APIDoc)

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

		r.Get("/static/*", handlers.StaticHandler)
		r.Get("/favicon.ico", handlers.StaticHandler)
		r.Get("/robots.txt", handlers.StaticHandler)
		r.Get("/{name}", handlers.FileAccessHeader)
		r.Post("/{name}", handlers.FileAccessHeader)
		r.Route("/"+config.Default.SelifPath, func(r chi.Router) {
			r.Get("/{name}", handlers.FileServeHandler)
		})
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

		r.NotFound(handlers.NotFound)
	})

	return r, nil
}
