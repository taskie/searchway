package main

import (
	"errors"
	"fmt"
	"github.com/taskie/srchway/lib"
	"os"
	"strings"
)

type OperationType int
const (
	OperationTypeNone OperationType = iota
	OperationTypeSearch
	OperationTypeInfo
	OperationTypeGet
	OperationTypeHelp
)

type Env struct {
	operation OperationType
	args         []string
	verbose      bool
	aurFlag      bool
	officialFlag bool
	jsonFlag     bool
}

func (env Env) repos() (repos []srchway.Repo) {
	repos = make([]srchway.Repo, 0)
	if env.officialFlag {
		repos = append(repos, srchway.OfficialRepo{})
	}
	if env.aurFlag {
		repos = append(repos, srchway.UserRepo{})
	}
	return
}

func version(env Env) (exitCode int) {
	fmt.Println("0.0")
	exitCode = 0
	return
}

func search(env Env) (exitCode int) {
	exitCode = 1
	query := strings.Join(env.args, "+")
	repos := env.repos()
	for _, repo := range repos {
		mode := srchway.NormalMode
		if env.jsonFlag {
			mode = srchway.JsonMode
		}
		err := repo.PrintSearchResponse(query, mode)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			exitCode = 0
		}
	}
	return
}

func info(env Env) (exitCode int) {
	exitCode = 1
	query := strings.Join(env.args, "+")
	repos := env.repos()
	for _, repo := range repos {
		mode := srchway.NormalMode
		if env.jsonFlag {
			mode = srchway.JsonMode
		}
		err := repo.PrintInfoResponse(query, mode)
		if err == nil {
			exitCode = 0
			break
		}
	}
	return
}

func get(env Env) (exitCode int) {
	exitCode = 1
	query := strings.Join(env.args, "+")
	repos := env.repos()
	for _, repo := range repos {
		_, err := repo.Get(query, "")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		} else {
			exitCode = 0
			break
		}
	}
	return
}

const usage = `usage: srchway [OPERATION] [OPTIONS] [QUERY]
OPERATION:
    -s, --search   search package
    -i, --info     show package info
    -g, --get      get PKGBUILD
    -h, --help     show help

OPTIONS:
    -a, --aur      use AUR
    -A, --auronly  use AUR only (no offcial repo)
    -j, --json     output raw JSON (when --search, --info)
    -v, --verbose  verbose mode`

func parseOption(arg string, env *Env) (err error) {
	if !strings.HasPrefix(arg, "--") && strings.HasPrefix(arg, "-") {
		for i := 1; i < len(arg); i += 1 {
			parseOption(arg[i:i+1], env)
		}
		return
	}

	switch arg {
	case "S":

	case "s", "--search":
		env.operation = OperationTypeSearch
	case "i", "--info":
		env.operation = OperationTypeInfo
	case "g", "--get", "G":
		env.operation = OperationTypeGet
	case "h", "--help":
		env.operation = OperationTypeHelp
	case "a", "--aur":
		env.aurFlag = true
	case "A", "--auronly":
		env.aurFlag = true
		env.officialFlag = false
	case "j", "--json":
		env.jsonFlag = true
	case "v", "--verbose":
		env.verbose = true
	default:
		err = errors.New("unknown option: " + arg)
		return
	}
	return
}

func parseArgs(args []string) (env Env, err error) {
	env.officialFlag = true
	breakCount := 0
	for i, arg := range args[1:] {
		if arg == "--" {
			breakCount = i + 1
			break
		} else if arg[0] == '-' {
			err = parseOption(arg, &env)
			if err != nil {
				return
			}
		} else {
			breakCount = i
			break
		}
	}
	if env.operation == OperationTypeNone {
		err = errors.New("you must specify just one operation type")
		return
	}
	env.args = args[breakCount + 1:]
	return
}

func main() {
	env, err := parseArgs(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	exitCode := 0
	switch env.operation {
	case OperationTypeSearch:
		exitCode = search(env)
	case OperationTypeInfo:
		exitCode = info(env)
	case OperationTypeGet:
		exitCode = get(env)
	case OperationTypeHelp:
		fmt.Println(usage)
	default:
		exitCode = 1
	}
	os.Exit(exitCode)
}
