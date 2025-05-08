## linx-server

Self-hosted file/media sharing website

```
linx-server [flags]
```

### Options

```
      --allow-hotlink                 Allow hot-linking of files
      --auth-basic                    Allow logging in with basic auth password
      --auth-cookie-expiry duration   Expiration time for access key cookies in seconds (set 0 to use session cookies)
      --auth-file string              Path to a file containing newline-separated scrypted auth keys
      --auth-remote-file string       Path to a file containing newline-separated scrypted auth keys for remote uploads
      --bind string                   Host to bind to (default "127.0.0.1:8080")
      --cleanup-every duration        How often to clean up expired files. A value of 0 means files will be cleaned up as they are accessed. (default 1h0m0s)
  -c, --config string                 Path to the config file (default "$HOME/.config/linx-server/config.toml")
      --custom-pages-path string      Path to directory containing .md files to render as custom pages
      --files-path string             Path to files directory (default "data/files")
      --force-random-filename         Force all uploads to use a random filename (default true)
      --graceful-shutdown duration    Maximum time to wait for requests to finish during shutdown (default 30s)
  -h, --help                          help for linx-server
      --max-expiry duration           Maximum expiration time. A value of 0 means no expiry.
      --max-size string               Maximum upload file size (default "4 GiB")
      --meta-path string              Path to metadata directory (default "data/meta")
      --no-direct-agents              Disable serving files directly for wget/curl user agents
      --no-logs                       Remove logging of each request
      --real-ip                       Use X-Real-IP/X-Forwarded-For headers
      --remote-uploads                Enable remote uploads (/upload?url=https://...)
      --s3-bucket string              S3 bucket to use for files and metadata
      --s3-endpoint string            S3 endpoint
      --s3-force-path-style           Force path-style addressing for S3 (e.g. https://s3.amazonaws.com/linx/example.txt)
      --s3-region string              S3 region
      --selif-path string             Path relative to site base url where files are accessed directly (default "selif")
      --site-name string              Name of the site (default "Linx")
      --site-url string               Site base url
      --tls-cert string               Path to ssl certificate (for https)
      --tls-key string                Path to ssl key (for https)
      --upload-max-memory string      Maximum memory to buffer multipart uploads; excess is written to temp files (default "32 MiB")
  -v, --version                       version for linx-server
```

### SEE ALSO

* [linx-server cleanup](linx-server_cleanup.md)	 - Manually clean up expired files
* [linx-server genkey](linx-server_genkey.md)	 - Generate auth file hashed keys
* [linx-server migrate](linx-server_migrate.md)	 - Migrate uploads to a new storage backend

