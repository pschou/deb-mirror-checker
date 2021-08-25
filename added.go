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
	"strings"

	"github.com/ulikunitz/xz"
)

func added(old_name, new_name string) {

	var zr io.Reader

	file_list := make(map[string]string)

	func() {
		if strings.HasPrefix(old_name, "http") {
			resp, err := client.Get(old_name)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			if strings.HasSuffix(old_name, ".gz") {
				gzr, err := gzip.NewReader(resp.Body)
				if err != nil {
					log.Println(err)
					return
				}
				defer gzr.Close()
				zr = io.Reader(gzr)
			} else if strings.HasSuffix(old_name, ".xz") {
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

			file, err := os.OpenFile(old_name, os.O_RDONLY, 0666)
			if err != nil {
				log.Println(err)
				return
			}
			defer file.Close()

			if strings.HasSuffix(old_name, ".gz") {
				gzr, err := gzip.NewReader(file)
				if err != nil {
					log.Println(err)
					return
				}
				defer gzr.Close()
				zr = io.Reader(gzr)
			} else if strings.HasSuffix(old_name, ".xz") {
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
		var line, filename, h_md5, h_sha1, h_sha256, h_sha512, size string
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
			case "Size":
				size = val
			case "MD5sum":
				h_md5 = val
			case "SHA1":
				h_sha1 = val
			case "SHA256":
				h_sha256 = val
			case "SHA512":
				h_sha512 = val
			case "":
				if filename != "" {
					file_list[filename] = fmt.Sprintf("%s|%s|%s|%s|%s", size, h_md5, h_sha1, h_sha256, h_sha512)
				}
				filename, h_md5, h_sha1, h_sha256, h_sha512, size = "", "", "", "", "", ""
				continue
			}
		}
	}()

	func() {
		if strings.HasPrefix(old_name, "http") {
			resp, err := client.Get(old_name)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			if strings.HasSuffix(old_name, ".gz") {
				gzr, err := gzip.NewReader(resp.Body)
				if err != nil {
					log.Println(err)
					return
				}
				defer gzr.Close()
				zr = io.Reader(gzr)
			} else if strings.HasSuffix(old_name, ".xz") {
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

			file, err := os.OpenFile(old_name, os.O_RDONLY, 0666)
			if err != nil {
				log.Println(err)
				return
			}
			defer file.Close()

			if strings.HasSuffix(old_name, ".gz") {
				gzr, err := gzip.NewReader(file)
				if err != nil {
					log.Println(err)
					return
				}
				defer gzr.Close()
				zr = io.Reader(gzr)
			} else if strings.HasSuffix(old_name, ".xz") {
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
		var line, filename, h_md5, h_sha1, h_sha256, h_sha512, size string
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
			case "Size":
				size = val
			case "MD5sum":
				h_md5 = val
			case "SHA1":
				h_sha1 = val
			case "SHA256":
				h_sha256 = val
			case "SHA512":
				h_sha512 = val
			case "":
				if filename != "" {
					hash, ok := file_list[filename]
					if !ok || hash != fmt.Sprintf("%s|%s|%s|%s|%s", size, h_md5, h_sha1, h_sha256, h_sha512) {
						fmt.Printf("%d %s\n", size, filename)
					}
				}
				filename, h_md5, h_sha1, h_sha256, h_sha512, size = "", "", "", "", "", ""
				continue
			}
		}
	}()
}
