package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"filippo.io/age"
	"github.com/aerth/mbox"
)

var mboxname = "my.mbox"

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	server := &http.Server{
		Handler: http.HandlerFunc(Handler),
		Addr:    ":8080",
	}
	inputfile := ""
	age_recipient := ""
	flag.StringVar(&server.Addr, "addr", "127.0.0.1:8080", "address to listen on")
	flag.StringVar(&mbox.Destination, "dest", mbox.Destination, "destination email address (optional)")
	flag.StringVar(&mboxname, "mbox", mboxname, "mbox filename")
	flag.StringVar(&inputfile, "html", inputfile, "path to form html file (optional, - for stdin)")
	flag.StringVar(&age_recipient, "age", age_recipient, "age recipient public key (optional, requires custom encrypted mbox reader)")
	flag.Parse()
	if age_recipient != "" {
		// quick check to see if the recipient is valid
		_, err := age.ParseX25519Recipient(age_recipient)
		if err != nil {
			log.Printf("invalid age recipient: %v", err)
			log.Printf("generate with age-keygen (https://filippo.io/age)")
			os.Exit(1)
		}
		mbox.AgeRecipient = age_recipient
	}
	if inputfile == "-" {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal("reading stdin:", err)
		}
		Formpage = b
	} else if inputfile == "x" || inputfile == "none" {
		Formpage = nil
	} else if inputfile != "" {
		b, err := os.ReadFile(inputfile)
		if err != nil {
			log.Fatal("reading file:", err)
		}
		Formpage = b
	}
	println("listening on", server.Addr)
	println("example: curl -d 'name=me&email=me@localhost&subject=hello&message=world' http://localhost:8080/")
	println("or use json: curl -H 'Content-Type: application/json' -d '{\"from\":\"me@localhost\",\"subject\":\"hello\",\"message\":\"world\"}' http://localhost:8080/")
	if err := server.ListenAndServe(); err != nil {
		println(err.Error())
		os.Exit(1)
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" && r.Header.Get("Content-Type") == "application/json" {
		HandleMboxJsonApi(w, r)
		return
	} else if r.Method == "POST" {
		HandleMboxForm(w, r)
		return
	}
	if len(Formpage) != 0 {
		w.Write(Formpage)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func HandleMboxJsonApi(w http.ResponseWriter, r *http.Request) {
	var msg mbox.Form
	if mbox.MailWriteCloser == nil {
		if err := mbox.Open(nil, mboxname); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error opening mbox file: %v", err)
			return
		}
	}
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error decoding json: %v", err)
		return
	}
	if msg.From == "" && msg.Message == "" && msg.Subject == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("empty message")
		return
	}
	err := mbox.Save(&msg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error saving message: %v", err)
		return
	}
	log.Printf("message received (json, %d bytes)", len(msg.Message))
}

func HandleMboxForm(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("error parsing form: %v", err)
		return
	}
	name, email := r.FormValue("name"), r.FormValue("email")
	subject, message := r.FormValue("subject"), r.FormValue("message")

	if strings.TrimSpace(message) == "" {
		HandleMboxJsonApi(w, r)
		return
	}
	if mbox.MailWriteCloser == nil {
		if err := mbox.Open(nil, mboxname); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error opening mbox file: %v", err)
			return
		}
	}

	msg := mbox.NewMessage(name, email, subject, message)
	err = mbox.Save(&msg)
	if err != nil {
		log.Printf("error saving message: %v", err)
		http.Error(w, "error saving message", http.StatusInternalServerError)
		return
	} else {
		log.Printf("message received (form, %d bytes)", len(msg.Message))
		http.Redirect(w, r, "/?sent", http.StatusFound)
	}

}

var Formpage = []byte(`<html>
  <form method="POST">
    Your Name: <input name="name"><br>
    Your Email: <input name="email"><br>
    Subject: <input name="subject"><br>
    Message: <input name="message"><br><br>
    <input type="submit" value="send mail">
  </form>
  </html>
  `)
