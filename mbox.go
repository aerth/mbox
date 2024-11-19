// The MIT License (MIT)
//
// Copyright (c) 2016 aerth
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//

// Package mbox saves a form to a local .mbox file (opengpg option)
/*

Usage of mbox library is as follows:

Define mbox.Destination variable in your program

Accept an email, populate the mbox.Form struct like this:
	mbox.From = "joe"
	mbox.Email = "joe@blowtorches.info
	mbox.Message = "hello world"
	mbox.Subject = "re: hello joe"
	mbox.Save()


*/
package mbox

import (

	//	"fmt"
	"io"
	"time"
	// email validation
	// input sanitization
)

// Form is a single email. No Attachments yet.
type Form struct {
	From     string    // may be empty "name <email>" format
	Subject  string    // may be empty
	Message  string    // the message string
	Sent     time.Time // optional, when the message was sent
	Received time.Time // optional, when the message was received (automatically set)
	Body     []byte    // experimental: possible future use, attachments?
}

var Version = "0.0.2-MIT"

// ValidationLevel is the level of email validation during Loop() (see Normalize)
// 0 = none, 1 = normalize, 2 = validate format, 3 = validate format and host
var ValidationLevel = 1

// Writable is an interface for writing to a file
// Default is 'Form' type, but you can implement your own for custom behavior
type Writable interface {
	// WriteTo writes the single form to a file/stream, and returns the number of bytes written and an error
	WriteTo(w io.Writer) (n int64, err error)
}

var _ Writable = (*Form)(nil)
