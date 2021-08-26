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
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/ulikunitz/xz"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func verify(name string, keyring openpgp.KeyRing) (err error) {
	var signature_block *armor.Block

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
				return err
			}
			defer gzr.Close()
			zr = io.Reader(gzr)
		} else if strings.HasSuffix(name, ".xz") {
			xzr, err := xz.NewReader(resp.Body)
			if err != nil {
				log.Println(err)
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
			log.Println(err)
			return err
		}
		defer file.Close()

		if strings.HasSuffix(name, ".gz") {
			gzr, err := gzip.NewReader(file)
			if err != nil {
				log.Println(err)
				return err
			}
			defer gzr.Close()
			zr = io.Reader(gzr)
		} else if strings.HasSuffix(name, ".xz") {
			xzr, err := xz.NewReader(file)
			if err != nil {
				log.Println(err)
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
	file_hashes := make(map[string]map[string]string)
	first_line := true
	var hash hash.Hash
	section := ""
	var line string
	for {
		if scanner.Scan() {
			line = scanner.Text()
		} else {
			break
		}
		if strings.HasPrefix(line, "-----") {
			if line == "-----BEGIN PGP SIGNED MESSAGE-----" {
				section = "head"
				continue
			}
			signature_str := line + "\n"
			for {
				if scanner.Scan() {
					signature_str += scanner.Text() + "\n"
				} else {
					break
				}
			}
			//fmt.Printf("Hash: %02x\n", hash.Sum(nil))
			//fmt.Printf("Signature:\n%s\n", signature_str)
			signature_block, err = armor.Decode(strings.NewReader(signature_str))
			//fmt.Printf("Signature_block: %+v\n", signature_block)
			break
		}
		if hash != nil && section != "head" {
			//fmt.Println([]byte(line))
			if first_line {
				first_line = false
			} else {
				hash.Write([]byte{'\r', '\n'})
			}
			hash.Write([]byte(line))
			//fmt.Printf("h: %02x\n", line)
		}
		if !strings.HasPrefix(line, " ") {
			if section == "head" {
				if strings.HasPrefix(line, "Hash: ") {
					kind := strings.TrimPrefix(line, "Hash: ")
					switch kind {
					case "SHA1":
						hash = sha1.New()
					case "SHA256":
						hash = sha256.New()
					case "SHA512":
						hash = sha512.New()
					}
					//fmt.Println("found hash", kind, hash)
				}
				if line == "" {
					section = "content"
				}
				continue
			}

			switch val := strings.TrimSpace(line); val {
			case "MD5Sum:":
				section = "MD5sum"
			case "SHA1:", "SHA256:", "SHA512:":
				section = strings.TrimSuffix(line, ":")
			}
		} else {
			parts := strings.Fields(strings.TrimSpace(line))
			if len(parts) > 2 {
				if _, ok := file_hashes[parts[2]]; !ok {
					file_hashes[parts[2]] = make(map[string]string)
				}
				file_hashes[parts[2]][section] = parts[0]
				file_hashes[parts[2]]["Size"] = parts[1]
			}
		}
	}

	var issuerKeyId uint64
	var keys []openpgp.Key
	var p packet.Packet

	if signature_block == nil {
		return errors.New("Missing signature block in file")
	}
	p, err = packet.Read(signature_block.Body)
	if err == io.EOF {
		return errors.New("Unable to read signature block")
	}
	if err != nil {
		fmt.Println("Error in signature:", err)
		return err
	}
	var signed_at time.Time

	switch sig := p.(type) {
	case *packet.Signature:
		if sig.IssuerKeyId == nil {
			return errors.New("Signature doesn't have an issuer")
		}
		issuerKeyId = *sig.IssuerKeyId
		signed_at = sig.CreationTime
	case *packet.SignatureV3:
		issuerKeyId = sig.IssuerKeyId
		signed_at = sig.CreationTime
	default:
		return errors.New("Signature block is invalid")
	}

	if keyring == nil {
		fmt.Printf("  %s - Signed by 0x%02X at %v\n", name, issuerKeyId, signed_at)
		return
	} else {
		fmt.Printf("Verifying %s - Signed by 0x%02X at %v\n", name, issuerKeyId, signed_at)
	}
	keys = keyring.KeysByIdUsage(issuerKeyId, packet.KeyFlagSign)

	if len(keys) == 0 {
		return errors.New("error: No matching public key found to verify")
	}

	for _, key := range keys {
		switch sig := p.(type) {
		case *packet.Signature:
			err = key.PublicKey.VerifySignature(hash, sig)
		case *packet.SignatureV3:
			err = key.PublicKey.VerifySignatureV3(hash, sig)
		default:
			fmt.Println("Could not determine key type")
			return errors.New("Invalid signature format / type")
		}

		if err == nil {
			break
		}
	}

	if err != nil {
		fmt.Println("  No public key matching signature could be found")
		return errors.New("  No public key matching signature could be found")
	} else {
		for filename, sums := range file_hashes {
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				// If the file does not exist, test to see if it is in the dist
				// directory with the InRelease file
				d, _ := path.Split(name)
				test_filename := path.Join(d, filename)
				if _, err := os.Stat(test_filename); !os.IsNotExist(err) {
					// Found it, so we'll test on this file name
					filename = test_filename
				} else {
					// We did not find it, so let us see if any of the compressed/uncompressed alternatives are there
					test_filename = strings.TrimSuffix(strings.TrimSuffix(test_filename, ".gz"), ".xz")
					switch {
					case func() bool { _, err := os.Stat(test_filename); return !os.IsNotExist(err) }():
					case func() bool { _, err := os.Stat(test_filename + ".gz"); return !os.IsNotExist(err) }():
					case func() bool { _, err := os.Stat(test_filename + ".xz"); return !os.IsNotExist(err) }():
					default:
						fmt.Println("missing", filename)
					}
					continue
				}
			}

			file_sums := getSums(filename)
			for k, v := range sums {
				if file_sums[k] != v {
					fmt.Printf("Failed_%s %s  %s != %s\n", k, filename, file_sums[k], v)
					err = errors.New("failed verification")
					continue
				}
				delete(sums, k)
			}
		}
	}
	return err
}
