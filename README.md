# A basic Debian/Ubuntu mirror Checker

The intended purposes of this tool is to be a mirror maintainer.  With this one can quickly scan for any corrupted repository files, make lists of files and their sizes for easy chunking into removable media, or selectively download packages which are newer than a specified date.

Overall, a debian/ubuntu maintainer / checker

# Usage

```bash
$ ./deb-mirror-checker 
Debian mirror checker

Usage:
  list [package...]  - Use "Packages" and dump out a list of repo files and their size
  make [path...]  - generate all the .sum files in a directory
  check [package...] - Use "Packages" to verify all the local repo files
  mtime [date] [baseurl] [package...] - Use "Packages" and dump out a list of remote files and their size modified after date.

Note: Your current working directory, "/home/schou/go/src/ubuntu-mirror-checker", must be the repo base directory.
  One can use Packages in .gz or .xz format and the file can be a local file or a URL endpoint.

```

# Examples

To grab a list of all the files defined in Packages.gz:
```bash
./deb-mirror-checker $( find dists/ -type f -name Packages.gz )
```

Let's say we want to download only the newest files for a particular Packages, we can do this easily using this along with wget:
```bash
$ ./deb-mirror-checker mtime 2021-08-01 https://archive.ubuntu.com/ubuntu https://archive.ubuntu.com/ubuntu/dists/focal-updates/main/binary-amd64/Packages.xz > newer.list
$ sed 's#^[0-9]* #https://archive.ubuntu.com/ubuntu/#' newer.list > newer_url.list
$ wget -nc -x -i newer_url.list
```

If one has already downloaded the Packages files and wants to instead, say, download all the newest repo files in this list:
```bash
$ ./deb-mirror-checker mtime 2021-07-01 https://archive.ubuntu.com/ubuntu $( find archive.ubuntu.com/ubuntu/dists/ -name Packages.gz ) > newer.list
$ sed 's#^[0-9]* #https://archive.ubuntu.com/ubuntu/#' newer.list > newer_url.list
$ wget -nc -x -i newer_url.list
```
