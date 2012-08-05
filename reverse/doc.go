// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gorilla/reverse produces reversible regular expressions that can be
used to generate URLs for a regexp-based mux.

For example, let's start compiling a simple regexp:

	regexp, err := reverse.Compile(`/foo/1(\d+)3`)

Now we can call regexp.Revert() passing variables to fill the capturing groups:

	// url is "/foo/123".
	url, err := regexp.Revert([]string{"2"}, nil)

Non-capturing groups are ignored, but named capturing groups can be filled
using a map (or a slice to treat them as positional variables):

	regexp, err := reverse.Compile(`/foo/1(?P<two>\d+)3`)
	if err != nil {
		panic(err)
	}
	// url is "/foo/123".
	url, err := re.Revert(nil, map[string]string{"two": "2"})

There are a few limitations that can't be changed:

1. Nested capturing groups are ignored; only the outermost groups become
a placeholder. So in `1(\d+([a-z]+))3` there is only one placeholder
although there are two capturing groups: re.Revert([]string{"2", "a"}, nil)
results in "123" and not "12a3".

2. Literals inside capturing groups are ignored; the whole group becomes
a placeholder.
*/
package reverse
