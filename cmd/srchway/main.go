package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/taskie/srchway"
)

func search(conf srchway.Conf) (exitCode int) {
	exitCode = 1
	repos := conf.Repos()
	for _, repo := range repos {
		err := repo.PrintSearchResponse(conf)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			exitCode = 0
		}
	}
	return
}

func info(conf srchway.Conf) (exitCode int) {
	exitCode = 1
	repos := conf.Repos()
	for _, repo := range repos {
		err := repo.PrintInfoResponse(conf)
		if err == nil {
			exitCode = 0
			break
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	return
}

func get(conf srchway.Conf) (exitCode int) {
	exitCode = 1
	repos := conf.Repos()
	for _, repo := range repos {
		_, err := repo.Get(conf)
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
    -s, --search    search package
    -i, --info      show package info
    -g, --get       get PKGBUILD
    -h, --help      show help

OPTIONS:
    -a, --aur       use AUR
    -A, --auronly   use AUR only (no offcial repo)
    -m, --multilib  use multilib repo
    -t, --testing   use testing repo    
    -j, --json      output raw JSON (when --search, --info)
    -v, --verbose   verbose mode`

func help(conf srchway.Conf) (exitCode int) {
	fmt.Println(usage)
	exitCode = 0
	return
}

func version(conf srchway.Conf) (exitCode int) {
	fmt.Println(srchway.VersionString)
	exitCode = 0
	return
}

func parseOption(arg string, conf *srchway.Conf) (err error) {
	if !strings.HasPrefix(arg, "--") && strings.HasPrefix(arg, "-") {
		for i := 1; i < len(arg); i++ {
			parseOption(arg[i:i+1], conf)
		}
		return
	}

	switch arg {
	case "S":
		// do nothing
	case "s", "--search":
		conf.Operation = srchway.OperationTypeSearch
	case "i", "--info":
		conf.Operation = srchway.OperationTypeInfo
	case "g", "--get", "G":
		conf.Operation = srchway.OperationTypeGet
	case "h", "--help":
		conf.Operation = srchway.OperationTypeHelp
	case "V", "--version":
		conf.Operation = srchway.OperationTypeVersion
	case "a", "--aur":
		conf.AurFlag = true
	case "A", "--auronly":
		conf.AurFlag = true
		conf.OfficialFlag = false
	case "m", "--multilib":
		conf.MultilibFlag = true
	case "t", "--testing":
		conf.TestingFlag = true
	case "j", "--json":
		conf.JsonFlag = true
	case "v", "--verbose":
		conf.Verbose = true
	default:
		err = errors.New("unknown option: " + arg)
		return
	}
	return
}

func parseArgs(args []string) (conf srchway.Conf, err error) {
	conf.OfficialFlag = true
	breakCount := 0
	for i, arg := range args[1:] {
		if arg == "--" {
			breakCount = i + 1
			break
		} else if arg[0] == '-' {
			err = parseOption(arg, &conf)
			if err != nil {
				return
			}
		} else {
			breakCount = i
			break
		}
	}
	if conf.Operation == srchway.OperationTypeNone {
		err = errors.New("you must specify just one operation type")
		return
	}
	conf.Args = args[breakCount+1:]
	return
}

func main() {
	conf, err := parseArgs(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	exitCode := 0
	switch conf.Operation {
	case srchway.OperationTypeSearch:
		exitCode = search(conf)
	case srchway.OperationTypeInfo:
		exitCode = info(conf)
	case srchway.OperationTypeGet:
		exitCode = get(conf)
	case srchway.OperationTypeHelp:
		exitCode = help(conf)
	case srchway.OperationTypeVersion:
		exitCode = version(conf)
	default:
		exitCode = 1
	}
	os.Exit(exitCode)
}
