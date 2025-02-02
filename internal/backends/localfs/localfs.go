package localfs

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/helpers"
)

type Backend struct {
	metaPath  string
	filesPath string
}

type MetadataJSON struct {
	DeleteKey    string   `json:"delete_key"`
	AccessKey    string   `json:"access_key,omitempty"`
	Sha256sum    string   `json:"sha256sum"`
	Mimetype     string   `json:"mimetype"`
	Size         int64    `json:"size"`
	Expiry       int64    `json:"expiry"`
	ArchiveFiles []string `json:"archive_files,omitempty"`
}

func (b Backend) Delete(key string) error {
	return errors.Join(
		os.Remove(path.Join(b.filesPath, key)),
		os.Remove(path.Join(b.metaPath, key)),
	)
}

func (b Backend) Exists(key string) (bool, error) {
	_, err := os.Stat(path.Join(b.filesPath, key))
	exists := true
	if err != nil && os.IsNotExist(err) {
		exists = false
		err = nil
	}
	return exists, err
}

func (b Backend) Head(key string) (backends.Metadata, error) {
	var metadata backends.Metadata
	f, err := os.Open(path.Join(b.metaPath, key))
	if os.IsNotExist(err) {
		return metadata, backends.ErrNotFound
	} else if err != nil {
		return metadata, backends.ErrBadMetadata
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
	metadata.Expiry = time.Unix(mjson.Expiry, 0)
	metadata.Size = mjson.Size

	return metadata, nil
}

func (b Backend) Get(key string) (backends.Metadata, io.ReadCloser, error) {
	metadata, err := b.Head(key)
	if err != nil {
		return metadata, nil, err
	}

	f, err := os.Open(path.Join(b.filesPath, key))
	return metadata, f, err
}

func (b Backend) ServeFile(key string, w http.ResponseWriter, r *http.Request) error {
	if _, err := b.Head(key); err != nil {
		return err
	}

	filePath := path.Join(b.filesPath, key)
	http.ServeFile(w, r, filePath)

	return nil
}

func (b Backend) writeMetadata(key string, metadata backends.Metadata) error {
	tmpPath := path.Join(b.metaPath, "."+key)

	mjson := MetadataJSON{
		DeleteKey:    metadata.DeleteKey,
		AccessKey:    metadata.AccessKey,
		Mimetype:     metadata.Mimetype,
		ArchiveFiles: metadata.ArchiveFiles,
		Sha256sum:    metadata.Sha256sum,
		Expiry:       metadata.Expiry.Unix(),
		Size:         metadata.Size,
	}

	tmp, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	if err = json.NewEncoder(tmp).Encode(mjson); err != nil {
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, path.Join(b.metaPath, key))
}

func (b Backend) Put(key string, r io.Reader, expiry time.Time, deleteKey, accessKey string) (backends.Metadata, error) {
	var m backends.Metadata
	tmpPath := path.Join(b.filesPath, "."+key)

	tmp, err := os.Create(tmpPath)
	if err != nil {
		return m, err
	}
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	bytes, err := io.Copy(tmp, r)
	if err != nil {
		return m, err
	} else if bytes == 0 {
		return m, backends.ErrFileEmpty
	}

	if _, err := tmp.Seek(0, 0); err != nil {
		return m, err
	}
	m, err = helpers.GenerateMetadata(tmp)
	if err != nil {
		return m, err
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		return m, err
	}

	m.Expiry = expiry
	m.DeleteKey = deleteKey
	m.AccessKey = accessKey
	m.ArchiveFiles, _ = helpers.ListArchiveFiles(m.Mimetype, m.Size, tmp)

	err = b.writeMetadata(key, m)
	if err != nil {
		return m, err
	}

	if err := tmp.Close(); err != nil {
		return m, err
	}

	return m, os.Rename(tmpPath, path.Join(b.filesPath, key))
}

func (b Backend) PutMetadata(key string, m backends.Metadata) error {
	return b.writeMetadata(key, m)
}

func (b Backend) Size(key string) (int64, error) {
	fileInfo, err := os.Stat(path.Join(b.filesPath, key))
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

func (b Backend) List() ([]string, error) {
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
