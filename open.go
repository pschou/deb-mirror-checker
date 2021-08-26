package main

import (
	"compress/gzip"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ulikunitz/xz"
)

func open(name string) (io.Reader, error, func()) {
	if strings.HasPrefix(name, "http") {
		resp, err := client.Get(name)
		if err != nil {
			log.Println(err)
			return nil, err, func() {}
		}

		if strings.HasSuffix(name, ".gz") {
			gzr, err := gzip.NewReader(resp.Body)
			if err != nil {
				log.Println(err)
				defer resp.Body.Close()
				return nil, err, func() {}
			}
			return gzr, err, func() {
				gzr.Close()
				resp.Body.Close()
			}
		} else if strings.HasSuffix(name, ".xz") {
			xzr, err := xz.NewReader(resp.Body)
			if err != nil {
				log.Println(err)
				defer resp.Body.Close()
				return nil, err, func() {}
			}
			return xzr, err, func() {
				xzr = nil
				resp.Body.Close()
			}
		} else {
			return resp.Body, err, func() {
				resp.Body.Close()
			}
		}
	} else {

		file, err := os.OpenFile(name, os.O_RDONLY, 0666)
		if err != nil {
			log.Println(err)
			return nil, err, func() {}
		}

		if strings.HasSuffix(name, ".gz") {
			gzr, err := gzip.NewReader(file)
			if err != nil {
				log.Println(err)
				defer file.Close()
				return nil, err, func() {}
			}
			defer gzr.Close()
			return gzr, err, func() {
				gzr.Close()
				file.Close()
			}
		} else if strings.HasSuffix(name, ".xz") {
			xzr, err := xz.NewReader(file)
			if err != nil {
				log.Println(err)
				defer file.Close()
				return nil, err, func() {}
			}
			return xzr, err, func() {
				xzr = nil
				file.Close()
			}
		} else {
			return file, err, func() {
				file.Close()
			}
		}
	}

}
