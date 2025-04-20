package localfs

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/helpers"
)

type Backend struct {
	metaPath  string
	filesPath string
}

type MetadataJSON struct {
	DeleteKey    string          `json:"delete_key"`
	AccessKey    string          `json:"access_key,omitzero"`
	Sha256sum    string          `json:"sha256sum"`
	Mimetype     string          `json:"mimetype"`
	Size         int64           `json:"size"`
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

	metaErr := metaRoot.Remove(key + ".json")
	if metaErr != nil {
		if errOldPath := metaRoot.Remove(key); errOldPath == nil {
			metaErr = nil
		}
	}

	filesRoot, err := os.OpenRoot(b.filesPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = filesRoot.Close()
	}()

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

	exists := true
	if _, err = filesRoot.Stat(key); err != nil && os.IsNotExist(err) {
		exists = false
		err = nil
	}
	return exists, err
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

	decoder := json.NewDecoder(f)

	mjson := MetadataJSON{}
	if err := decoder.Decode(&mjson); err != nil {
		return metadata, backends.ErrBadMetadata
	}

	metadata.DeleteKey = mjson.DeleteKey
	metadata.AccessKey = mjson.AccessKey
	metadata.Mimetype = mjson.Mimetype
	metadata.ArchiveFiles = mjson.ArchiveFiles
	metadata.Sha256sum = mjson.Sha256sum
	metadata.Expiry = time.Time(mjson.Expiry)
	metadata.Size = mjson.Size

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
		DeleteKey:    metadata.DeleteKey,
		AccessKey:    metadata.AccessKey,
		Mimetype:     metadata.Mimetype,
		ArchiveFiles: metadata.ArchiveFiles,
		Sha256sum:    metadata.Sha256sum,
		Expiry:       backends.Expiry(metadata.Expiry),
		Size:         metadata.Size,
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
	key string,
	r io.Reader,
	expiry time.Time,
	deleteKey, accessKey string,
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

	bytes, err := io.Copy(f, r)
	if err != nil {
		return m, err
	} else if bytes == 0 {
		return m, backends.ErrFileEmpty
	}

	if _, err := f.Seek(0, 0); err != nil {
		return m, err
	}
	m, err = helpers.GenerateMetadata(f)
	if err != nil {
		return m, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return m, err
	}

	m.Expiry = expiry
	m.DeleteKey = deleteKey
	m.AccessKey = accessKey
	m.ArchiveFiles, _ = helpers.ListArchiveFiles(m.Mimetype, m.Size, f)

	err = b.writeMetadata(key, m)
	if err != nil {
		return m, err
	}

	if err := f.Close(); err != nil {
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

func (b Backend) List(_ context.Context) ([]string, error) {
	files, err := os.ReadDir(b.filesPath)
	if err != nil {
		return nil, err
	}

	output := make([]string, 0, len(files))
	for _, file := range files {
		output = append(output, file.Name())
	}

	return output, nil
}

func NewLocalfsBackend(metaPath string, filesPath string) Backend {
	return Backend{
		metaPath:  metaPath,
		filesPath: filesPath,
	}
}
