# linx-server

[![Build](https://github.com/gabe565/linx-server/actions/workflows/build.yaml/badge.svg)](https://github.com/gabe565/linx-server/actions/workflows/build.yaml)

Self-hosted file/media sharing website.

### Clients
**Official**
- CLI: **linx-client** - [Source](https://github.com/andreimarcu/linx-client)

**Unofficial**
- Android: **LinxShare** - [Source](https://github.com/iksteen/LinxShare/) | [Google Play](https://play.google.com/store/apps/details?id=org.thegraveyard.linxshare)
- CLI: **golinx** - [Source](https://github.com/mutantmonkey/golinx)


### Features
- Display common filetypes (image, video, audio, markdown, pdf)  
- Display syntax-highlighted code with in-place editing
- Documented API with keys for restricting uploads
- Torrent download of files using web seeding
- File expiry, deletion key, file access key, and random filename options


### Screenshots
<img width="730" src="https://github.com/user-attachments/assets/80d826d0-b99b-44a5-8bb0-da6597347fcd#gh-dark-mode-only" />
<img width="730" src="https://github.com/user-attachments/assets/2db9679f-dce0-42fb-b2a3-944bfa6d9fbc#gh-light-mode-only" />

<img width="180" src="https://github.com/user-attachments/assets/b20183e6-b726-4bb3-9640-271b2aa101ba#gh-dark-mode-only" /> <img width="180" src="https://github.com/user-attachments/assets/1560d327-b70f-48cb-8cdb-d1dd426547c7#gh-dark-mode-only" /> <img width="180" src="https://github.com/user-attachments/assets/e7f79bcd-acd0-4f64-88ef-d2065f7cdcc6#gh-dark-mode-only" /> <img width="180" src="https://github.com/user-attachments/assets/fe74dda3-216e-4a4d-8fb1-44f1af2bd34d#gh-dark-mode-only" />
<img width="180" src="https://github.com/user-attachments/assets/4623a8df-459f-4934-a18a-32e1b39eb9a3#gh-light-mode-only" /> <img width="180" src="https://github.com/user-attachments/assets/03c6efcc-b0cd-4dd9-bafc-61fe52d160d5#gh-light-mode-only" /> <img width="180" src="https://github.com/user-attachments/assets/0796c178-b5c9-4666-b46f-b29f88c5ea73#gh-light-mode-only" /> <img width="180" src="https://github.com/user-attachments/assets/91b3b9de-3374-4290-be83-8cafc9fc3c13#gh-light-mode-only" />


## Getting started

### Using Docker
1. Create `data` directory and run `chown -R 65534:65534 data`
2. Optionally, create a config file ([example](config_example.toml)), we'll refer to it as `config.toml` in the following examples

Example running
```shell
docker run \
  -p 8080:8080 \
  -v /path/to/config.toml:/data/config.toml \
  -v /path/to/data:/data \
  ghcr.io/gabe565/linx-server
```

Example with Docker Compose:
```yaml
services:
  linx-server:
    container_name: linx-server
    image: ghcr.io/gabe565/linx-server
    command: --config=/data/config.toml
    volumes:
      - /path/to/data:/data
      - /path/to/config.toml:/data/config.toml
    ports:
      - "8080:8080"
    restart: unless-stopped
```
Ideally, you would use a reverse proxy such as nginx or caddy to handle TLS certificates.

### Using a binary release

1. Grab the latest binary from the [releases](https://github.com/gabe565/linx-server/releases)
2. Run `linx-server --config=path/to/config.toml`


## Usage

### Configuration
All configuration options are accepted either as arguments or can be placed in a file as such (see [example](config_example.toml)):
```toml
bind = '127.0.0.1:8080'
site-name = 'myLinx'
max-size = '4 MiB'
max-expiry = '24h'
# ... etc
```
...and then run `linx-server --config=path/to/config.toml`

### Options
See the [example configuration file](config_example.toml) or the [command-line docs](docs/linx-server.md).

Any config can be provided as an environment variable by capitalizing it, changing `-` to `_`, and prefixing it with `LINX_`.

## Deployment
Linx-server supports being deployed in a subdirectory (ie. example.com/mylinx/) as well as on its own (example.com/).


### 1. Using the built-in http server
Run linx-server normally.

### 2. Using the built-in https server
Run linx-server with the `cert-file = path/to/cert.file` and `key-file = path/to/key.file` options.

### 3. Using fastcgi

A suggested deployment is running nginx in front of linx-server serving through fastcgi.
This allows you to have nginx handle the TLS termination for example.  
An example configuration:
```
server {
    ...
    server_name yourlinx.example.org;
    ...
    
    client_max_body_size 4096M;
    location / {
        fastcgi_pass 127.0.0.1:8080;
        include fastcgi_params;
    }
}
```
And run linx-server with the `fastcgi = true` option.

## Author
- Andrei Marcu, https://andreim.net
- Gabe Cook, https://gabecook.com
