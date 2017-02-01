package srchway

import (
	"strings"
	"os"
	"path"
	"errors"
)

type PrintMode int

const (
	NormalMode PrintMode = iota
	JsonMode
)

type Repo interface {
	Search(query string) (bytes []byte, err error)
	Info(query string) (bytes []byte, err error)
	Get(query string, outFilePath string) (err error)
	PrintSearchResponse(query string, mode PrintMode) (err error)
	PrintInfoResponse(query string, mode PrintMode) (err error)
}

func joinOrNoneString(ss []string) (s string) {
	if len(ss) == 0 {
		s = "None"
	} else {
		s = strings.Join(ss, " ")
	}
	return
}

func joinOrNoneStringForOptDepends(ss []string) (s string) {
	if len(ss) == 0 {
		s = "None"
	} else {
		s = strings.Join(ss, "\n                  ")
	}
	return
}

func createOutFile(outFilePath string, url string) (file *os.File, err error) {
	if outFilePath == "" {
		outFilePath = path.Base(url)
	}
	fi, err := os.Stat(outFilePath);
	switch {
	case err != nil:
		// how to handle err? (through)
		file, err = os.Create(outFilePath)
		return
	case fi.IsDir():
		outFilePath2 := outFilePath + "/" + path.Base(url)
		file, err = createOutFile(outFilePath2, url)
		return
	default:
		err = errors.New("file exists")
		return
	}
}
