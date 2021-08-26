// Copyright 2021 Paul Schou
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/araddon/dateparse"
)

var version = ""

func main() {
	var exitcode int
	if len(os.Args) > 1 && os.Args[1] == "make" {
		for _, name := range os.Args[2:] {
			process(name)
		}
	} else if len(os.Args) > 1 && os.Args[1] == "check" {
		for _, name := range os.Args[2:] {
			err := parse(name)
			if err != nil {
				exitcode = 1
			}
		}
	} else if len(os.Args) > 2 && os.Args[1] == "verify" {
		keyRing, err := loadKeys(os.Args[2])
		if err != nil {
			fmt.Println("Error opening keyring file:", err)
			for _, name := range os.Args[3:] {
				verify(name, nil)
			}
			os.Exit(1)
		}
		for _, name := range os.Args[3:] {
			err := verify(name, keyRing)
			if err != nil {
				fmt.Println("error:", err)
				exitcode = 1
			}
		}
	} else if len(os.Args) > 1 && os.Args[1] == "list" {
		for _, name := range os.Args[2:] {
			list(name)
		}
	} else if len(os.Args) > 1 && os.Args[1] == "sum" {
		pb := &sum_passback{file_sizes: make(map[string]string)}
		for _, name := range os.Args[2:] {
			sum(name, pb)
		}
		fmt.Println("Files:", pb.count)
		fmt.Println("Total size:", pb.total)
	} else if len(os.Args) == 4 && os.Args[1] == "added" {
		added(os.Args[2], os.Args[3])
	} else if len(os.Args) > 3 && os.Args[1] == "mtime" {
		t, err := dateparse.ParseAny(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		url := strings.TrimSuffix(os.Args[3], "/") + "/"
		for _, name := range os.Args[4:] {
			mtime(name, t, url)
		}
	} else {
		dir, _ := os.Getwd()
		fmt.Printf("Debian mirror checker, written by Paul Schou gitlab.com/pschou/deb-mirror-checker (version: %s)\n\n", version)
		fmt.Println("Usage:\n",
			" added [package_old] [package_new] - Compare two \"Packages\" and list files added with their size.\n",
			" check [package...]                - Use \"Packages\" to validate checksums of all the local repo files\n",
			" verify PGP_KeyRing.pub [pgp_file...] - Verify PGP armored signature either attached or detached and validate checksums\n",
			"                                        The .pgp file must have the signed file in the same directory without the .pgp\n",
			" list [package...]                 - Use \"Packages\" and dump out a list of repo files and their size\n",
			" make [path...]                    - generate all the .sum files in a directory\n",
			" mtime [date] [baseurl] [package...] - Use \"Packages\" and dump out a list of remote files and their size modified after date.\n",
			" sum [package...]                  - Use \"Packages\" and total the number unique files and their size\n",
		)
		fmt.Printf("Note: Your current working directory, %q, must be the repo base directory.\n", dir)
		fmt.Println("Packages can be also provided in .gz or .xz formats and the file can be a local file or a URL endpoint.")
		return
	}
	os.Exit(exitcode)
}
