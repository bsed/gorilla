// Copyright 2011 Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flattener

import (
	"os"
	"reflect"
	"strings"
)

// Item represents an item in a flattened structure.
type Item struct {
	// Full name in dotted notation.
	Path     string
	// Value.
	Value    reflect.Value
	// Parent: a map, struct or slice.
	Parent   reflect.Value
	// Field name if parent is a struct or key if parent is a map.
	Name     string
	// True if parent is a slice, false otherwise.
	Multiple bool
}

// Flatten converts nested structs or maps to an array of Item with paths
// in dotted notation.
//
// See in the documentation the supported types and constraints.
func Flatten(i interface{}) ([]*Item, os.Error) {
	value := recursiveIndirect(reflect.ValueOf(i))
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil, os.NewError("Interface must be a struct.")
	}
	items := make([]*Item, 0)
	flatten(value, value, &items, "", "", false)
	return items, nil
}

func flatten(value, parent reflect.Value, items *[]*Item, path, name string, multiple bool) {
	kind := value.Kind()
	if kind == reflect.Map {
		flattenMap(value, items, path)
	} else if kind == reflect.Slice {
		flattenSlice(value, items, path)
	} else if kind == reflect.Struct {
		flattenStruct(value, items, path)
	} else {
		item := &Item{
			Path:     path,
			Value:    value,
			Parent:   parent,
			Name:     name,
			Multiple: multiple,
		}
		*items = append(*items, item)
	}
}

func flattenMap(value reflect.Value, items *[]*Item, path string) {
	//flatten(mValue, value, items, key(path, mName), key, false)
}

func flattenSlice(value reflect.Value, items *[]*Item, path string) {
	// Don't need to check if the type is supported because it is checked
	// in flattenMap or flattenStruct.
	num := value.Len()
	for i := 0; i < num; i++ {
		flatten(value.Index(i), value, items, path, "", true)
	}
}

func flattenStruct(value reflect.Value, items *[]*Item, path string) {
	num := value.NumField()
	sType := value.Type()
	for i := 0; i < num; i++ {
		fValue := recursiveIndirect(value.Field(i))
		if !isSupportedType(fValue.Type()) {
			continue
		}
		// Use the name defined in the tag, if available.
		field := sType.Field(i)
		fName := field.Tag.Get("schema")
		if fName == "" {
			fName = field.Name
		}
		flatten(fValue, value, items, key(path, fName), field.Name, false)
	}
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

func recursiveIndirect(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	return value
}

func isSupportedType(t reflect.Type) bool {
	// TODO
	return true
}

func key(parts ...string) string {
	s := make([]string, 0)
	for _, part := range parts {
		if part != "" {
			s = append(s, part)
		}
	}
	return strings.Join(s, ".")
}
