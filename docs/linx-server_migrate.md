## linx-server migrate

Migrate uploads to a new storage backend

```
linx-server migrate [flags]
```

### Options

```
      --concurrency int       Number of uploads to migrate in parallel (default 4)
  -c, --config string         Path to the config file (default "$HOME/.config/linx-server/config.toml")
      --files-path string     Path to files directory (default "data/files")
  -f, --from string           Source backend (one of s3, local)
  -h, --help                  help for migrate
      --meta-path string      Path to metadata directory (default "data/meta")
      --no-logs               Disable logging of migrated files
      --s3-bucket string      S3 bucket to use for files and metadata
      --s3-endpoint string    S3 endpoint
      --s3-force-path-style   Force path-style addressing for S3 (e.g. https://s3.amazonaws.com/linx/example.txt)
      --s3-region string      S3 region
  -t, --to string             Destination backend (one of s3, local)
```

### SEE ALSO

* [linx-server](linx-server.md)	 - Self-hosted file/media sharing website

