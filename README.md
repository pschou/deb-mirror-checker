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

Note: Your current working directory, "/tmp", must be the repo base directory.
Packages can be also provided in .gz or .xz formats and the file can be a local file or a URL endpoint.
```

Return code 0 means that all checksums are correct, and exitcode 1 means at least one checksum did not match.  The output may be string matched for the word missing to detect any files not verified.  The idea here is, it is better to know when one has a bad file more than when a file is missing, hence the exitcode boolean.

# Examples

To grab a list of all the files defined in Packages.gz:
```bash
$ deb-mirror-checker list $( find dists/ -type f -name Packages.gz )
```

Determine the entire size of the required files for the repository:
```bash
$ deb-mirror-checker sum $( find dists/ -type f -name Packages.gz )
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

Verify using Packages.gz:
```bash
$ deb-mirror-checker check $( find dists/ -type f -name Packages.gz )
```

Verify chain of custody using a PGP keyring and deb packages using the InRelease files:
```bash
$ deb-mirror-checker verify /tmp/Hockeypuck.keys dists/bionic-proposed/InRelease
Loading keys from /tmp/Hokeypuck.key
  1) Loaded KeyID: 0x5EDB1B62EC4926EA
  2) Loaded KeyID: 0x3B4FE6ACC0B21F32
  3) Loaded KeyID: 0x871920D1991BC93C
  4) Loaded KeyID: 0x40976EAF437D05B5
Verifying dists/bionic-proposed/InRelease has been signed by 0x3B4FE6ACC0B21F32 at 2021-08-25 08:17:28 -0400 EDT...
...
```

Verify chain of custody using a PGP keyring and the image file checksums using SHA256SUMS files:
```bash
$ deb-mirror-checker verify /tmp/Hockeypuck.keys $( find dists/ -name SHA256SUMS.gpg )
Loading keys from /tmp/Hockeypuck.key
  1) Loaded KeyID: 0x5EDB1B62EC4926EA
  2) Loaded KeyID: 0x3B4FE6ACC0B21F32
  3) Loaded KeyID: 0x871920D1991BC93C
  4) Loaded KeyID: 0x40976EAF437D05B5
Verifying dists/bionic/main/installer-amd64/current/images/SHA256SUMS.gpg has been signed by 0x3B4FE6ACC0B21F32 at 2018-04-25 17:23:28 -0400 EDT...
Verifying dists/bionic/main/installer-i386/current/images/SHA256SUMS.gpg has been signed by 0x3B4FE6ACC0B21F32 at 2018-04-25 17:23:19 -0400 EDT...
Verifying dists/bionic-proposed/main/installer-amd64/current/images/SHA256SUMS.gpg has been signed by 0x3B4FE6ACC0B21F32 at 2020-08-03 04:58:27 -0400 EDT...
Verifying dists/bionic-proposed/main/installer-i386/current/images/SHA256SUMS.gpg has been signed by 0x3B4FE6ACC0B21F32 at 2020-08-03 05:13:51 -0400 EDT...
Verifying dists/bionic-updates/main/installer-amd64/current/images/SHA256SUMS.gpg has been signed by 0x3B4FE6ACC0B21F32 at 2020-08-05 08:43:56 -0400 EDT...
```
