package mbox_test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"sync"
	"testing"

	"filippo.io/age"
	"github.com/aerth/mbox"
)

// TestAgeEncryption tests age encryption and decryption for multiple mbox entries
// This serves as a simple example of how to implement a secure message storage system in a single mbox file
func TestAgeEncryption(t *testing.T) {
	gen, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Public key: %s\n", gen.Recipient())
	fmt.Printf("Private key: %s\n", gen.String())
	outfile := new(bytes.Buffer)
	var mu sync.Mutex
	msg := mbox.Form{
		From:    "age1u997c6ekf0mqcjr28mfctd2lf53hf7hay0tyr058ysle6vzfe9qqlnkd7d",
		Subject: "Hello, Bob",
		Message: "Hello, Bob! I hope you are well. I am sending you a secret message.",
	}

	// recip, err := age.ParseX25519Recipient("age1667eglrxwtz6hzgz2m0n70vkmkyjk6xggp8c223fqrje8wuedewqeeqkcy")
	// if err != nil {
	// 	t.Fatal(err)
	// 	return
	// }
	recip := gen.Recipient()
	for i := 0; i < 30; i++ {
		tmp := new(bytes.Buffer)
		encrypter, err := age.Encrypt(tmp, recip)
		if err != nil {
			t.Fatal(err)
		}
		_, err = msg.WriteTo(encrypter)
		if err != nil {
			t.Fatal(err)
		}
		if err := encrypter.Close(); err != nil {
			t.Fatal(err)
		}
		// if using encrypted mbox, separate the encrypted messages with a separator
		//
		mu.Lock()
		fmt.Fprintf(outfile, "%s", tmp.String())
		mu.Unlock()
	}

	{ // decrypt
		// id, err := age.ParseX25519Identity("AGE-SECRET-KEY-1LEVCZA4F9ESN8ULKWWYFF45HK7EYGLN6J5F0TW56QL284YT27JES6YVGZ3")
		// if err != nil {
		// 	t.Fatal(err)
		// }
		id := gen
		slices := bytes.Split(bytes.TrimSpace(outfile.Bytes()), []byte("age-encryption.org/v1\n"))
		l := len(slices)
		for i, slice := range slices {
			//slice = bytes.TrimSpace(slice)
			if len(slice) == 0 {
				log.Printf("empty slice %d/%d: %s", i+1, l, string(slice))
				continue
			}
			// re-add the first line of age header
			slice = append([]byte("age-encryption.org/v1\n"), slice...)
			//fmt.Printf("slice %d/%d: %d bytes\n", i+1, l, len(string(slice)))
			dec, err := age.Decrypt(bytes.NewReader(slice), id)
			if err != nil {
				log.Printf("bad slice %d/%d: %q", i+1, l, string(slice))
				t.Fatal(err)
			}

			msg, err := io.ReadAll(dec)
			if err != nil {
				log.Printf("bad slice %d/%d: %q", i+1, l, string(slice))
				t.Fatal(err)
			} else {
				//	log.Printf("good slice %d/%d: %q", i+1, l, string(slice))
			}

			fmt.Printf("\n\nDecrypted message %d/%d:\n%s", i+1, l, string(msg))
		}
	}

}

// TestGzip tests gzip compression and decompression for a single mbox entry
func TestGzip(t *testing.T) {

	msg := mbox.Form{
		From:    "Alice <alice@localhost>",
		Subject: "Hello, Bob",
		Message: "Hello, Bob! I hope you are well. I am sending you a message",
	}
	{
		gzipped := new(bytes.Buffer)
		exp, _ := msg.WriteTo(io.Discard)
		enc := gzip.NewWriter(gzipped)
		_, err := msg.WriteTo(enc)
		if err != nil {
			t.Fatal(err)
		}
		enc.Close()
		got := gzipped.Len()
		fmt.Printf("gzipped message: %d compressed to %d bytes (%2.2f %%)\n", exp, got, 100-(100*float64(got)/float64(exp)))
		if int64(got) >= exp {
			t.Fatalf("expected %d bytes, got %d", exp, got)
		}
		//fmt.Printf("%s\n", gzipped.String())
	}
}
