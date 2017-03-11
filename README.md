# srchway

Application to search Arch Linux official/user repositories and download PKGBUILD file.

## Usage

```
usage: srchway [OPERATION] [OPTIONS] [QUERY]
OPERATION:
    -s, --search    search package
    -i, --info      show package info
    -g, --get       get PKGBUILD
    -h, --help      show help
    -V, --version   show version

OPTIONS:
    -a, --aur       use AUR
    -A, --auronly   use AUR only (no offcial repo)
	-m, --multilib  use multilib repo
	-t, --testing   use testing repo
    -j, --json      output raw JSON (when --search, --info)
    -v, --verbose   verbose mode
```

### Search

```bash
srchway -s emacs
srchway -sm gcc
srchway -smt lib32-
srchway -sa ttf-
srchway -sA ttf-
```

### Info

```bash
srchway -i linux
srchway -i core/linux
srchway -i testing/linux
srchway -iA linux-rt
```

### Get

```bash
srchway -g linux
srchway -g core/linux
srchway -g testing/linux
srchway -gA linux-rt
```

# contrib/srchway-dl

*Potentially Dangerous!*

Shell script to download/clone sources written on PKGBUILD and check MD5/SHA256/SHA512 sums.

## Usage

```
usage: srchway-dl [DIRECTORY which has PKGBUILD]
usage: srchway-dl
```

```bash
srchway -g extra/zsh
cd zsh
srchway-dl
```
# LICENSE

Apache License 2.0
