package srchway

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/fatih/color"
)

type OfficialRepo struct{}

const OfficialBaseURL = "https://www.archlinux.org/packages"
const OfficialCorePackageURL = "https://projects.archlinux.org/svntogit/packages.git/snapshot/packages"
const OfficialCommunityPackageURL = "https://projects.archlinux.org/svntogit/community.git/snapshot/community-packages"

type OfficialSearchResponse struct {
	Version int
	Limit   int
	Valid   bool
	Results []OfficialSearchResult
}

type OfficialSearchResult struct {
	Arch           string
	BuildDate      string `json:"build_date"`
	CompressedSize int    `json:"compressed_size"`
	Conflicts      []string
	Depends        []string
	Epoch          int
	FileName       string
	FlagDate       string `json:"flag_date"`
	Groups         []string
	InstalledSize  int    `json:"installed_size"`
	LastUpdate     string `json:"last_update"`
	Licenses       []string
	Maintainers    []string
	Packager       string
	PkgBase        string
	PkgDesc        string
	PkgName        string
	PkgRel         string
	PkgVer         string
	Provides       []string
	Replaces       []string
	Repo           string
	Url            string
}

func (repo OfficialRepo) Search(query string) (bytes []byte, err error) {
	url := OfficialBaseURL + "/search/json/?arch=x86_64&q=" + query
	fmt.Println(url)
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	bytes, err = ioutil.ReadAll(resp.Body)
	return
}

func (repo OfficialRepo) Info(query string) (bytes []byte, err error) {
	url := OfficialBaseURL + "/search/json/?arch=x86_64&q=" + query
	fmt.Println(url)
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	bytes, err = ioutil.ReadAll(resp.Body)
	return
}

func (repo OfficialRepo) Get(query string) (err error) {
	err = nil
	return
}

func (repo OfficialRepo) ParseSearchResponse(bytes []byte) (response OfficialSearchResponse, err error) {
	err = json.Unmarshal(bytes, &response)
	return
}

func (repo OfficialRepo) PrintSearchResponse(query string, mode PrintMode) (err error) {
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
			color.New(color.FgBlue).Add(color.Bold).Print(pkg.Repo)
			color.New(color.Bold).Printf("/%s %s-%s\n", pkg.PkgName, pkg.PkgVer, pkg.PkgRel)
			fmt.Printf("    %s\n", pkg.PkgDesc)
		}
	case JsonMode:
		fmt.Println(string(bytes[:]))
	}
	return
}
