package mbox_test

import (
	"bytes"
	"context"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aerth/mbox"
)

func TestMboxConcurrency(t *testing.T) {
	testMboxConcurrency(t, true)
}
func TestMbox(t *testing.T) {
	testMboxConcurrency(t, false)
}

// simple test to write to stdout
func TestStdoutMbox(t *testing.T) {
	msg := mbox.Form{
		From:    "aerth <aerth@localhost>",
		Subject: "It works",
		Message: "This really works!",
	}
	msg.WriteTo(os.Stdout)
	msg.WriteTo(os.Stdout)
	msg.WriteTo(os.Stdout)
}

// test multiple writes to a buffer (concurrent and non-concurrent)
func testMboxConcurrency(t *testing.T, conc bool) {
	err := mbox.Open(context.Background(), "test.mbox")
	// Report errors
	if err != nil {
		t.Errorf("opening mbox: %v", err)
		return
	}
	var wg sync.WaitGroup
	limit := 10
	if false {
		limit = 0
	}
	for i := 1; i <= limit; i++ {
		wg.Add(1)
		fn := func() { // Build the email
			defer wg.Done()
			var form mbox.Form
			form.From = "aerth <aerth@localhost>"
			form.Subject = "As seen on TV!!! " + strconv.Itoa(i)
			form.Message = "This really works! " + strconv.Itoa(i)
			if len(paragraphs) != 0 { // use paragraph from input text
				form.Message = paragraphs[rand.Intn(len(paragraphs))]
			}
			err = mbox.Save(&form) // Add the message to the mailbox
			// Report errors
			if err != nil {
				t.Errorf("writing message: %v", err)
			}
		}
		if conc {
			go fn()
		} else {
			fn()
		}
	}
	time.Sleep(time.Second) // only so message date is different
	wg.Wait()               // goroutine messages
	Save := mbox.Save
	type Form = mbox.Form
	err = Save(&Form{From: "test@test.test", Subject: "test empty lines", Message: "test empty lines\none\n\ntwo\n\n\nthree\n\n\n\nfour\n\n\n\n\nfive\n\n\n\n\n\nend?"})
	if err != nil {
		t.Fatalf("writing message: %v", err)
	}
	err = Save(&Form{From: "", Subject: "", Message: "it works with empty from and subject, and is timestamped"})
	if err != nil {
		t.Fatalf("writing message: %v", err)
	}
	err = Save(&Form{From: "", Subject: "", Message: "it works with empty from and subject, and is timestamped"})
	if err != nil {
		t.Fatalf("writing message: %v", err)
	}
	err = Save(&Form{From: "", Subject: "", Message: "it works with empty from and subject, and is timestamped"})
	if err != nil {
		t.Fatalf("writing message: %v", err)
	}
	log.Printf("now run: `mutt -F .muttrc`")
	mbox.Close() // stop the Loop goroutine that was started by mbox.Open
}

func shorten(s string) string {
	if len(s) > 64 {
		return s[:64]
	}
	return s
}

func getlines(name string) [][]byte {
	buf, err := os.ReadFile(name)
	if err != nil {
		panic(err.Error())
	}
	buf = bytes.Replace(buf, []byte("\r\n"), []byte("\n"), -1)
	buf = bytes.Split(buf, []byte("\nCHAPTER I."))[1]
	buf = bytes.Split(buf, []byte("*** END OF THE PROJECT"))[0]
	return bytes.Split(buf, []byte("\n"))
}
func getparagraphs(lines [][]byte) []string {
	var tmp []string
	var paragraphs []string
	numpg := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(string(line))
		empty := len(trimmed) == 0
		if empty {
			numpg++
		}
		if !empty && numpg > 2 {
			numpg = 0
			x := len(tmp)
			y := x
			for tmp[y-(y-x)-1] == "" {
				x--
			}
			paragraphs = append(paragraphs, strings.Join(tmp[:x], "\n"))
			if false {
				log.Printf("paragraph %d: %q", i, shorten(strings.Join(tmp[:x], "\n")))
			}
			tmp = []string{}
			continue
		}
		if empty && len(tmp) == 0 {
			numpg = 0
			continue
		}
		if len(tmp) == 0 && !firstcharOK(trimmed) {
			// skip lines that are cut off
			continue
		}
		tmp = append(tmp, string(line))
	}
	return paragraphs
}

// loadParagraphs loads paragraphs from a project gutenberg file
func loadParagraphs(name string) []string {
	lines := getlines(name)
	paragraphs := getparagraphs(lines)
	if len(paragraphs) > 5 {
		return paragraphs[5:]
	}
	return paragraphs
}

// firstcharOK returns true if the first character of the string is a valid first character for a paragraph
func firstcharOK(trimmed string) bool {
	okchar := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ“'\"```")
	for _, c := range okchar {
		if trimmed[0] == c {
			return true
		}
	}
	return false
}

var paragraphs = func() []string {
	x := loadParagraphs("testdata/pg11.txt")
	//log.Printf("got %d paragraphs", len(x))
	return x
}()
