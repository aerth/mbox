# mbox file writer library

super simple mbox file writer library for go

[![godev](https://pkg.go.dev/badge/github.com/aerth/mbox)](https://pkg.go.dev/github.com/aerth/mbox#pkg-index)
[![Go Report Card](https://goreportcard.com/badge/github.com/aerth/mbox)](https://goreportcard.com/report/github.com/aerth/mbox)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE.md)
[![New Issue](https://img.shields.io/badge/new-issue-blue.svg)](https://github.com/aerth/mbox/issues/new)
![Go Version](https://img.shields.io/github/go-mod/go-version/aerth/mbox)
![Latest Release](https://img.shields.io/github/v/release/aerth/mbox)

## more info

See examples in the 'examples' directory.

And test files (mbox_test.go, ext_test.go) for more examples.

Report issues: [github.com/aerth/mbox/issues](https://github.com/aerth/mbox/issues)

New: Now with support for age-encryption (set mbox.AgeRecipient to a public key string to activate)

### example: write message to any io.Writer
    
```go
func example() {
    var form mbox.Form
    form.From = "Alice <alice@localhost>"
    form.Subject = "As seen on TV!!!"
    form.Message = "Bob, this really works!"
    form.WriteTo(os.Stdout)
}
```

### example: save an "email" to an mbox file

```go
package main

import (
	"context"
	"os"

	"github.com/aerth/mbox"
)

func example() {
	mbox.Destination = "me@localhost" // optional, for ALL 'To' fields

	// Build the email
	var form mbox.Form
	form.From = "Alice <alice@localhost>"
	form.Subject = "As seen on TV!!!"
	form.Message = "Bob, this really works!"

	// Use global mbox, choose file name
	if true {
		mbox.Open(context.Background(), "my.mbox")
		// Save message to mailbox. If concurrent writes will happen, use a mutex.
		mbox.Save(&form)
		mbox.Save(&form)
		mbox.Save(&form)
		mbox.Close() // close after all writes are done
	} else {
		// alternatively, write a Form object directly to an io.Writer
		form.WriteTo(os.Stdout)
	}
}
```

### example: reading the mbox file with mutt

```bash
mutt -R -f my.mbox
```
