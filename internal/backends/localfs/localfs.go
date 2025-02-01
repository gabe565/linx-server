package localfs

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"time"

	"gabe565.com/linx-server/internal/backends"
	"gabe565.com/linx-server/internal/helpers"
)

type LocalfsBackend struct {
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

func (b LocalfsBackend) Delete(key string) (err error) {
	err = os.Remove(path.Join(b.filesPath, key))
	if err != nil {
		return
	}
	err = os.Remove(path.Join(b.metaPath, key))
	return
}

func (b LocalfsBackend) Exists(key string) (bool, error) {
	_, err := os.Stat(path.Join(b.filesPath, key))
	return err == nil, err
}

func (b LocalfsBackend) Head(key string) (metadata backends.Metadata, err error) {
	f, err := os.Open(path.Join(b.metaPath, key))
	if os.IsNotExist(err) {
		return metadata, backends.NotFoundErr
	} else if err != nil {
		return metadata, backends.BadMetadata
	}
	defer f.Close()

	decoder := json.NewDecoder(f)

	mjson := MetadataJSON{}
	if err := decoder.Decode(&mjson); err != nil {
		return metadata, backends.BadMetadata
	}

	metadata.DeleteKey = mjson.DeleteKey
	metadata.AccessKey = mjson.AccessKey
	metadata.Mimetype = mjson.Mimetype
	metadata.ArchiveFiles = mjson.ArchiveFiles
	metadata.Sha256sum = mjson.Sha256sum
	metadata.Expiry = time.Unix(mjson.Expiry, 0)
	metadata.Size = mjson.Size

	return
}

func (b LocalfsBackend) Get(key string) (metadata backends.Metadata, f io.ReadCloser, err error) {
	metadata, err = b.Head(key)
	if err != nil {
		return
	}

	f, err = os.Open(path.Join(b.filesPath, key))
	if err != nil {
		return
	}

	return
}

func (b LocalfsBackend) ServeFile(key string, w http.ResponseWriter, r *http.Request) (err error) {
	_, err = b.Head(key)
	if err != nil {
		return
	}

	filePath := path.Join(b.filesPath, key)
	http.ServeFile(w, r, filePath)

	return
}

func (b LocalfsBackend) writeMetadata(key string, metadata backends.Metadata) error {
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

func (b LocalfsBackend) Put(key string, r io.Reader, expiry time.Time, deleteKey, accessKey string) (backends.Metadata, error) {
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
		return m, backends.FileEmptyError
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

func (b LocalfsBackend) PutMetadata(key string, m backends.Metadata) error {
	return b.writeMetadata(key, m)
}

func (b LocalfsBackend) Size(key string) (int64, error) {
	fileInfo, err := os.Stat(path.Join(b.filesPath, key))
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}

func (b LocalfsBackend) List() ([]string, error) {
	var output []string

	files, err := ioutil.ReadDir(b.filesPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		output = append(output, file.Name())
	}

	return output, nil
}

func NewLocalfsBackend(metaPath string, filesPath string) LocalfsBackend {
	return LocalfsBackend{
		metaPath:  metaPath,
		filesPath: filesPath,
	}
}
