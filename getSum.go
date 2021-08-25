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
	"fmt"
	"os"
	"path"
	"strings"
)

func getSums(filename string) (sums map[string]string) {
	sums = make(map[string]string)

	dir_name, file_name := path.Split(filename)
	sum_name := path.Join(dir_name, fmt.Sprintf(".%s.sum", file_name))

	if _, err := os.Stat(sum_name); os.IsNotExist(err) {
		processFile(filename)
	}

	sum_file, err := os.OpenFile(sum_name, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("missing", filename)
		return nil
	}
	defer sum_file.Close()

	sum_scanner := bufio.NewScanner(sum_file)
	var line string
	for {
		if sum_scanner.Scan() {
			line = sum_scanner.Text()
		} else {
			break
		}
		parts := strings.SplitN(line, ": ", 2)
		if len(parts) == 2 {
			sums[parts[0]] = strings.TrimSpace(parts[1])
		}
	}
	return
}
