package mbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"filippo.io/age"
	"github.com/goware/emailx"
)

// Writer channel is used to write emails to the mbox file one at a time
// Set before calling Open function.
// Channel is not closed by this package. (see Close function)
var Writer = make(chan Writable, 100)

// MailWriteCloser is the file to write to, default is os.Stdout. Set properly with Open function or similar.
var MailWriteCloser io.WriteCloser

var mainctx context.Context
var cancelwrite context.CancelFunc

func SetContext(ctx context.Context, cancel context.CancelFunc) {
	mainctx = ctx
	cancelwrite = cancel
}

var wg sync.WaitGroup

// Open mbox file, rw+create+append mode ( step 1 )
// If file is empty, we use os.Stdout
// use Close() to stop Loop goroutine
func Open(ctx context.Context, file string) (err error) {
	if mainctx != nil && mainctx.Err() == nil {
		panic("mail file is already open")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if file == "" {
		MailWriteCloser = os.Stdout
	} else {
		MailWriteCloser, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			MailWriteCloser = os.Stdout
			return err
		}
	}
	mainctx, cancelwrite = context.WithCancel(ctx)
	// Writer receives one email at a time
	wg.Add(1)
	go Loop(Writer, MailWriteCloser, wg.Done)
	return err
}

// AgeRecipient is the public key, to activate auto-encryption.
// Use 'age-keygen' from https://filippo.io/age
//
// Example: age1u997c6ekf0mqcjr28mfctd2lf53hf7hay0tyr058ysle6vzfe9qqlnkd7d
var AgeRecipient string

// Separator, if non-nil, is called after writing mbox message,
// useful for a custom encrypted mbox file implementation.
var Separator func(io.Writer)

// Destination is the address where mail is "sent",
// its useful to change this to the address you will be replying to.
//
// Example: "me@localhost" or empty string
var Destination = ""

// Loop goroutine closes the mbox file when context is finished,
// then calls donefn (wg.Wait() for example)
func Loop(incoming chan Writable, mailout io.WriteCloser, donefn func()) {
	if donefn == nil {
		donefn = func() {}
	}
	for {
		select {
		case <-mainctx.Done():
			cancelwrite()
			mailout.Close()
			donefn()
			return
		case form := <-incoming:
			if AgeRecipient == "" {
				_, err := form.WriteTo(mailout)
				if err != nil {
					log.Printf("error writing to mbox, please restart the program: %v", err)
				}
				if Separator != nil {
					Separator(mailout)
				}
				continue
			}
			if AgeRecipient != "" {
				recip, err := age.ParseX25519Recipient(AgeRecipient)
				if err != nil {
					cancelwrite()
					mailout.Close()
					donefn()
					panic(err.Error())
				}
				if Separator == nil {
					fmt.Fprintf(mailout, "# Encrypted message (age+aerth/mbox):\n")
				}
				encryptor, err := age.Encrypt(mailout, recip)
				if err != nil {
					log.Printf("error writing to mbox, please restart the program: %v", err)
					continue
				}
				_, err = form.WriteTo(encryptor)
				if err != nil {
					log.Printf("error writing to mbox, please restart the program: %v", err)
					continue
				}
				if err := encryptor.Close(); err != nil {
					log.Printf("error writing to mbox, please restart the program: %v", err)
					continue
				}
				if Separator != nil {
					Separator(mailout)
				}
				continue
			}

		}
	}
}

// Close the mbox file when finished (not always necessary)
func Close() {
	cancelwrite()
	wg.Wait()
}

func NewMessage(name, email, subject, message string) Form {
	var msg Form
	if name != "" && email != "" {
		msg.From = name + " <" + email + ">"
	} else if email != "" {
		msg.From = email
	} else {
		msg.From = name
	}
	msg.Subject = subject
	msg.Message = message
	return msg
}

