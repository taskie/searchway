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
	"strings"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/fatih/color"
	"github.com/termie/go-shutil"
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

func (repo OfficialRepo) BuildSearchQueryItems(conf Conf) (queryItems []QueryItem) {
	repoNames := []string{"Core", "Extra", "Community"}
	if conf.TestingFlag {
		repoNames = append(repoNames, "Testing", "Community-Testing")
	}
	if conf.MultilibFlag {
		repoNames = append(repoNames, "Multilib")
		if conf.TestingFlag {
			repoNames = append(repoNames, "Multilib-Testing")
		}
	}
	queryItems = []QueryItem{{Key: "arch", Values: []string{"x86_64"}}}
	for _, repoName := range repoNames {
		queryItems = append(queryItems, QueryItem{Key: "repo", Values: []string{repoName}})
	}
	queryItems = append(queryItems, QueryItem{Key: "q", Values: conf.Args})
	return
}

func (repo OfficialRepo) Search(conf Conf) (bytes []byte, err error) {
	queryItems := repo.BuildSearchQueryItems(conf)
	url := OfficialBaseURL + "/search/json/?" + BuildQueryString(queryItems)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	bytes, err = ioutil.ReadAll(resp.Body)
	return
}

func (repo OfficialRepo) ParseSearchResponse(bytes []byte) (response OfficialSearchResponse, err error) {
	err = json.Unmarshal(bytes, &response)
	return
}

func (repo OfficialRepo) PrintSearchResponse(conf Conf) (err error) {
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
			color.New(color.FgBlue).Add(color.Bold).Print(pkg.Repo)
			color.New(color.Bold).Printf("/%s %s-%s\n", pkg.PkgName, pkg.PkgVer, pkg.PkgRel)
			fmt.Printf("    %s\n", pkg.PkgDesc)
		}
	}
	return
}

type OfficialInfoResponse OfficialSearchResult

func (repo OfficialRepo) InfoFromPackage(repoName string, pkgName string) (bytes []byte, err error) {
	url := OfficialBaseURL + fmt.Sprintf("/%s/x86_64/%s/json", repoName, pkgName)
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	bytes, err = ioutil.ReadAll(resp.Body)
	return
}

func (repo OfficialRepo) InfoFromSearch(conf Conf) (bytes []byte, err error) {
	bytes, err = repo.Search(conf)
	if err != nil {
		return
	}
	res, err := repo.ParseSearchResponse(bytes)
	if err != nil {
		return bytes, err
	}
	results := []OfficialSearchResult{}
	for _, r := range res.Results {
		if r.PkgName == conf.Args[0] {
			results = append(results, r)
		}
	}
	switch len(results) {
	case 0:
		err = errors.New("not found")
		return
	case 1:
		bytes, err = repo.InfoFromPackage(results[0].Repo, results[0].PkgName)
		return
	default:
		names := []string{}
		for _, result := range results {
			names = append(names, result.Repo+"/"+result.PkgName)
		}
		err = errors.New("found multiple results (" + strings.Join(names, ", ") + ")")
		return
	}
}

func (repo OfficialRepo) Info(conf Conf) (bytes []byte, err error) {
	if len(conf.Args) == 0 {
		err = errors.New("please specify package name")
		return
	}
	query := conf.Args[0]
	if strings.Contains(query, "/") {
		parts := strings.Split(query, "/")
		bytes, err = repo.InfoFromPackage(parts[0], parts[1])
		return
	}
	bytes, err = repo.InfoFromSearch(conf)
	return
}

func (repo OfficialRepo) ParseInfoResponse(bytes []byte) (response OfficialInfoResponse, err error) {
	err = json.Unmarshal(bytes, &response)
	return
}

func (repo OfficialRepo) PrintInfoResponse(conf Conf) (err error) {
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
	}
	return
}

func (repo OfficialRepo) GetInfoToDownload(conf Conf) (res OfficialInfoResponse, isCommunity bool, url string, err error) {
	bytes, err := repo.Info(conf)
	if err != nil {
		return
	}
	res, err = repo.ParseInfoResponse(bytes)
	if err != nil {
		return
	}
	switch res.Repo {
	case "core", "extra", "testing":
		isCommunity = false
	default:
		isCommunity = true
	}
	if isCommunity {
		url = OfficialCommunityPackageURL + "/" + res.PkgBase + ".tar.gz"
	} else {
		url = OfficialCorePackageURL + "/" + res.PkgBase + ".tar.gz"
	}
	return
}

func (repo OfficialRepo) DownloadTarGz(conf Conf, url string, outDir string) (newOutFilePath string, err error) {
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

func (repo OfficialRepo) Get(conf Conf) (newOutFilePath string, err error) {
	info, isCommunity, url, err := repo.GetInfoToDownload(conf)
	if err != nil {
		return
	}

	destDir := path.Join(conf.OutDir, info.PkgName)
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
	var packagesDir string
	if isCommunity {
		packagesDir = "community-packages"
	} else {
		packagesDir = "packages"
	}

	srcDir := path.Join(newOutFilePath, packagesDir, info.PkgName, "repos", info.Repo+"-x86_64")
	color.New(color.FgBlue).Add(color.Bold).Println("Copying " + srcDir + " to " + destDir + " ...")
	err = shutil.CopyTree(srcDir, destDir, nil)
	if err != nil {
		return
	}

	err = os.RemoveAll(tempDir)
	return
}
