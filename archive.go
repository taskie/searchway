package srchway

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
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

func UnarchiveTarItem(tarReader *tar.Reader, outFilePath string) (itemPath string, err error) {
	header, err := tarReader.Next()
	if err != nil {
		return
	}

	itemPath = filepath.Join(outFilePath, header.Name)
	info := header.FileInfo()

	if info.IsDir() {
		err = os.MkdirAll(itemPath, info.Mode())
		return
	}

	parentDir := path.Dir(itemPath)
	_ = os.MkdirAll(parentDir, 0755)

	file, err := os.OpenFile(itemPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return
	}
	defer file.Close()
	_, err = io.Copy(file, tarReader)
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
		itemPath, e := UnarchiveTarItem(tarReader, outFilePath)
		if e == io.EOF {
			break
		} else if e != nil {
			fmt.Fprintln(os.Stderr, itemPath)
			fmt.Fprintln(os.Stderr, e)
			err = e
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
