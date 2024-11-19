package main

import "github.com/aerth/mbox"

func main() {

	mbox.ValidationLevel = 1 // see ValidationLevels
	mbox.Destination = "me@localhost"

	// Choose file name
	mbox.Open(nil, "my.mbox")

	// Build the email
	var form mbox.Form
	form.From = "notify <notify@localhost>"
	form.Subject = "As seen on TV!!!"
	form.Message = "This really works!"

	// Save message to mailbox
	if mbox.Save(&form) == nil {
		print("Saved to 'my.mbox'\n")
	}

}