// Save assigns Received time and sends an entire email to the writer.
func Save(form Writable) error {
	if MailWriteCloser == nil {
		panic("MailWriteCloser is nil, use Open()")
	}
	select {
	case <-mainctx.Done():
		return mainctx.Err()
	case Writer <- form:
	}
	return nil
}

// Normalize capitalization of email address
//
// see also: ValidationLevel (set 3 for full validation, 0 for none)
func (form *Form) Normalize() error {
	if ValidationLevel == 0 {
		return nil
	}

	// lvl 1 Normalize email address capitalization
	form.From = emailx.Normalize(form.From)

	if ValidationLevel > 1 {
		// lvl 2
		if form.From == "@" || form.From == " " || !strings.ContainsAny(form.From, "@") {
			return errors.New("email is empty or not valid")
		}

		if ValidationLevel > 2 {
			// lvl 3
			err := emailx.Validate(form.From)
			if err != nil {
				if err == emailx.ErrInvalidFormat {
					return errors.New("email is not valid format.")
				}
				if err == emailx.ErrUnresolvableHost {
					return errors.New("email is not valid format.")
				}
				return fmt.Errorf("email validation error: %v", err)
			}
		}
	}

	return nil
}

var NoSubjectLine = "[No Subject]" // default subject line if none is provided
var NoFromLine = "Unknown"         // default from line if none is provided

// Write form to mbox file
// TODO see RFC1123Z, RFC3339, RFC5322
func (form *Form) WriteTo(w io.Writer) (int64, error) {
	fm := strings.TrimSpace(form.Message)
	if fm == "" && len(form.Body) == 0 && form.Subject == "" && form.From == "" {
		return 0, errors.New("too many empty fields (message, body, subject, from)")
	}
	if form.Received.IsZero() {
		form.Received = time.Now().UTC()
	}
	if form.From == "" {
		form.From = NoFromLine
	}
	if form.Subject == "" {
		form.Subject = NoSubjectLine
	}
	if er := form.Normalize(); er != nil {
		return 0, er
	}
	mailtime := form.Received.Format("Mon Jan 2 15:04:05.99999 2006")
	mailtime2 := form.Received.Format("Mon, 2 Jan 2006 15:04:05.99999 -0700")

	space := string([]byte{0x20})
	// try and extract email address from From
	fromaddr := form.From
	if strings.Contains(fromaddr, "<") {
		fromaddr = strings.Split(fromaddr, "<")[1]
		fromaddr = strings.Split(fromaddr, ">")[0]
	} else if strings.Contains(fromaddr, " ") {
		fromaddr = strings.Replace(fromaddr, " ", "_", -1) // experimental: replace spaces with underscores
	}
	var x int64
	var n2 int
	lines := []string{
		"From" + space + strings.Replace(fromaddr, " ", "+", -1) + space + mailtime,
		"Return-path: <" + form.From + ">",
		"Delivery-date: " + mailtime2,
		"To: " + Destination, // skips if Destination is empty
		"Envelope-to: " + Destination + "\n",
		"Subject: " + form.Subject,
		"From: " + form.From,
		"Date: " + mailtime2,
	}
	for _, line := range lines {
		if strings.HasSuffix(strings.TrimSpace(line), ":") {
			continue // skip empty destination and other empty lines
		}
		n2, err := w.Write([]byte(line + "\n"))
		x += int64(n2)
		if err != nil {
			return x, err
		}
	}

	// end header
	n2, err := w.Write([]byte{'\n'})
	x += int64(n2)
	if err != nil {
		return x, err
	}

	// write message
	if fm != "" {
		n2, err = w.Write([]byte(fm + "\n"))
		x += int64(n2)
		if err != nil {
			return x, err
		}
	}

	// experimental: attachments
	if len(form.Body) != 0 {
		n2, err = w.Write(form.Body)
		x += int64(n2)
		if err != nil {
			return x, err
		}
	}
	// end message
	n2, err = w.Write([]byte("\n\n\n"))
	x += int64(n2)
	if err != nil {
		return x, err
	}
	return x, err
}
