// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"os"
	"reflect"
	"strings"
)

// Item represents an item in a flattened structure.
type Item struct {
	// Full path in dotted notation.
	Path   string
	// Key for maps, field name for structs or zero for slices.
	Name   string
	// Value.
	Value  reflect.Value
	// Parent: a map, struct or slice.
	Parent reflect.Value
}

// Flatten converts nested structs or maps to an array of Item with paths
// in dotted notation.
//
// See in the documentation the supported types and constraints.
func Flatten(i interface{}) ([]*Item, os.Error) {
	value := recursiveIndirect(reflect.ValueOf(i))
	items := make([]*Item, 0)
	kind := value.Kind()
	if kind == reflect.Map || kind == reflect.Struct {
		if err := flatten(&items, value, value, "", ""); err != nil {
			return nil, err
		}
	} else {
		return nil, os.NewError("Interface must be a map or struct.")
	}
	return items[:], nil
}

func flatten(items *[]*Item, value, parent reflect.Value, path, name string) os.Error {
	switch value.Kind() {
	case reflect.Map:
		return flattenMap(items, value, path)
	case reflect.Slice:
		return flattenSlice(items, value, path)
	case reflect.Struct:
		return flattenStruct(items, value, path)
	}
	item := &Item{
		Path:   path,
		Name:   name,
		Value:  value,
		Parent: parent,
	}
	*items = append(*items, item)
	return nil
}

func flattenMap(items *[]*Item, value reflect.Value, path string) os.Error {
	// Only map[string]anyOfTheBaseTypes.
	stringKey := value.Type().Key().Kind() == reflect.String
	if !stringKey || !isSupportedBasicType(value.Type().Elem()) {
		return os.NewError("Map must be map[string]SupportedTypes.")
	}
	keys := value.MapKeys()
	for _, k := range keys {
		mKey := k.String()
		mValue := recursiveIndirect(value.MapIndex(k))
		flatten(items, mValue, value, key(path, mKey), mKey)
	}
	return nil
}

func flattenSlice(items *[]*Item, value reflect.Value, path string) os.Error {
	// Don't need to check if the type is supported because it is checked
	// in flattenMap or flattenStruct.
	num := value.Len()
	for i := 0; i < num; i++ {
		sValue := recursiveIndirect(value.Index(i))
		flatten(items, sValue, value, path, "")
	}
	return nil
}

func flattenStruct(items *[]*Item, value reflect.Value, path string) os.Error {
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
		flatten(items, fValue, value, key(path, fName), field.Name)
	}
	return nil
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

// isSupportedType returns true for supported field types.
func isSupportedType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if isSupportedBasicType(t) {
		return true
	} else {
		switch t.Kind() {
		case reflect.Slice:
			// Only []anyOfTheBaseTypes.
			return isSupportedBasicType(t.Elem())
		case reflect.Map:
			// Only map[string]anyOfTheBaseTypes.
			stringKey := t.Key().Kind() == reflect.String
			if stringKey && isSupportedBasicType(t.Elem()) {
				return true
			}
		}
	}
	return false
}

// isSupportedBasicType returns true for supported basic field types.
//
// Only basic types can be used in maps/slices values.
func isSupportedBasicType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Bool,
		reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64,
		reflect.String,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64,
		reflect.Struct:
		return true
	}
	return false
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
