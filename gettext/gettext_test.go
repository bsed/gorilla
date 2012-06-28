// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gettext

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
)

func decode(value []byte) ([]byte, error) {
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(value)))
	b, err := base64.StdEncoding.Decode(decoded, value)
	if err != nil {
		return nil, err
	}
	return decoded[:b], nil
}

func newFile(testName string, t *testing.T) (f *os.File) {
	// Use a local file system, not NFS.
	// On Unix, override $TMPDIR in case the user
	// has it set to an NFS-mounted directory.
	dir := ""
	if runtime.GOOS != "windows" {
		dir = "/tmp"
	}
	f, err := ioutil.TempFile(dir, "_Go_"+testName)
	if err != nil {
		t.Fatalf("open %s: %s", testName, err)
	}
	return
}

// From Python's gettext tests
var gnuMoData = `3hIElQAAAAAGAAAAHAAAAEwAAAALAAAAfAAAAAAAAACoAAAAFQAAAKkAAAAjAAAAvwAAAKEAAADj
AAAABwAAAIUBAAALAAAAjQEAAEUBAACZAQAAFgAAAN8CAAAeAAAA9gIAAKEAAAAVAwAABQAAALcD
AAAJAAAAvQMAAAEAAAADAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAABQAAAAYAAAACAAAAAFJh
eW1vbmQgTHV4dXJ5IFlhY2gtdABUaGVyZSBpcyAlcyBmaWxlAFRoZXJlIGFyZSAlcyBmaWxlcwBU
aGlzIG1vZHVsZSBwcm92aWRlcyBpbnRlcm5hdGlvbmFsaXphdGlvbiBhbmQgbG9jYWxpemF0aW9u
CnN1cHBvcnQgZm9yIHlvdXIgUHl0aG9uIHByb2dyYW1zIGJ5IHByb3ZpZGluZyBhbiBpbnRlcmZh
Y2UgdG8gdGhlIEdOVQpnZXR0ZXh0IG1lc3NhZ2UgY2F0YWxvZyBsaWJyYXJ5LgBtdWxsdXNrAG51
ZGdlIG51ZGdlAFByb2plY3QtSWQtVmVyc2lvbjogMi4wClBPLVJldmlzaW9uLURhdGU6IDIwMDAt
MDgtMjkgMTI6MTktMDQ6MDAKTGFzdC1UcmFuc2xhdG9yOiBKLiBEYXZpZCBJYsOhw7FleiA8ai1k
YXZpZEBub29zLmZyPgpMYW5ndWFnZS1UZWFtOiBYWCA8cHl0aG9uLWRldkBweXRob24ub3JnPgpN
SU1FLVZlcnNpb246IDEuMApDb250ZW50LVR5cGU6IHRleHQvcGxhaW47IGNoYXJzZXQ9aXNvLTg4
NTktMQpDb250ZW50LVRyYW5zZmVyLUVuY29kaW5nOiBub25lCkdlbmVyYXRlZC1CeTogcHlnZXR0
ZXh0LnB5IDEuMQpQbHVyYWwtRm9ybXM6IG5wbHVyYWxzPTI7IHBsdXJhbD1uIT0xOwoAVGhyb2F0
d29iYmxlciBNYW5ncm92ZQBIYXkgJXMgZmljaGVybwBIYXkgJXMgZmljaGVyb3MAR3V2ZiB6YnFo
eXIgY2ViaXZxcmYgdmFncmVhbmd2YmFueXZtbmd2YmEgbmFxIHlicG55dm1uZ3ZiYQpmaGNjYmVn
IHNiZSBsYmhlIENsZ3ViYSBjZWJ0ZW56ZiBvbCBjZWJpdnF2YXQgbmEgdmFncmVzbnByIGdiIGd1
ciBUQUgKdHJnZ3JrZyB6cmZmbnRyIHBuZ255YnQgeXZvZW5lbC4AYmFjb24Ad2luayB3aW5rAA==`

func TestReadMO(t *testing.T) {
	equalString := func(s1, s2 string) {
		if s1 != s2 {
			t.Errorf("Expected %q, got %q.", s1, s2)
		}
	}

	b, err := decode([]byte(gnuMoData))
	if err != nil {
		t.Fatal(err)
	}
	c := NewCatalog()
	if err := c.ReadMO(bytes.NewReader(b)); err != nil {
		t.Fatal(err)
	}

	// gettext
	equalString(c.Gettext("albatross"), "albatross")
	equalString(c.Gettext("mullusk"), "bacon")
	equalString(c.Gettext("Raymond Luxury Yach-t"), "Throatwobbler Mangrove")
	equalString(c.Gettext("nudge nudge"), "wink wink")
	equalString(c.Gettext("There is %s file"), "Hay %s fichero")
	// ngettext
	equalString(c.Ngettext("There is %s file", "There is %s file", 1), "Hay %s fichero")
	equalString(c.Ngettext("There is %s file", "There are %s files", 2), "Hay %s ficheros")
}

func TestWriteMO(t *testing.T) {
	equalString := func(s1, s2 string) {
		if s1 != s2 {
			t.Errorf("Expected %q, got %q.", s1, s2)
		}
	}

	b, err := decode([]byte(gnuMoData))
	if err != nil {
		t.Fatal(err)
	}
	c := NewCatalog()
	if err := c.ReadMO(bytes.NewReader(b)); err != nil {
		t.Fatal(err)
	}

	f1 := newFile("testWriteMO", t)
	defer f1.Close()
	c.WriteMO(f1)

	f2, err := os.Open(f1.Name())
	defer f2.Close()
	if err != nil {
		t.Fatal(err)
	}
	c2 := NewCatalog()
	if err := c2.ReadMO(f2); err != nil {
		t.Fatal(err)
	}

	// gettext
	equalString(c2.Gettext("albatross"), "albatross")
	equalString(c2.Gettext("mullusk"), "bacon")
	equalString(c2.Gettext("Raymond Luxury Yach-t"), "Throatwobbler Mangrove")
	equalString(c2.Gettext("nudge nudge"), "wink wink")
	equalString(c2.Gettext("There is %s file"), "Hay %s fichero")
	// ngettext
	equalString(c2.Ngettext("There is %s file", "There is %s file", 1), "Hay %s fichero")
	equalString(c2.Ngettext("There is %s file", "There are %s files", 2), "Hay %s ficheros")
}
