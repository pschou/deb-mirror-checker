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
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func verify(name string, keyring openpgp.KeyRing) (err error) {
	var signature_block *armor.Block
	var p packet.Packet
	//var file_sig *packet.Signature
	var issuerKeyId uint64
	var keys []openpgp.Key
	//var hashTag [2]byte
	//var hashSuffix []byte

	zr, err, file_close := open(name)
	if err != nil {
		return
	}
	defer file_close()

	scanner := bufio.NewScanner(zr)
	file_hashes := make(map[string]map[string]string)
	first_line, canonical := true, true
	SECTION := 0
	HEAD := 1
	CONTENT := 2
	var hash hash.Hash
	var line, hash_section string
	for {
		if scanner.Scan() {
			line = scanner.Text()
		} else {
			break
		}
		if strings.HasPrefix(line, "-----") {
			if line == "-----BEGIN PGP SIGNED MESSAGE-----" {
				SECTION = HEAD
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
			p, err = packet.Read(signature_block.Body)
			//fmt.Println("assigning p", p)
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
				issuerKeyId = *sig.IssuerKeyId
				signed_at = sig.CreationTime
				if hash == nil {
					hash = sig.Hash.New()
					canonical = false
				}
				//hashTag = sig.HashTag
				//hashSuffix = sig.HashSuffix
			case *packet.SignatureV3:
				issuerKeyId = sig.IssuerKeyId
				signed_at = sig.CreationTime
				if hash == nil {
					hash = sig.Hash.New()
					canonical = false
				}
				//hashTag = sig.HashTag
				//hashSuffix = []byte{}
			default:
				return errors.New("Signature block is invalid")
			}

			if issuerKeyId == 0 {
				return errors.New("Signature doesn't have an issuer")
			}

			if keyring == nil {
				fmt.Printf("  %s - Signed by 0x%02X at %v\n", name, issuerKeyId, signed_at)
				return
			} else {
				fmt.Printf("Verifying %s has been signed by 0x%02X at %v...\n", name, issuerKeyId, signed_at)
			}
			keys = keyring.KeysByIdUsage(issuerKeyId, packet.KeyFlagSign)

			if len(keys) == 0 {
				return errors.New("error: No matching public key found to verify")
			}
			if len(keys) > 1 {
				fmt.Println("warning: More than one public key found matching KeyID")
			}

			//fmt.Printf("Signature_block: %+v\n", signature_block)
			if SECTION == 0 && strings.HasSuffix(name, ".gpg") {
				r, err, signed_close := open(strings.TrimSuffix(name, ".gpg"))
				if err == nil {
					zr = r
					scanner = bufio.NewScanner(zr)
					defer signed_close()
					SECTION = CONTENT
					continue
				}
				//hash = p.Hash.New()
			}
			break
		}
		if hash != nil && SECTION == CONTENT {
			//fmt.Println([]byte(line))
			if canonical {
				if first_line {
					first_line = false
				} else {
					hash.Write([]byte{'\r', '\n'})
					//fmt.Printf("h: %02x\n", []byte{'\r', '\n'})
				}
				hash.Write([]byte(line))
				//fmt.Printf("h: %02x\n", line)
			} else {
				hash.Write([]byte(line + "\n"))
			}
		}
		if !strings.HasPrefix(line, " ") {
			if SECTION == HEAD {
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
					SECTION = CONTENT
				}
				continue
			}

			switch val := strings.TrimSpace(line); val {
			case "MD5Sum:":
				hash_section = "MD5sum"
			case "SHA1:", "SHA256:", "SHA512:":
				hash_section = strings.TrimSuffix(line, ":")
			}
		} else {
			parts := strings.Fields(strings.TrimSpace(line))
			if len(parts) > 2 {
				if _, ok := file_hashes[parts[2]]; !ok {
					file_hashes[parts[2]] = make(map[string]string)
				}
				file_hashes[parts[2]][hash_section] = parts[0]
				file_hashes[parts[2]]["Size"] = parts[1]
			}
		}
	}

	if signature_block == nil {
		return errors.New("Missing signature block in file")
	}

	if len(keys) > 0 {
		//fmt.Println("length of keys:", len(keys))
		//fmt.Printf("current hash: %02x\n", hash.Sum(nil))
		//fmt.Printf("hash suffix: %02x\n", hashSuffix)
		//hash.Write(hashSuffix)
		//fmt.Printf("post suffix: %02x\n", hash.Sum(nil))
		//fmt.Printf("hash tag: %02x\n", hashTag)
		switch sig := p.(type) {
		case *packet.Signature:
			err = keys[0].PublicKey.VerifySignature(hash, sig)
		case *packet.SignatureV3:
			err = keys[0].PublicKey.VerifySignatureV3(hash, sig)
		default:
			fmt.Println("Could not determine key type")
			return errors.New("Invalid signature format / type")
		}

		if err != nil {
			fmt.Println("Failed verification")
			return errors.New("Failed verification")
		}
		//if err == nil {
		//	break
		//}
	}

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
	return err
}
