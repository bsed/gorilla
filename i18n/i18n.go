// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package i18n

// Translator types can return translations for messages and plurals.
type Translator interface {
	// Message returns a translation for the given message key.
	// Extra arguments can be passed to format the translation
	// using fmt.Sprintf().
	Message(key string, a ...interface{}) string
	// Plural returns a plural translation for the given message key and count.
	// Extra arguments can be passed to format the translation
	// using fmt.Sprintf().
	Plural(key string, count int, a ...interface{}) string
}
