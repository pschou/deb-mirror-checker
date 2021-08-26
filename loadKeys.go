package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/openpgp"
)

func loadKeys(keyfile string) (keyring openpgp.EntityList, err error) {
	var keyRingReader io.Reader
	keyRingReader, err = os.Open(keyfile)
	var loaded_keys openpgp.EntityList
	if err != nil {
		fmt.Println("Error opening keyring file:", err)
		return
	} else {
		fmt.Println("Loading keys from", keyfile)
	}

	scanner := bufio.NewScanner(keyRingReader)
	var line, keystr string
	var i int
	for {
		if scanner.Scan() {
			line = scanner.Text()
		} else {
			break
		}
		keystr += line + "\n"
		if strings.TrimSpace(line) == "-----END PGP PUBLIC KEY BLOCK-----" {
			i++
			loaded_keys, err = openpgp.ReadArmoredKeyRing(strings.NewReader(keystr))
			if err == nil {
				for _, key := range loaded_keys {
					keyring = append(keyring, key)
					fmt.Printf("  %d) Loaded KeyID: 0x%02X\n", i, key.PrimaryKey.KeyId)
				}
				keystr = ""
			} else {
				fmt.Printf("  %d) Invalid key: %g\n", i, err)
			}
		}
		if len(keyring) > 0 {
			err = nil
		}
	}
	//for _, entity := range []*openpgp.Entity(keyring) {
	//	fmt.Printf("Loaded KeyID: 0x%02X\n", entity.PrimaryKey.KeyId)
	//}
	return
}
