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
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

func process(name string) {
	fi, err := os.Stat(name)
	if err != nil {
		//fmt.Println(err)
		return
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		files, err := ioutil.ReadDir(name)
		if err != nil {
			log.Println(err)
			return
		}
		for _, f := range files {
			process(path.Join(name, f.Name()))
		}
	case mode.IsRegular():
		processFile(name)
	}
}

var buf = make([]byte, 512*1024)

func processFile(name string) {
	dir_name, file_name := path.Split(name)
	if strings.HasPrefix(file_name, ".") {
		//fmt.Println("has prefix", file_name)
		return
	}
	sum_name := path.Join(dir_name, fmt.Sprintf(".%s.sum", file_name))
	//fmt.Println("making", sum_name)
	//stat, _ := os.Stat(name)
	if _, err := os.Stat(sum_name); !os.IsNotExist(err) {
		//fmt.Printf("stat: %+v\n", stat)
		//fmt.Println("already exists", sum_name)
		return
	}

	file, err := os.OpenFile(name, os.O_RDONLY, 0666)
	defer file.Close()
	if err != nil {
		//log.Println(err)
		return
	}

	h_md5 := md5.New()
	h_sha1 := sha1.New()
	h_sha256 := sha256.New()
	total := uint64(0)

	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Println(err)
			return
		}

		if n < len(buf) {
			to_write := buf[:n]
			h_md5.Write(to_write)
			h_sha1.Write(to_write)
			h_sha256.Write(to_write)
		} else {
			h_md5.Write(buf)
			h_sha1.Write(buf)
			h_sha256.Write(buf)
		}
		total = total + uint64(n)
	}

	out, err := os.Create(sum_name)
	if err != nil {
		log.Println(err)
		return
	}
	defer out.Close()
	fmt.Fprintf(out, "Size: %d\n", total)
	fmt.Fprintf(out, "MD5sum: %x\n", h_md5.Sum(nil))
	fmt.Fprintf(out, "SHA1: %x\n", h_sha1.Sum(nil))
	fmt.Fprintf(out, "SHA256: %x\n", h_sha256.Sum(nil))
}
