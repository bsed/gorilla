package hello

import (
	"fmt"
	"net/http"

	"code.google.com/p/gorilla/appengine/sessions"
	"code.google.com/p/gorilla/mux"
)

var router = new(mux.Router)
var dStore = sessions.NewDatastoreStore("", []byte("my-secret-key"))
	//[]byte("1234567890123456"))
var mStore = sessions.NewMemcacheStore("", []byte("my-secret-key"))
	//[]byte("1234567890123456"))

func init() {
	// Register a couple of routes.
	router.HandleFunc("/", homeHandler).Name("home")
	router.HandleFunc("/{salutation}/{name}", helloHandler).Name("hello")
	router.HandleFunc("/memcache-session", memcacheSessionHandler).Name("memcache-session")
	router.HandleFunc("/datastore-session", datastoreSessionHandler).Name("datastore-session")
	// Send all incoming requests to router.
	http.Handle("/", router)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	url1, _ := router.GetRoute("hello").URL("salutation", "hello", "name", "world")
	url2, _ := router.GetRoute("datastore-session").URL()
	url3, _ := router.GetRoute("memcache-session").URL()
	fmt.Fprintf(w, "Try a <a href='%s'>hello</a>. Or a <a href='%s'>datastore</a> or <a href='%s'>memcache</a> session.", url1, url2, url3)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	vars := mux.Vars(r)
	fmt.Fprintf(w, "%s, %s!", vars["salutation"], vars["name"])
}

func datastoreSessionHandler(w http.ResponseWriter, r *http.Request) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := dStore.Get(r, "session-name")
	if value, ok := session.Values["foo"]; ok {
		// Show a previously saved value.
		fmt.Fprintf(w, `Value for datastore session["foo"] is "%s".`, value)
	} else {
		// Set a value.
		session.Values["foo"] = "bar"
		// Save it.
		err := session.Save(r, w)
		fmt.Fprintf(w, `No value found for datastore session["foo"]. Saved a new one (errors: %v).`, err)
	}
}

func memcacheSessionHandler(w http.ResponseWriter, r *http.Request) {
	// Get a session. We're ignoring the error resulted from decoding an
	// existing session: Get() always returns a session, even if empty.
	session, _ := mStore.Get(r, "session-name")
	if value, ok := session.Values["foo"]; ok {
		// Show a previously saved value.
		fmt.Fprintf(w, `Value for memcache session["foo"] is "%s".`, value)
	} else {
		// Set a value.
		session.Values["foo"] = "bar"
		// Save it.
		err := session.Save(r, w)
		fmt.Fprintf(w, `No value found for memcache session["foo"]. Saved a new one (errors: %v).`, err)
	}
}
