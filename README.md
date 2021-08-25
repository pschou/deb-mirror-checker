# A basic Debian/Ubuntu mirror Checker

The intended purposes of this tool is to be a mirror maintainer.  With this one can quickly scan for any corrupted repository files, make lists of files and their sizes for easy chunking into removable media, or selectively download packages which are newer than a specified date.

Overall, a debian/ubuntu maintainer / checker

# Usage

```bash
$ deb-mirror-checker 
Debian mirror checker, written by Paul Schou gitlab.com/pschou/deb-mirror-checker (version: 0.1.DATECODE)

Usage:
  added [package_old] [package_new] - Compare two "Packages" and list files added with their size.
  check [package...]                - Use "Packages" to validate checksums of all the local repo files
  verify PGP_pub_keys [package...]  - Verify PGP signature in "InRelease" and validate checksums
  list [package...]                 - Use "Packages" and dump out a list of repo files and their size
  make [path...]                    - generate all the .sum files in a directory
  mtime [date] [baseurl] [package...] - Use "Packages" and dump out a list of remote files and their size modified after date.
  sum [package...]                  - Use "Packages" and total the number unique files and their size

Note: Your current working directory, "/tmp/mirror.umd.edu/ubuntu", must be the repo base directory.
Packages can be also provided in .gz or .xz formats and the file can be a local file or a URL endpoint.
```

Return code 0 means that all checksums are correct, and exitcode 1 means at least one checksum did not match.  The output may be string matched for the word missing to detect any files not verified.  The idea here is, it is better to know when one has a bad file more than when a file is missing, hence the exitcode boolean.

# Examples

To grab a list of all the files defined in Packages.gz:
```bash
$ deb-mirror-checker $( find dists/ -type f -name Packages.gz )
```

Let's say we want to download only the newest files for a particular Packages, we can do this easily using this along with wget:
```bash
$ deb-mirror-checker mtime 2021-08-01 https://archive.ubuntu.com/ubuntu https://archive.ubuntu.com/ubuntu/dists/focal-updates/main/binary-amd64/Packages.xz > newer.list
$ sed 's#^[0-9]* #https://archive.ubuntu.com/ubuntu/#' newer.list > newer_url.list
$ wget -nc -x -i newer_url.list
```

If one has already downloaded the Packages files and wants to instead, say, download all the newest repo files in this list:
```bash
$ deb-mirror-checker mtime 2021-07-01 https://archive.ubuntu.com/ubuntu $( find archive.ubuntu.com/ubuntu/dists/ -name Packages.gz ) > newer.list
$ sed 's#^[0-9]* #https://archive.ubuntu.com/ubuntu/#' newer.list > newer_url.list
$ wget -nc -x -i newer_url.list
```

After one has downloaded all the newest packages, to chunk these files for ease of transport, onto DVDs, one may use:
```bash
$ zip -0 -s 8G pool.zip -r archive.ubuntu.com/ubuntu/pool
```

Verify using a PGP keyring:
```bash
$ deb-mirror-checker verify /tmp/Hokeypuck.pgp $( find dists/ -name InRelease )
Loaded KeyID: 0x5EDB1B62EC4926EA
Loaded KeyID: 0x871920D1991BC93C
Loaded KeyID: 0x3B4FE6ACC0B21F32
Loaded KeyID: 0x40976EAF437D05B5
Verifying dists/bionic/InRelease - Signed by 0x3B4FE6ACC0B21F32 at 2018-04-26 19:38:40 -0400 EDT
Verifying dists/bionic-backports/InRelease - Signed by 0x3B4FE6ACC0B21F32 at 2021-08-25 08:17:30 -0400 EDT
Verifying dists/bionic-proposed/InRelease - Signed by 0x3B4FE6ACC0B21F32 at 2021-08-25 08:17:28 -0400 EDT
```
