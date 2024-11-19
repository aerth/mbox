package webmbox

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"

	"github.com/aerth/mbox"
	"github.com/microcosm-cc/bluemonday"
)

// this file contains functions that will be removed soon

var FieldsBlacklist = map[string]bool{
	"cosgo":           true,
	"captchaid":       true,
	"captchasolution": true,
	"captcha":         true,
}
var FieldsWhitelist = map[string]bool{}

// ParseQuery returns a mbox.Form from url.Values
func ParseQuery(query url.Values) mbox.Writable {
	p := bluemonday.StrictPolicy()
	form := new(mbox.Form)
	additionalFields := ""
	name, email := "", ""
	for k, v := range query {
		k = strings.ToLower(k)
		if k == "name" {
			name = v[0]
			if email != "" {
				form.From = name + " <" + email + ">"
			} else {
				form.From = name
			}
		} else if k == "email" {
			email = v[0]
			if name != "" {
				form.From = name + " <" + email + ">"
			} else {
				form.From = email
			}
		} else if k == "subject" {
			form.Subject = v[0]
			form.Subject = p.Sanitize(form.Subject)
		} else if k == "message" {
			form.Message = k + ": " + v[0] + "<br>\n"
			form.Message = p.Sanitize(form.Message)
		} else if k != "cosgo" && k != "captchaid" && k != "captchasolution" && !FieldsBlacklist[k] {
			if (FieldsWhitelist[k] || len(FieldsWhitelist) == 0) && !FieldsBlacklist[k] {
				additionalFields = additionalFields + k + ": " + v[0] + "<br>\n"
			}
		}
	}
	if form.Subject == "" || form.Subject == " " {
		form.Subject = "[New Message]"
	}
	if additionalFields != "" {
		if form.Message == "" {
			form.Message = form.Message + "Message:\n<br>" + p.Sanitize(additionalFields)
		} else {
			form.Message = form.Message + "\n<br>Additional:\n<br>" + p.Sanitize(additionalFields)
		}
	}

	return form
}

// rel2real Relative to Real path name
func rel2real(file string) (realpath string) {
	pathdir, _ := path.Split(file)
	if pathdir == "" {
		realpath, _ = filepath.Abs(file)
	} else {
		realpath = file
	}
	return realpath
}
