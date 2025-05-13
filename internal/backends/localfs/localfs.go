package localfs

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"iter"
	"net/http"
	"os"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/helpers"
)

var _ backends.ListBackend = Backend{}

type Backend struct {
	metaPath  string
	filesPath string
}

type MetadataJSON struct {
	OriginalName string          `json:"original_name,omitzero"`
	DeleteKey    string          `json:"delete_key"`
	AccessKey    string          `json:"access_key,omitzero"`
	Sha256sum    string          `json:"sha256sum,omitzero"`
	Checksum     string          `json:"checksum"`
	Mimetype     string          `json:"mimetype"`
	Expiry       backends.Expiry `json:"expiry,omitzero"`
	ArchiveFiles []string        `json:"archive_files,omitzero"`
}

func (b Backend) Delete(_ context.Context, key string) error {
	metaRoot, err := os.OpenRoot(b.metaPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = metaRoot.Close()
	}()

	filesRoot, err := os.OpenRoot(b.filesPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = filesRoot.Close()
	}()

	metaErr := metaRoot.Remove(key + ".json")
	if metaErr != nil {
		if errOldPath := metaRoot.Remove(key); errOldPath == nil {
			metaErr = nil
		}
	}

	return errors.Join(filesRoot.Remove(key), metaErr)
}

func (b Backend) Exists(_ context.Context, key string) (bool, error) {
	filesRoot, err := os.OpenRoot(b.filesPath)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = filesRoot.Close()
	}()

	if _, err := filesRoot.Stat(key); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (b Backend) Head(_ context.Context, key string) (backends.Metadata, error) {
	var metadata backends.Metadata

	metaRoot, err := os.OpenRoot(b.metaPath)
	if err != nil {
		return metadata, err
	}
	defer func() {
		_ = metaRoot.Close()
	}()

	f, err := metaRoot.Open(key + ".json")
	if err != nil {
		if f, err = metaRoot.Open(key); err != nil {
			if os.IsNotExist(err) {
				return metadata, backends.ErrNotFound
			}
			return metadata, backends.ErrBadMetadata
		}
	}
	defer func() {
		_ = f.Close()
	}()

	var mjson MetadataJSON
	if err := json.NewDecoder(f).Decode(&mjson); err != nil {
		return metadata, backends.ErrBadMetadata
	}

	metadata.OriginalName = mjson.OriginalName
	metadata.DeleteKey = mjson.DeleteKey
	metadata.AccessKey = mjson.AccessKey
	metadata.Mimetype = mjson.Mimetype
	metadata.ArchiveFiles = mjson.ArchiveFiles
	metadata.Checksum = mjson.Checksum
	if metadata.Checksum == "" {
		metadata.Checksum = mjson.Sha256sum
	}
	metadata.Expiry = time.Time(mjson.Expiry)

	if stat, err := f.Stat(); err == nil {
		metadata.ModTime = stat.ModTime()
	}

	filesRoot, err := os.OpenRoot(b.filesPath)
	if err != nil {
		return metadata, err
	}
	defer func() {
		_ = filesRoot.Close()
	}()

	fileStat, err := filesRoot.Stat(key)
	if err != nil {
		return metadata, err
	}
	metadata.Size = fileStat.Size()

	return metadata, nil
}

func (b Backend) Get(ctx context.Context, key string) (backends.Metadata, io.ReadCloser, error) {
	metadata, err := b.Head(ctx, key)
	if err != nil {
		return metadata, nil, err
	}

	f, err := os.OpenInRoot(b.filesPath, key)
	return metadata, f, err
}

func (b Backend) ServeFile(key string, w http.ResponseWriter, r *http.Request) error {
	if _, err := b.Head(r.Context(), key); err != nil {
		return err
	}

	filesRoot, err := os.OpenRoot(b.filesPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = filesRoot.Close()
	}()

	http.ServeFileFS(w, r, filesRoot.FS(), key)
	return nil
}

func (b Backend) writeMetadata(key string, metadata backends.Metadata) error {
	mjson := MetadataJSON{
		OriginalName: metadata.OriginalName,
		DeleteKey:    metadata.DeleteKey,
		AccessKey:    metadata.AccessKey,
		Mimetype:     metadata.Mimetype,
		ArchiveFiles: metadata.ArchiveFiles,
		Checksum:     metadata.Checksum,
		Expiry:       backends.Expiry(metadata.Expiry),
	}

	metaRoot, err := os.OpenRoot(b.metaPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = metaRoot.Close()
	}()

	var success bool
	path := key + ".json"
	f, err := metaRoot.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
		if !success {
			_ = metaRoot.Remove(path)
		}
	}()

	if err = json.NewEncoder(f).Encode(mjson); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	success = true
	return nil
}

func (b Backend) Put(
	_ context.Context,
	r io.Reader,
	key string,
	size int64,
	opts backends.PutOptions,
) (backends.Metadata, error) {
	var m backends.Metadata

	filesRoot, err := os.OpenRoot(b.filesPath)
	if err != nil {
		return m, err
	}
	defer func() {
		_ = filesRoot.Close()
	}()

	var success bool
	f, err := filesRoot.Create(key)
	if err != nil {
		return m, err
	}
	defer func() {
		_ = f.Close()
		if !success {
			_ = filesRoot.Remove(key)
		}
	}()

	m, err = helpers.GenerateMetadata(io.TeeReader(r, f))
	if err != nil {
		return m, err
	}

	_, _ = f.Seek(0, io.SeekStart)

	switch {
	case m.Size == 0:
		return m, backends.ErrFileEmpty
	case size > 0 && m.Size != size:
		return m, backends.ErrSizeMismatch
	}

	m.OriginalName = opts.OriginalName
	m.Expiry = opts.Expiry
	m.DeleteKey = opts.DeleteKey
	m.AccessKey = opts.AccessKey
	m.ArchiveFiles, _ = helpers.ListArchiveFiles(m.Mimetype, m.Size, f)

	if err := f.Close(); err != nil {
		return m, err
	}

	err = b.writeMetadata(key, m)
	if err != nil {
		return m, err
	}

	success = true
	return m, nil
}

func (b Backend) PutMetadata(_ context.Context, key string, m backends.Metadata) error {
	return b.writeMetadata(key, m)
}

func (b Backend) Size(_ context.Context, key string) (int64, error) {
	filesRoot, err := os.OpenRoot(b.filesPath)
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = filesRoot.Close()
	}()

	fileInfo, err := filesRoot.Stat(key)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

func (b Backend) List(_ context.Context) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		files, err := os.ReadDir(b.filesPath)
		if err != nil {
			yield("", err)
			return
		}

		for _, file := range files {
			if file.IsDir() {
				continue
			}
			if !yield(file.Name(), nil) {
				return
			}
		}
	}
}

func New(metaPath string, filesPath string) Backend {
	return Backend{
		metaPath:  metaPath,
		filesPath: filesPath,
	}
}
