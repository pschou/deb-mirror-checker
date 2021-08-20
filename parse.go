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
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/ulikunitz/xz"
)

func parse(name string) {
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
				log.Println(err)
				return
			}
			defer gzr.Close()
			zr = io.Reader(gzr)
		} else if strings.HasSuffix(name, ".xz") {
			xzr, err := xz.NewReader(resp.Body)
			if err != nil {
				log.Println(err)
				return
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
			log.Println(err)
			return
		}
		defer file.Close()

		if strings.HasSuffix(name, ".gz") {
			gzr, err := gzip.NewReader(file)
			if err != nil {
				log.Println(err)
				return
			}
			defer gzr.Close()
			zr = io.Reader(gzr)
		} else if strings.HasSuffix(name, ".xz") {
			xzr, err := xz.NewReader(file)
			if err != nil {
				log.Println(err)
				return
			}
			//defer xzr.Close()
			defer func() { xzr = nil }()
			zr = io.Reader(xzr)
		} else {
			zr = file
		}
	}

	scanner := bufio.NewScanner(zr)
	var size, h_sha256, h_sha1, h_md5, filename string
parse_line:
	for line := scanner.Text(); scanner.Scan(); line = scanner.Text() {
		parts := strings.SplitN(line, ": ", 2)
		val := ""
		if len(parts) == 2 {
			val = strings.TrimSpace(parts[1])
		}
		switch parts[0] {
		case "Filename":
			filename = val
		case "Size":
			size = val
		case "MD5sum":
			h_md5 = val
		case "SHA1":
			h_sha1 = val
		case "SHA256":
			h_sha256 = val
		case "":
			if filename != "" {

				/*if _, err := os.Stat(filename); os.IsNotExist(err) {
					fmt.Println("missing", filename)
					continue parse_line
				}*/

				dir_name, file_name := path.Split(filename)
				sum_name := path.Join(dir_name, fmt.Sprintf(".%s.sum", file_name))

				if _, err := os.Stat(sum_name); os.IsNotExist(err) {
					processFile(filename)
				}

				sum_file, err := os.OpenFile(sum_name, os.O_RDONLY, 0666)
				if err != nil {
					//log.Println(err)
					fmt.Println("missing", filename)
					continue parse_line
					//return
				}
				//defer sum_file.Close()

				sum_scanner := bufio.NewScanner(sum_file)
				for line := sum_scanner.Text(); sum_scanner.Scan(); line = sum_scanner.Text() {
					parts := strings.SplitN(line, ": ", 2)
					val := ""
					if len(parts) == 2 {
						val = strings.TrimSpace(parts[1])
					}
					switch parts[0] {
					case "Size":
						if size != val {
							fmt.Println("failed", filename)
							sum_file.Close()
							continue parse_line
						}
					case "MD5sum":
						if h_md5 != val {
							fmt.Println("failed", filename)
							sum_file.Close()
							continue parse_line
						}
					case "SHA1":
						if h_sha1 != val {
							fmt.Println("failed", filename)
							sum_file.Close()
							continue parse_line
						}
					case "SHA256":
						if h_sha256 != val {
							fmt.Println("failed", filename)
							sum_file.Close()
							continue parse_line
						}
					}
				}
				sum_file.Close()
				size, h_sha256, h_sha1, h_md5, filename = "", "", "", "", ""
			}

		}
	}
}
