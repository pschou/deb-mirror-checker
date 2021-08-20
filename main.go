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
	if len(os.Args) > 1 && os.Args[1] == "make" {
		for _, name := range os.Args[2:] {
			process(name)
		}
	} else if len(os.Args) > 1 && os.Args[1] == "check" {
		for _, name := range os.Args[2:] {
			parse(name)
		}
	} else if len(os.Args) > 1 && os.Args[1] == "list" {
		for _, name := range os.Args[2:] {
			list(name)
		}
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
			" list [package...]  - Use \"Packages\" and dump out a list of repo files and their size\n",
			" make [path...]  - generate all the .sum files in a directory\n",
			" check [package...] - Use \"Packages\" to verify all the local repo files\n",
			" mtime [date] [baseurl] [package...] - Use \"Packages\" and dump out a list of remote files and their size modified after date.\n")
		fmt.Printf("Note: Your current working directory, %q, must be the repo base directory.\n", dir)
		fmt.Println("  One can use Packages in .gz or .xz format and the file can be a local file or a URL endpoint.")
		return
	}
}
