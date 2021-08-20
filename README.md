# A basic Ubuntu/Debian mirror checker

The intended purpose of this tool is to be able to scan a mirror quickly for any files that are corrupted by providing these files to the commandline of this checker program.  For example:

```bash
./ubunt-mirror-checker $( find dists/ -type f -mtime -10 -name Packages.gz )
```

# Usage

```bash
$ ./ubuntu-mirror-checker 
Ubunbtu mirror checker

Usage:
  list [package...]  - Use "Packages" and dump out a list of repo files and their size
  make [path...]  - generate all the .sum files in a directory
  check [package...] - Use "Packages" to verify all the local repo files
  mtime [date] [baseurl] [package...] - Use "Packages" and dump out a list of remote files and their size modified after date.

Note: Your current working directory, "/home/schou/go/src/ubuntu-mirror-checker", must be the repo base directory.
  One can use Packages in .gz or .xz format and the file can be a local file or a URL endpoint.

```

# Example

Let's say we want to download only the newest files for a package, we can do this easily using this along with wget:
```bash
$ ./ubuntu-mirror-checker mtime 2021-08-01 https://archive.ubuntu.com/ubuntu https://archive.ubuntu.com/ubuntu/dists/focal-updates/main/binary-amd64/Packages.xz > newer.list
$ sed 's#^[0-9]* #https://archive.ubuntu.com/ubuntu/#' newer.list > newer_url.list
$ wget -nc -x -i newer_url.list
```
