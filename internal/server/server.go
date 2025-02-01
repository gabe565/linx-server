package server

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andreimarcu/linx-server/assets"
	"github.com/andreimarcu/linx-server/internal/auth/apikeys"
	"github.com/andreimarcu/linx-server/internal/backends/localfs"
	"github.com/andreimarcu/linx-server/internal/backends/s3"
	"github.com/andreimarcu/linx-server/internal/cleanup"
	"github.com/andreimarcu/linx-server/internal/config"
	"github.com/andreimarcu/linx-server/internal/csp"
	"github.com/andreimarcu/linx-server/internal/custompages"
	"github.com/andreimarcu/linx-server/internal/handlers"
	"github.com/andreimarcu/linx-server/internal/headers"
	"github.com/andreimarcu/linx-server/internal/templates"
	"github.com/andreimarcu/linx-server/internal/torrent"
	"github.com/andreimarcu/linx-server/internal/upload"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

func Setup() (*web.Mux, error) {
	mux := web.New()

	// middleware
	mux.Use(middleware.RequestID)

	if config.Default.RealIp {
		mux.Use(middleware.RealIP)
	}

	if !config.Default.NoLogs {
		mux.Use(middleware.Logger)
	}

	mux.Use(middleware.Recoverer)
	mux.Use(middleware.AutomaticOptions)
	mux.Use(csp.ContentSecurityPolicy(csp.CSPOptions{
		Policy:         config.Default.ContentSecurityPolicy,
		ReferrerPolicy: config.Default.ReferrerPolicy,
		Frame:          config.Default.XFrameOptions,
	}))
	mux.Use(headers.AddHeaders(config.Default.AddHeaders))

	if config.Default.AuthFile != "" {
		mux.Use(apikeys.NewApiKeysMiddleware(apikeys.AuthOptions{
			AuthFile:      config.Default.AuthFile,
			UnauthMethods: []string{"GET", "HEAD", "OPTIONS", "TRACE"},
			BasicAuth:     config.Default.BasicAuth,
			SiteName:      config.Default.SiteName,
			SitePath:      config.Default.SitePath,
		}))
	}

	// make directories if needed
	err := os.MkdirAll(config.Default.FilesDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("could not create files directory: %w", err)
	}

	err = os.MkdirAll(config.Default.MetaDir, 0700)
	if err != nil {
		return nil, fmt.Errorf("could not create metadata directory: %w", err)
	}

	if config.Default.SiteURL != "" {
		// ensure siteURL ends wth '/'
		if lastChar := config.Default.SiteURL[len(config.Default.SiteURL)-1:]; lastChar != "/" {
			config.Default.SiteURL = config.Default.SiteURL + "/"
		}

		parsedUrl, err := url.Parse(config.Default.SiteURL)
		if err != nil {
			return nil, fmt.Errorf("could not parse siteurl: %w", err)
		}

		config.Default.SitePath = parsedUrl.Path
	} else {
		config.Default.SitePath = "/"
	}

	config.Default.SelifPath = strings.TrimLeft(config.Default.SelifPath, "/")
	if lastChar := config.Default.SelifPath[len(config.Default.SelifPath)-1:]; lastChar != "/" {
		config.Default.SelifPath = config.Default.SelifPath + "/"
	}

	if config.Default.S3Bucket != "" {
		config.StorageBackend = s3.NewS3Backend(config.Default.S3Bucket, config.Default.S3Region, config.Default.S3Endpoint, config.Default.S3ForcePathStyle)
	} else {
		config.StorageBackend = localfs.NewLocalfsBackend(config.Default.MetaDir, config.Default.FilesDir)
		if config.Default.CleanupEveryMinutes > 0 {
			go cleanup.PeriodicCleanup(time.Duration(config.Default.CleanupEveryMinutes)*time.Minute, config.Default.FilesDir, config.Default.MetaDir, config.Default.NoLogs)
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
	nameRe := regexp.MustCompile("^" + config.Default.SitePath + `(?P<name>[a-z0-9-\.]+)$`)
	selifRe := regexp.MustCompile("^" + config.Default.SitePath + config.Default.SelifPath + `(?P<name>[a-z0-9-\.]+)$`)
	selifIndexRe := regexp.MustCompile("^" + config.Default.SitePath + config.Default.SelifPath + `$`)
	torrentRe := regexp.MustCompile("^" + config.Default.SitePath + `(?P<name>[a-z0-9-\.]+)/torrent$`)

	if config.Default.AuthFile == "" || config.Default.BasicAuth {
		mux.Get(config.Default.SitePath, handlers.Index)
		mux.Get(config.Default.SitePath+"paste/", handlers.Paste)
	} else {
		mux.Get(config.Default.SitePath, http.RedirectHandler(config.Default.SitePath+"API", 303))
		mux.Get(config.Default.SitePath+"paste/", http.RedirectHandler(config.Default.SitePath+"API/", 303))
	}
	mux.Get(config.Default.SitePath+"paste", http.RedirectHandler(config.Default.SitePath+"paste/", 301))

	mux.Get(config.Default.SitePath+"API/", handlers.APIDoc)
	mux.Get(config.Default.SitePath+"API", http.RedirectHandler(config.Default.SitePath+"API/", 301))

	if config.Default.RemoteUploads {
		mux.Get(config.Default.SitePath+"upload", upload.Remote)
		mux.Get(config.Default.SitePath+"upload/", upload.Remote)

		if config.Default.RemoteAuthFile != "" {
			config.RemoteAuthKeys = apikeys.ReadAuthKeys(config.Default.RemoteAuthFile)
		}
	}

	mux.Post(config.Default.SitePath+"upload", upload.POSTHandler)
	mux.Post(config.Default.SitePath+"upload/", upload.POSTHandler)
	mux.Put(config.Default.SitePath+"upload", upload.PUTHandler)
	mux.Put(config.Default.SitePath+"upload/", upload.PUTHandler)
	mux.Put(config.Default.SitePath+"upload/:name", upload.PUTHandler)

	mux.Delete(config.Default.SitePath+":name", handlers.Delete)

	mux.Get(config.Default.SitePath+"static/*", handlers.StaticHandler)
	mux.Get(config.Default.SitePath+"favicon.ico", handlers.StaticHandler)
	mux.Get(config.Default.SitePath+"robots.txt", handlers.StaticHandler)
	mux.Get(nameRe, handlers.FileAccessHeader)
	mux.Post(nameRe, handlers.FileAccessHeader)
	mux.Get(selifRe, handlers.FileServeHandler)
	mux.Get(selifIndexRe, handlers.Unauthorized)
	mux.Get(torrentRe, torrent.FileTorrentHandler)

	if config.Default.CustomPagesDir != "" {
		custompages.InitializeCustomPages(config.Default.CustomPagesDir)
		for fileName := range custompages.Names {
			mux.Get(config.Default.SitePath+fileName, handlers.MakeCustomPage(fileName))
			mux.Get(config.Default.SitePath+fileName+"/", handlers.MakeCustomPage(fileName))
		}
	}

	mux.NotFound(handlers.NotFound)

	return mux, nil
}
