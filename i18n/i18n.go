// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package i18n

// Translator types return translations for messages and plurals.
type Translator interface {
	// Get returns a translation for the given key.
	// Extra arguments can be passed to format the translation
	// using fmt.Sprintf().
	Get(key string, a ...interface{}) string
	// GetPlural returns a plural translation for the given key and count.
	// Extra arguments can be passed to format the translation
	// using fmt.Sprintf().
	GetPlural(key string, count int, a ...interface{}) string
}
