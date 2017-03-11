package srchway

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/fatih/color"
	shutil "github.com/termie/go-shutil"
)

const UserBaseURL = "https://aur.archlinux.org"
const UserRPCURL = UserBaseURL + "/rpc.php"

type UserRepo struct {
}

type UserSearchResponse struct {
	Version     int
	Type        string
	ResultCount int
	Results     []UserSearchResult
}

type UserSearchResult struct {
	ID             int
	Name           string
	PackageBaseID  int
	PackageBase    string
	Version        string
	Description    string
	URL            string
	NumVotes       int
	OutOfDate      int
	Maintainer     string
	FirstSubmitted int
	LastModified   int
	License        string
	URLPath        string
	CategoryID     int
}

func (repo UserRepo) Search(conf Conf) (bytes []byte, err error) {
	queryItems := []QueryItem{
		{Key: "type", Values: []string{"search"}},
		{Key: "arg", Values: conf.Args},
	}
	url := UserRPCURL + "?" + BuildQueryString(queryItems)
	res, err := http.Get(url)
	if err != nil {
		return
	}
	defer res.Body.Close()
	bytes, err = ioutil.ReadAll(res.Body)
	return
}

func (repo UserRepo) ParseSearchResponse(bytes []byte) (response UserSearchResponse, err error) {
	err = json.Unmarshal(bytes, &response)
	return
}

func (repo UserRepo) PrintSearchResponse(conf Conf) (err error) {
	bytes, err := repo.Search(conf)
	if err != nil {
		return
	}
	if conf.JsonFlag {
		fmt.Println(string(bytes[:]))
	} else {
		res, err := repo.ParseSearchResponse(bytes)
		if err != nil {
			return err
		}
		for _, pkg := range res.Results {
			color.New(color.FgBlue).Add(color.Bold).Print("aur")
			color.New(color.Bold).Printf("/%s ", pkg.Name)
			if pkg.OutOfDate != 0 {
				color.New(color.FgRed).Add(color.Bold).Print(pkg.Version)
			} else {
				color.New(color.Bold).Print(pkg.Version)
			}
			fmt.Printf(" (%d)\n", pkg.NumVotes)
			fmt.Printf("    %s\n", pkg.Description)
		}
	}
	return
}

func (repo UserRepo) Info(conf Conf) (bytes []byte, err error) {
	queryItems := []QueryItem{
		{Key: "type", Values: []string{"info"}},
		{Key: "arg", Values: conf.Args},
	}
	url := UserRPCURL + "?" + BuildQueryString(queryItems)
	res, err := http.Get(url)
	if err != nil {
		return
	}
	defer res.Body.Close()
	bytes, err = ioutil.ReadAll(res.Body)
	return
}

type UserInfoResponse struct {
	Version     int
	Type        string
	ResultCount int
	Results     UserInfoResult
}
type UserInfoResult UserSearchResult

func (repo UserRepo) ParseInfoResponse(bytes []byte) (response UserInfoResponse, err error) {
	err = json.Unmarshal(bytes, &response)
	return
}

func (repo UserRepo) PrintInfoResponse(conf Conf) (err error) {
	bytes, err := repo.Info(conf)
	if err != nil {
		return
	}
	if conf.JsonFlag {
		fmt.Println(string(bytes[:]))
	} else {
		res, err := repo.ParseInfoResponse(bytes)
		if err != nil {
			return err
		}
		pkg := res.Results
		str := `Repository      : %s
Name            : %s
Version         : %s
Description     : %s
URL             : %s
License         : %s
Maintainer      : %s
Submitted       : %s
Last Modified   : %s
Snapshot        : %s
Votes           : %d
`
		re := regexp.MustCompile("\\s*([^:]+:)([^\\n]*)\\n")
		str = re.ReplaceAllString(str, "\x1b[1m$1\x1b[0m$2\n")
		submitDate := time.Unix(int64(pkg.FirstSubmitted), 0)
		modifiedDate := time.Unix(int64(pkg.LastModified), 0)
		fmt.Printf(str, "aur", pkg.Name, pkg.Version, pkg.Description, pkg.URL, pkg.License,
			pkg.Maintainer, submitDate, modifiedDate,
			pkg.URLPath, pkg.NumVotes)
	}
	return
}

func (repo UserRepo) GetInfoToDownload(conf Conf) (res UserInfoResponse, url string, err error) {
	bytes, err := repo.Info(conf)
	if err != nil {
		return
	}
	res, err = repo.ParseInfoResponse(bytes)
	if err != nil {
		return
	}
	url = UserBaseURL + res.Results.URLPath
	return
}

func (repo UserRepo) DownloadTarGz(conf Conf, url string, outDir string) (newOutFilePath string, err error) {
	outFile, newOutFilePath, err := createOutFile(outDir, url)
	defer outFile.Close()
	if err != nil {
		return
	}
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(outFile, resp.Body)
	return
}

func (repo UserRepo) Get(conf Conf) (newOutFilePath string, err error) {
	info, url, err := repo.GetInfoToDownload(conf)
	if err != nil {
		return
	}
	result := info.Results

	destDir := path.Join(conf.OutDir, result.Name)
	_, err = os.Stat(destDir)
	if err == nil {
		err = errors.New(destDir + " already exists")
		return
	}
	err = nil

	tempDir, err := ioutil.TempDir("", "srchway-")
	if err != nil {
		return
	}

	color.New(color.FgBlue).Add(color.Bold).Println("Downloading " + url + " ...")
	tarGzPath, err := repo.DownloadTarGz(conf, url, tempDir)
	if err != nil {
		return
	}

	color.New(color.FgBlue).Add(color.Bold).Println("Extracting " + tarGzPath + " ...")
	newOutFilePath, err = ExtractAndRemoveTarGz(tarGzPath)

	srcDir := path.Join(newOutFilePath, result.Name)
	color.New(color.FgBlue).Add(color.Bold).Println("Copying " + srcDir + " to " + destDir + " ...")
	err = shutil.CopyTree(srcDir, destDir, nil)
	if err != nil {
		return
	}

	err = os.RemoveAll(tempDir)
	return
}
