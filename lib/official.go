package srchway

import (
	"code.cloudfoundry.org/bytefmt"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
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
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	bytes, err = ioutil.ReadAll(resp.Body)
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

type OfficialInfoResponse OfficialSearchResult

func (repo OfficialRepo) InfoFromPackage(repoName string, pkgName string) (bytes []byte, err error) {
	url := OfficialBaseURL + fmt.Sprintf("/%s/x86_64/%s/json", repoName, pkgName)
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return
	}
	bytes, err = ioutil.ReadAll(resp.Body)
	return
}

func (repo OfficialRepo) InfoFromSearch(query string) (bytes []byte, err error) {
	bytes, err = repo.Search(query)
	if err != nil {
		return
	}
	res, err := repo.ParseSearchResponse(bytes)
	if err != nil {
		return bytes, err
	}
	var result OfficialSearchResult
	for _, r := range res.Results {
		if r.PkgName == query {
			result = r
		}
	}
	if result.PkgName == query {
		bytes, err = repo.InfoFromPackage(result.Repo, result.PkgName)
		return
	} else {
		err = errors.New("not found")
		return bytes, err
	}
}

func (repo OfficialRepo) Info(query string) (bytes []byte, err error) {
	if strings.Contains(query, "/") {
		parts := strings.Split(query, "/")
		bytes, err = repo.InfoFromPackage(parts[0], parts[1])
		return
	} else {
		bytes, err = repo.InfoFromSearch(query)
		return
	}
}

func (repo OfficialRepo) ParseInfoResponse(bytes []byte) (response OfficialInfoResponse, err error) {
	err = json.Unmarshal(bytes, &response)
	return
}

func (repo OfficialRepo) PrintInfoResponse(query string, mode PrintMode) (err error) {
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
		str := `Repository      : %s
Name            : %s
Version         : %s-%s
Description     : %s
Architecture    : %s
URL             : %s
Licenses        : %s
Groups          : %s
Provides        : %s
Depends On      : %s
Optional Deps   : %s
Conflicts With  : %s
Replaces        : %s
Download Size   : %s
Installed Size  : %s
Packager        : %s
Build Date      : %s
`
		re := regexp.MustCompile("\\s*([^:]+:)([^\\n]*)\\n")
		str = re.ReplaceAllString(str, "\x1b[1m$1\x1b[0m$2\n")
		buildDate, _ := time.Parse(time.RFC3339, res.BuildDate)
		deps := make([]string, 0)
		optdeps := make([]string, 0)
		for _, v := range res.Depends {
			if strings.Contains(v, ":") {
				optdeps = append(optdeps, v)
			} else {
				deps = append(deps, v)
			}
		}
		fmt.Printf(str, res.Repo, res.PkgName, res.PkgVer, res.PkgRel, res.PkgDesc, res.Arch, res.Url,
			joinOrNoneString(res.Licenses), joinOrNoneString(res.Groups), joinOrNoneString(res.Provides),
			joinOrNoneString(deps), joinOrNoneStringForOptDepends(optdeps), joinOrNoneString(res.Conflicts), joinOrNoneString(res.Replaces),
			bytefmt.ByteSize(uint64(res.CompressedSize)), bytefmt.ByteSize(uint64(res.InstalledSize)),
			res.Packager, buildDate)
	case JsonMode:
		fmt.Println(string(bytes[:]))
	}
	return
}

func (repo OfficialRepo) Get(query string, outFilePath string) (newOutFilePath string, err error) {
	bytes, err := repo.Info(query)
	if err != nil {
		return
	}
	res, err := repo.ParseInfoResponse(bytes)
	if err != nil {
		return
	}
	var url string
	switch res.Repo {
	case "core", "extra":
		url = OfficialCorePackageURL + "/" + res.PkgBase + ".tar.gz"
	default:
		url = OfficialCommunityPackageURL + "/" + res.PkgBase + ".tar.gz"
	}
	outFile, newOutFilePath, err := createOutFile(outFilePath, url)
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
