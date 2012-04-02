// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gorilla/pat is a request router and dispatcher with a pat-like
interface. It is an alternative to gorilla/mux and showcases the power of
its API. Package pat is documented at:

	http://gopkgdoc.appspot.com/pkg/github.com/bmizerany/pat

Let's start registering a couple of URL paths and handlers:

	func main() {
		r := pat.New()
		r.Get("/", http.HandlerFunc(HomeHandler))
		r.Get("/products", http.HandlerFunc(ProductsHandler))
		r.Get("/articles", http.HandlerFunc(ArticlesHandler))
		http.Handle("/", r)
	}

Here we register three routes mapping URL paths to handlers. This is
equivalent to how http.HandleFunc() works: if an incoming GET request matches
one of the paths, the corresponding handler is called passing
(http.ResponseWriter, *http.Request) as parameters.

Paths can have variables. They are defined using the format {name} or
{name:pattern}. If a regular expression pattern is not defined, the matched
variable will be anything until the next slash. For example:

	r := mux.NewRouter()
	r.Get("/products/{key}", http.HandlerFunc(ProductHandler))
	r.Get("/articles/{category}/", http.HandlerFunc(ArticlesCategoryHandler))
	r.Get("/articles/{category}/{id:[0-9]+}", http.HandlerFunc(ArticleHandler))

The names are used to create a map of route variables which are stored in the
URL query, prefixed by a colon:

	category := req.URL.Query().Get(":category")

As in the gorilla/mux package, other matchers can be added to the registered
routes, and URLs can be built as well. Check the mux documentation for more
details:

	http://gorilla-web.appspot.com/pkg/gorilla/mux/
*/
package pat
