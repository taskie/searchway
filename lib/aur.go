package srchway

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/fatih/color"
)

const UserBaseURL = "https://aur.archlinux.org"
const UserRPCURL = UserBaseURL + "/rpc.php"

type UserRepo struct {
}

type UserSearchResponse struct {
	Version int
	Type string
	ResultCount int
	Results []UserSearchResult
}

type UserSearchResult struct {
	ID int
	Name string
	PackageBaseID int
	PackageBase string
	Version string
	Description string
	URL string
	NumVotes int
	OutOfDate int
	Maintainer string
	FirstSubmitted int
	LastModified int
	License string
	URLPath string
	CategoryID int
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

func (repo UserRepo) Get(query string) (err error) {
	err = nil
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
