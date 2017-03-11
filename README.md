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
srchway -sA ttf-
```

### Info

```bash
srchway -i core/linux
srchway -i linux
srchway -ia slack-desktop
```

### Get

```bash
srchway -g core/gcc
srchway -g gcc
srchway -ga linux-rt
```

# srchdown

*Potentially Dangerous!*

Shell script to download/clone sources written on PKGBUILD and check MD5/SHA256/SHA512 sums.

## Usage

```
usage: srchdown [DIRECTORY which has PKGBUILD]
usage: srchdown
```

```bash
srchway -g extra/zsh
tar zvxf zsh.tar.gz
cd packages/zsh/trunk
srchdown
```
# LICENSE

Apache License 2.0
