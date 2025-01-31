package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/andreimarcu/linx-server/internal/auth/apikeys"
	"github.com/andreimarcu/linx-server/internal/backends/localfs"
	"github.com/andreimarcu/linx-server/internal/backends/s3"
	"github.com/andreimarcu/linx-server/internal/cleanup"
	"github.com/andreimarcu/linx-server/internal/config"
	"github.com/andreimarcu/linx-server/internal/csp"
	"github.com/andreimarcu/linx-server/internal/custompages"
	"github.com/andreimarcu/linx-server/internal/handlers"
	"github.com/andreimarcu/linx-server/internal/headers"
	"github.com/andreimarcu/linx-server/internal/torrent"
	"github.com/andreimarcu/linx-server/internal/upload"
	"github.com/flosch/pongo2"
	"github.com/vharitonsky/iniflags"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

func setup() *web.Mux {
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
		log.Fatal("Could not create files directory:", err)
	}

	err = os.MkdirAll(config.Default.MetaDir, 0700)
	if err != nil {
		log.Fatal("Could not create metadata directory:", err)
	}

	if config.Default.SiteURL != "" {
		// ensure siteURL ends wth '/'
		if lastChar := config.Default.SiteURL[len(config.Default.SiteURL)-1:]; lastChar != "/" {
			config.Default.SiteURL = config.Default.SiteURL + "/"
		}

		parsedUrl, err := url.Parse(config.Default.SiteURL)
		if err != nil {
			log.Fatal("Could not parse siteurl:", err)
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
	p2l, err := NewPongo2TemplatesLoader()
	if err != nil {
		log.Fatal("Error: could not load templates", err)
	}
	TemplateSet := pongo2.NewSet("assets/templates", p2l)
	err = PopulateTemplatesMap(TemplateSet, config.Templates)
	if err != nil {
		log.Fatal("Error: could not load templates", err)
	}

	config.StaticBox = rice.MustFindBox("assets/static")
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

	return mux
}

func main() {
	flag.StringVar(&config.Default.Bind, "bind", "127.0.0.1:8080",
		"host to bind to (default: 127.0.0.1:8080)")
	flag.StringVar(&config.Default.FilesDir, "filespath", "files/",
		"path to files directory")
	flag.StringVar(&config.Default.MetaDir, "metapath", "meta/",
		"path to metadata directory")
	flag.BoolVar(&config.Default.BasicAuth, "basicauth", false,
		"allow logging by basic auth password")
	flag.BoolVar(&config.Default.NoLogs, "nologs", false,
		"remove stdout output for each request")
	flag.BoolVar(&config.Default.AllowHotlink, "allowhotlink", false,
		"Allow hotlinking of files")
	flag.StringVar(&config.Default.SiteName, "sitename", "",
		"name of the site")
	flag.StringVar(&config.Default.SiteURL, "siteurl", "",
		"site base url (including trailing slash)")
	flag.StringVar(&config.Default.SelifPath, "selifpath", "selif",
		"path relative to site base url where files are accessed directly")
	flag.Int64Var(&config.Default.MaxSize, "maxsize", 4*1024*1024*1024,
		"maximum upload file size in bytes (default 4GB)")
	flag.Uint64Var(&config.Default.MaxExpiry, "maxexpiry", 0,
		"maximum expiration time in seconds (default is 0, which is no expiry)")
	flag.StringVar(&config.Default.CertFile, "certfile", "",
		"path to ssl certificate (for https)")
	flag.StringVar(&config.Default.KeyFile, "keyfile", "",
		"path to ssl key (for https)")
	flag.BoolVar(&config.Default.RealIp, "realip", false,
		"use X-Real-IP/X-Forwarded-For headers as original host")
	flag.BoolVar(&config.Default.Fastcgi, "fastcgi", false,
		"serve through fastcgi")
	flag.BoolVar(&config.Default.RemoteUploads, "remoteuploads", false,
		"enable remote uploads")
	flag.StringVar(&config.Default.AuthFile, "authfile", "",
		"path to a file containing newline-separated scrypted auth keys")
	flag.StringVar(&config.Default.RemoteAuthFile, "remoteauthfile", "",
		"path to a file containing newline-separated scrypted auth keys for remote uploads")
	flag.StringVar(&config.Default.ContentSecurityPolicy, "contentsecuritypolicy",
		"default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'self';",
		"value of default Content-Security-Policy header")
	flag.StringVar(&config.Default.FileContentSecurityPolicy, "filecontentsecuritypolicy",
		"default-src 'none'; img-src 'self'; object-src 'self'; media-src 'self'; style-src 'self' 'unsafe-inline'; frame-ancestors 'self';",
		"value of Content-Security-Policy header for file access")
	flag.StringVar(&config.Default.ReferrerPolicy, "referrerpolicy",
		"same-origin",
		"value of default Referrer-Policy header")
	flag.StringVar(&config.Default.FileReferrerPolicy, "filereferrerpolicy",
		"same-origin",
		"value of Referrer-Policy header for file access")
	flag.StringVar(&config.Default.XFrameOptions, "xframeoptions", "SAMEORIGIN",
		"value of X-Frame-Options header")
	flag.Var(&config.Default.AddHeaders, "addheader",
		"Add an arbitrary header to the response. This option can be used multiple times.")
	flag.BoolVar(&config.Default.NoDirectAgents, "nodirectagents", false,
		"disable serving files directly for wget/curl user agents")
	flag.StringVar(&config.Default.S3Endpoint, "s3-endpoint", "",
		"S3 endpoint")
	flag.StringVar(&config.Default.S3Region, "s3-region", "",
		"S3 region")
	flag.StringVar(&config.Default.S3Bucket, "s3-bucket", "",
		"S3 bucket to use for files and metadata")
	flag.BoolVar(&config.Default.S3ForcePathStyle, "s3-force-path-style", false,
		"Force path-style addressing for S3 (e.g. https://s3.amazonaws.com/linx/example.txt)")
	flag.BoolVar(&config.Default.ForceRandomFilename, "force-random-filename", false,
		"Force all uploads to use a random filename")
	flag.Uint64Var(&config.Default.AccessKeyCookieExpiry, "access-cookie-expiry", 0, "Expiration time for access key cookies in seconds (set 0 to use session cookies)")
	flag.StringVar(&config.Default.CustomPagesDir, "custompagespath", "",
		"path to directory containing .md files to render as custom pages")
	flag.Uint64Var(&config.Default.CleanupEveryMinutes, "cleanup-every-minutes", 0,
		"How often to clean up expired files in minutes (default is 0, which means files will be cleaned up as they are accessed)")

	iniflags.Parse()

	mux := setup()

	if config.Default.Fastcgi {
		var listener net.Listener
		var err error
		if config.Default.Bind[0] == '/' {
			// UNIX path
			listener, err = net.ListenUnix("unix", &net.UnixAddr{Name: config.Default.Bind, Net: "unix"})
			cleanup := func() {
				log.Print("Removing FastCGI socket")
				os.Remove(config.Default.Bind)
			}
			defer cleanup()
			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				sig := <-sigs
				log.Print("Signal: ", sig)
				cleanup()
				os.Exit(0)
			}()
		} else {
			listener, err = net.Listen("tcp", config.Default.Bind)
		}
		if err != nil {
			log.Fatal("Could not bind: ", err)
		}

		log.Printf("Serving over fastcgi, bound on %s", config.Default.Bind)
		fcgi.Serve(listener, mux)
	} else if config.Default.CertFile != "" {
		log.Printf("Serving over https, bound on %s", config.Default.Bind)
		err := graceful.ListenAndServeTLS(config.Default.Bind, config.Default.CertFile, config.Default.KeyFile, mux)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("Serving over http, bound on %s", config.Default.Bind)
		err := graceful.ListenAndServe(config.Default.Bind, mux)
		if err != nil {
			log.Fatal(err)
		}
	}
}
