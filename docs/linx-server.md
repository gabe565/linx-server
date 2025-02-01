## linx-server

Self-hosted file/media sharing website

```
linx-server [flags]
```

### Options

```
      --access-key-cookie-expiry uint         Expiration time for access key cookies in seconds (set 0 to use session cookies)
      --add-header string                     Add an arbitrary header to the response. This option can be used multiple times.
      --allow-hotlink                         Allow hot-linking of files
      --auth-file string                      Path to a file containing newline-separated scrypted auth keys
      --basic-auth                            Allow logging in with basic auth password
      --bind string                           Host to bind to (default "127.0.0.1:8080")
      --cleanup-every duration                How often to clean up expired files. A value of 0 means files will be cleaned up as they are accessed.
  -c, --config string                         Path to the config file (default "$HOME/.config/linx-server/config.toml")
      --content-security-policy string        Value of default Content-Security-Policy header (default "default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'self';")
      --custom-pages-path string              Path to directory containing .md files to render as custom pages
      --fastcgi                               Serve through fastcgi
      --file-content-security-policy string   Value of Content-Security-Policy header for file access (default "default-src 'none'; img-src 'self'; object-src 'self'; media-src 'self'; style-src 'self' 'unsafe-inline'; frame-ancestors 'self';")
      --file-referrer-policy string           Value of Referrer-Policy header for file access (default "same-origin")
      --files-dir string                      Path to files directory (default "files")
      --force-random-filename                 Force all uploads to use a random filename
  -h, --help                                  help for linx-server
      --max-expiry duration                   Maximum expiration time. A value of 0 means no expiry.
      --max-size string                       Maximum upload file size in bytes (default "4 GiB")
      --meta-dir string                       Path to metadata directory (default "meta")
      --no-direct-agents                      Disable serving files directly for wget/curl user agents
      --no-logs                               Remove logging of each request
      --real-ip                               Use X-Real-IP/X-Forwarded-For headers
      --referrer-policy string                Value of default Referrer-Policy header (default "same-origin")
      --remote-auth-file string               Path to a file containing newline-separated scrypted auth keys for remote uploads
      --remote-uploads                        Enable remote uploads
      --s3-bucket string                      S3 bucket to use for files and metadata
      --s3-endpoint string                    S3 endpoint
      --s3-force-path-style                   Force path-style addressing for S3 (e.g. https://s3.amazonaws.com/linx/example.txt)
      --s3-region string                      S3 region
      --selif-path string                     Path relative to site base url where files are accessed directly (default "selif")
      --site-name string                      Name of the site
      --site-url string                       Site base url
      --tls-cert string                       Path to ssl certificate (for https)
      --tls-key string                        Path to ssl key (for https)
      --x-frame-options string                Value of X-Frame-Options header (default "SAMEORIGIN")
```

### SEE ALSO

* [linx-server cleanup](linx-server_cleanup.md)	 - Manually clean up expired files
* [linx-server genkey](linx-server_genkey.md)	 - Generate auth file hashed keys

