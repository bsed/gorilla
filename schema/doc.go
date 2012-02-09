// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gorilla/schema fills a struct with form values.

The basic usage is really simple. Given this struct:

	type Person struct {
		Name  string
		Phone string
	}

...we can fill it passing a map to the Load() function:

	values := map[string][]string{
		"Name":  {"John"},
		"Phone": {"999-999-999"},
	}
	person := new(Person)
	decoder := NewDecoder()
	decoder.Decode(person, values)

This is just a simple example and it doesn't make a lot of sense to create
the map manually. Typically it will come from a http.Request object and
will be of type url.Values: http.Request.Form or http.Request.MultipartForm.

Note: it is a good idea to set a StructLoader instance as a package global,
because it caches meta-data about structs, and a instance can be shared safely:

	var decoder = NewDecoder()

To define custom names for fields, use a struct tag "schema". To not populate
certain fields, use a dash as the name and it will be ignored:

	type Person struct {
		Name  string `schema:"name"`  // custom name
		Phone string `schema:"phone"` // custom name
		Admin bool   `schema:"-"`     // this field is never set
	}

The supported field types in the destination struct are:

	* bool
	* float variants (float32, float64)
	* int variants (int, int8, int16, int32, int64)
	* string
	* uint variants (uint, uint8, uint16, uint32, uint64)
	* structs
	* slices of any of the above types
	* a pointer to one of the above types

Non-supported types are simply ignored.

To fill nested structs, keys must use a dotted notation as the "path" for the
field. So for example, to fill the struct Person below:

	type Phone struct {
		Label  string
		Number string
	}

	type Person struct {
		Name  string
		Phone Phone
	}

...the source map must have the keys "Name", "Phone.Label" and "Phone.Number".
This means that an HTML form to fill a Person struct must look like this:

	<form>
		<input type="text" name="Name">
		<input type="text" name="Phone.Label">
		<input type="text" name="Phone.Number">
	</form>

Single values are filled using the first value for a key from the source map.
Slices are filled using all values for a key from the source map. So to fill
a Person with multiple Phone values, like:

	type Person struct {
		Name   string
		Phones []Phone
	}

...an HTML form that accepts three Phone values would look like this:

	<form>
		<input type="text" name="Name">
		<input type="text" name="Phones.0.Label">
		<input type="text" name="Phones.0.Number">
		<input type="text" name="Phones.1.Label">
		<input type="text" name="Phones.1.Number">
		<input type="text" name="Phones.2.Label">
		<input type="text" name="Phones.2.Number">
	</form>

Notice that only for slices of structs the slice index is required.
This is needed for disambiguation: if the nested struct also has a slice
field, we could not represent it.
*/
package schema
