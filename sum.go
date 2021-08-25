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
	"strconv"
	"strings"
	"sync"

	"github.com/ulikunitz/xz"
)

type sum_passback struct {
	lock       sync.Mutex
	file_sizes map[string]string
	count      uint
	total      uint64
}

func sum(name string, pb *sum_passback) {

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
	var line, filename, size string
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
		case "":
			if filename != "" {
				fsize, ok := pb.file_sizes[filename]
				if ok {
					if size != fsize {
						fmt.Println("Warning", filename, "has two different sizes,", fsize, "and", size)
					}
				} else {
					val, err := strconv.ParseUint(size, 10, 64)
					if err == nil {
						pb.file_sizes[filename] = size
						pb.total = pb.total + val
						pb.count++
					}
				}
			}
			filename, size = "", ""
			continue
		}
	}
}
