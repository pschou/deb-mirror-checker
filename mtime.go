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
	"time"

	"github.com/araddon/dateparse"
	"github.com/ulikunitz/xz"
)

func mtime(name string, mt time.Time, url string) {

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
	var filename, size string
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
		case "":
			if filename != "" {
				resp, err := client.Head(url + filename)
				modtime := resp.Header.Get("Last-Modified")
				t, err := dateparse.ParseStrict(modtime)
				if err == nil && t.After(mt) {
					fmt.Println(size, filename)
				} else {
					//fmt.Println("skipped", size, filename, t)
				}
			}
			filename, size = "", ""
			continue
		}
	}
}
