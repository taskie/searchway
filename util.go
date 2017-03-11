package srchway

import (
	"errors"
	"net/url"
	"os"
	"path"
	"strings"
)

const VersionString = "0.2.0a"

type QueryItem struct {
	Key    string
	Values []string
}

func BuildQueryString(qs []QueryItem) (queryString string) {
	ss := make([]string, 0)
	for _, item := range qs {
		escapedValues := make([]string, len(item.Values))
		for j, value := range item.Values {
			escapedValues[j] = url.QueryEscape(value)
		}
		value := strings.Join(escapedValues, "+")
		ss = append(ss, url.QueryEscape(item.Key)+"="+value)
	}
	queryString = strings.Join(ss, "&")
	return
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

func createOutFile(outFilePath string, url string) (file *os.File, newOutFilePath string, err error) {
	if outFilePath == "" {
		outFilePath = path.Base(url)
	}
	fi, err := os.Stat(outFilePath)
	switch {
	case err != nil:
		// how to handle err? (through)
		file, err = os.Create(outFilePath)
		if err == nil {
			newOutFilePath = outFilePath
		}
		return
	case fi.IsDir():
		outFilePath2 := outFilePath + "/" + path.Base(url)
		file, newOutFilePath, err = createOutFile(outFilePath2, url)
		return
	default:
		err = errors.New("file exists")
		return
	}
}
