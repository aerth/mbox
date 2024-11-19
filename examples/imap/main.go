package main

import (
	"bytes"
	"flag"
	"fmt"
	"net/mail"
	"os"
	"path/filepath"
	"time"

	"github.com/aerth/mbox"
	"github.com/xarg/imap"
)

func main() {
	var (
		filename         = "imap.mbox"
		seq              = "1:5"
		fetchAll         = os.Getenv("IMAP_ALL") != ""
		imaphost         = os.Getenv("IMAP_HOST")
		c                *imap.Client
		cmd              *imap.Command
		rsp              *imap.Response
		err              error
		user, pass              = os.Getenv("IMAP_USER"), os.Getenv("IMAP_PASS")
		startnum, endnum string = "1", "5"
		showversion      bool
	)

	flag.Usage = func() {
		exename := filepath.Base(os.Args[0])
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", exename)
		fmt.Fprintf(os.Stderr, "\t  %s -f filename\n", exename)
		fmt.Fprintf(os.Stderr, "\t  %s -f filename -a\n", exename)
		fmt.Fprintf(os.Stderr, "\t  %s -f filename -seq 1:5\n", exename)
		fmt.Fprintf(os.Stderr, "\t  %s -f filename -fetchfrom 1 -fetchto 5\n", exename)
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Environment variables:\n")
		fmt.Fprintf(os.Stderr, "\tIMAP_USER\n")
		fmt.Fprintf(os.Stderr, "\tIMAP_PASS\n")
		fmt.Fprintf(os.Stderr, "\tIMAP_HOST\n")
		fmt.Fprintf(os.Stderr, "\tIMAP_ALL\n")
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Command line flags:\n")
		flag.PrintDefaults()
	}
	flag.StringVar(&filename, "f", filename, "mbox filename")
	flag.BoolVar(&fetchAll, "a", fetchAll, "fetch all messages (or IMAP_ALL=1, see -seq flag)")
	flag.StringVar(&seq, "seq", seq, "sequence of messages to fetch, eg: 1:5 or 1,2,3,4,5 or 1:*\nsee also -fetchfrom and -fetchto, comma separated RFC 3501 sequence-set ABNF rule")
	flag.StringVar(&startnum, "fetchfrom", startnum, "start at including message number")
	flag.StringVar(&endnum, "fetchto", endnum, "end including message number \ncan use * if your shell allows --fetchto=\\* or -fetchto=*\n")
	flag.BoolVar(&showversion, "version", false, "show version and exit")
	flag.Parse()
	if showversion {
		fmt.Printf("mbox version %s\n", mbox.Version)
		fmt.Printf("library/source: https://github.com/aerth/mbox\n")
		return
	}

	if len(startnum) == 0 {
		startnum = "1"
	}
	if len(endnum) == 0 {
		endnum = "5"
	}
	if startnum != "1" || endnum != "5" {
		seq = startnum + ":" + endnum
	}
	if user == "" || pass == "" {
		fmt.Println("Use IMAP_USER, IMAP_PASS, and IMAP_HOST variables")
		return
	}
	// Connect to the server
	c, err = imap.DialTLS(imaphost, nil)

	if c == nil || err != nil {
		fmt.Println("Error: can't connect.")
		if err != nil {
			fmt.Println(err)
		}
		return
	}
	// Remember to log out and close the connection when finished
	defer c.Logout(30 * time.Second)

	// Print server greeting (first response in the unilateral server data queue)
	fmt.Println("Server says hello:", c.Data[0].Info)
	c.Data = nil

	// Enable encryption, if supported by the server
	if c.Caps["STARTTLS"] {
		c.StartTLS(nil)
	}

	// Authenticate
	if c.State() == imap.Login {
		c.Login(user, pass)
	}

	// List all top-level mailboxes, wait for the command to finish
	cmd, err = imap.Wait(c.List("", "%"))
	if err != nil {
		fmt.Println(err)
		return
	}

	// Open a mailbox (synchronous command - no need for imap.Wait)
	c.Select("INBOX", true)
	fmt.Println("\nMailbox status:\n", c.Mailbox)

	fmt.Println("Saving messages to mbox file: 'imap.mbox'")
	// fetch all messages
	var set *imap.SeqSet
	if fetchAll {
		set, _ = imap.NewSeqSet("1:*")
	} else {
		set, _ = imap.NewSeqSet(seq)
	}
	// if c.Mailbox.Messages >= 10 {
	// 	set.AddRange(c.Mailbox.Messages-9, c.Mailbox.Messages)
	// } else {
	// 	set.Add("1:*")
	// }
	// cmd, err = c.Fetch(set, "RFC822.HEADER")
	// if err != nil {
	// 	fmt.Println("Error processing:", err)
	// 	return
	// }

	cmd, err = c.Fetch(set, "RFC822.HEADER", "RFC822.TEXT")
	if err != nil {
		fmt.Println(err)
		return
	}

	var i int = 1
	if err := mbox.Open(nil, filename); err != nil {
		fmt.Println(err)
		return
	}
	for cmd.InProgress() {
		c.Recv(-1)
		for _, rsp = range cmd.Data {
			header := imap.AsBytes(rsp.MessageInfo().Attrs["RFC822.HEADER"])
			if msg, _ := mail.ReadMessage(bytes.NewReader(header)); msg != nil {
				body := imap.AsBytes(rsp.MessageInfo().Attrs["RFC822.TEXT"])
				if len(body) > 0 {
					var form mbox.Form
					form.From = msg.Header.Get("Return-path")
					form.Subject = msg.Header.Get("Subject")
					form.Message = string(body)
					mbox.Save(&form)
					fmt.Printf("Message #%v saved to mbox\n", i)
					i++
				}
			}
		}
		cmd.Data = nil
	}
	mbox.Close()

	// Check command completion status
	if rsp, err := cmd.Result(imap.OK); err != nil {
		if err == imap.ErrAborted {
			fmt.Println("Fetch command aborted")
		} else {
			fmt.Println("Fetch error:", rsp.Info)
		}
	}
}
