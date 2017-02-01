package srchway

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
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

func (repo UserRepo) Search(query string) (bytes []byte, err error) {
	url := UserRPCURL + "?type=search&arg=" + query
	fmt.Println(url)
	res, err := http.Get(url)
	defer res.Body.Close()
	if err != nil {
		return
	}
	bytes, err = ioutil.ReadAll(res.Body)
	return
}

func (repo UserRepo) ParseSearchResponse(bytes []byte) (response UserSearchResponse, err error) {
	err = json.Unmarshal(bytes, &response)
	return
}

func (repo UserRepo) PrintSearchResponse(query string, mode PrintMode) (err error) {
	bytes, err := repo.Search(query)
	if err != nil {
		return
	}
	switch mode {
	case NormalMode:
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
	case JsonMode:
		fmt.Println(string(bytes[:]))
	}
	return
}

func (repo UserRepo) Info(query string) (bytes []byte, err error) {
	url := UserRPCURL + "?type=info&arg=" + query
	fmt.Println(url)
	res, err := http.Get(url)
	defer res.Body.Close()
	if err != nil {
		return
	}
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

func (repo UserRepo) PrintInfoResponse(query string, mode PrintMode) (err error) {
	bytes, err := repo.Info(query)
	if err != nil {
		return
	}
	switch mode {
	case NormalMode:
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
	case JsonMode:
		fmt.Println(string(bytes[:]))
	}
	return
}

func (repo UserRepo) Get(query string, outFilePath string) (err error) {
	bytes, err := repo.Info(query)
	if err != nil {
		return
	}
	res, err := repo.ParseInfoResponse(bytes)
	if err != nil {
		return
	}
	url := UserBaseURL + "/" + res.Results.URLPath
	outFile, err := createOutFile(outFilePath, url)
	defer outFile.Close()
	if err != nil {
		return
	}
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	fmt.Printf("Downloading %s...\n", outFile.Name())
	_, err = io.Copy(outFile, resp.Body)
	return
}
