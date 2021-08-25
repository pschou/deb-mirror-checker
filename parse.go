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
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ulikunitz/xz"
)

func parse(name string) (err error) {
	fmt.Println("Checking", name)

	var zr io.Reader

	if strings.HasPrefix(name, "http") {
		resp, err := client.Get(name)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		if strings.HasSuffix(name, ".gz") {
			gzr, err := gzip.NewReader(resp.Body)
			if err != nil {
				return err
			}
			defer gzr.Close()
			zr = io.Reader(gzr)
		} else if strings.HasSuffix(name, ".xz") {
			xzr, err := xz.NewReader(resp.Body)
			if err != nil {
				return err
			}
			//defer xzr.Close()
			defer func() { xzr = nil }()
			zr = io.Reader(xzr)
		} else {
			zr = resp.Body
		}

	} else {

		file, err := os.OpenFile(name, os.O_RDONLY, 0666)
		if err != nil {
			return err
		}
		defer file.Close()

		if strings.HasSuffix(name, ".gz") {
			gzr, err := gzip.NewReader(file)
			if err != nil {
				return err
			}
			defer gzr.Close()
			zr = io.Reader(gzr)
		} else if strings.HasSuffix(name, ".xz") {
			xzr, err := xz.NewReader(file)
			if err != nil {
				return err
			}
			//defer xzr.Close()
			defer func() { xzr = nil }()
			zr = io.Reader(xzr)
		} else {
			zr = file
		}
	}

	scanner := bufio.NewScanner(zr)
	var line, filename string
	sums := make(map[string]string)
	for {
		if scanner.Scan() {
			line = scanner.Text()
		} else {
			break
		}
		parts := strings.SplitN(line, ": ", 2)
		val := ""
		if len(parts) == 2 {
			val = strings.TrimSpace(parts[1])
		}
		switch parts[0] {
		case "Filename":
			filename = val
		case "Size", "MD5sum", "SHA1", "SHA256", "SHA512":
			sums[parts[0]] = val
		case "":
			if filename != "" {
				if _, err := os.Stat(filename); os.IsNotExist(err) {
					fmt.Println("missing", filename)
					continue
				}
				file_sums := getSums(filename)
				for k, v := range sums {
					if file_sums[k] != v {
						fmt.Printf("getsums: %+v\n", file_sums)
						fmt.Printf("Failed_%s %s (%s != %s)\n", k, filename, file_sums[k], v)
						err = errors.New("failed verification")
						continue
					}
					delete(sums, k)
				}

			}

		}
	}
	return
}
