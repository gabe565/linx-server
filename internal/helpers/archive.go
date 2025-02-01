package helpers

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"slices"
)

type ReadSeekerAt interface {
	io.Reader
	io.Seeker
	io.ReaderAt
}

func ListArchiveFiles(mimetype string, size int64, r ReadSeekerAt) ([]string, error) {
	var files []string
	defer slices.Sort(files)
	var err error
	switch mimetype {
	case "application/x-tar":
		tReadr := tar.NewReader(r)
		for {
			hdr, err := tReadr.Next()
			if err != nil {
				if err == io.EOF {
					return files, nil
				}
				return files, err
			}
			if hdr.Typeflag == tar.TypeDir || hdr.Typeflag == tar.TypeReg {
				files = append(files, hdr.Name)
			}
		}
	case "application/x-gzip":
		gzf, err := gzip.NewReader(r)
		if err != nil {
			return files, err
		}

		tReadr := tar.NewReader(gzf)
		for {
			hdr, err := tReadr.Next()
			if err != nil {
				if err == io.EOF {
					return files, nil
				}
				return files, err
			}
			if hdr.Typeflag == tar.TypeDir || hdr.Typeflag == tar.TypeReg {
				files = append(files, hdr.Name)
			}
		}
	case "application/x-bzip":
		bzf := bzip2.NewReader(r)
		tReadr := tar.NewReader(bzf)
		for {
			hdr, err := tReadr.Next()
			if err != nil {
				if err == io.EOF {
					return files, nil
				}
				return files, err
			}
			if hdr.Typeflag == tar.TypeDir || hdr.Typeflag == tar.TypeReg {
				files = append(files, hdr.Name)
			}
		}
	case "application/zip":
		zf, err := zip.NewReader(r, size)
		if err != nil {
			return files, err
		}
		files = slices.Grow(files, len(zf.File))
		for _, f := range zf.File {
			files = append(files, f.Name)
		}
	}

	return files, err
}
