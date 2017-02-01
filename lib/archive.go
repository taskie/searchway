package srchway

import (
	"io"
	"os"
	"path/filepath"
	"compress/gzip"
	"archive/tar"
	"strings"
)

func ExtractGz(inFilePath string) (outFilePath string, err error) {
	inFile, err := os.Open(inFilePath)
	if err != nil {
		return
	}
	defer inFile.Close()

	outFilePath = strings.TrimSuffix(inFilePath, ".gz")
	outFile, err := os.Create(outFilePath)
	if err != nil {
		return
	}
	defer outFile.Close()

	reader, err := gzip.NewReader(inFile)
	if err != nil {
		return
	}
	defer reader.Close()

	_, err = io.Copy(outFile, reader)
	return
}

func ExtractAndRemoveGz(inFilePath string) (outFilePath string, err error) {
	outFilePath, err = ExtractGz(inFilePath)
	if err != nil {
		return
	}
	err = os.Remove(inFilePath)
	return
}

func UnarchiveTar(inFilePath string) (outFilePath string, err error) {
	// http://blog.ralch.com/tutorial/golang-working-with-tar-and-gzip/
	reader, err := os.Open(inFilePath)
	if err != nil {
		return
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	outFilePath = strings.TrimSuffix(inFilePath, ".tar")

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return outFilePath, err
		}

		path := filepath.Join(outFilePath, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return outFilePath, err
			}
			continue
		}

		file, err:= os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return outFilePath, err
		}
		defer file.Close()
		_, err =io.Copy(file, tarReader)
		if err != nil {
			return outFilePath, err
		}
	}
	return
}

func UnarchiveAndRemoveTar(inFilePath string) (outFilePath string, err error) {
	outFilePath, err = UnarchiveTar(inFilePath)
	if err != nil {
		return
	}
	err = os.Remove(inFilePath)
	return
}

func ExtractAndRemoveTarGz(inFilePath string) (outFilePath string, err error) {
	outFilePath, err = ExtractAndRemoveGz(inFilePath)
	if err != nil {
		return
	}
	outFilePath, err = UnarchiveAndRemoveTar(outFilePath)
	return
}