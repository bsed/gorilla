// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// ----------------------------------------------------------------------------
// Public interface
// ----------------------------------------------------------------------------

// Load fills a struct with form values.
//
// The first parameter must be a pointer to a struct. The second is a map,
// typically url.Values, http.Request.Form or http.Request.MultipartForm.
//
// This function is capable of filling nested structs recursivelly using map
// keys as "paths" in dotted notation.
//
// See the package documentation for a full explanation of the mechanics.
func Load(i interface{}, data map[string][]string) os.Error {
	v := reflect.ValueOf(i)
	for v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return os.NewError("Interface must be a pointer to struct.")
	}
	rv := v.Elem()
	for path, values := range data {
		if parts := strings.Split(path, "."); len(parts) > 0 {
			loadValue(path, values[:], rv, parts[:])
		}
	}
	return nil
}

// ----------------------------------------------------------------------------
// loader
// ----------------------------------------------------------------------------

// loadValue sets the value for a path in a struct.
//
// - path is the ummodified key for this value, in dotted notation.
//
// - values are the ummodified values to be set.
//
// - rv is the current struct being walked.
//
// - parts are the remaining path parts to be walked until the last
// is reached, when the value is set.
//
// TODO support struct values in maps and slices at some point.
// Currently maps and slices can be of the base types only.
func loadValue(path string, values []string, rv reflect.Value, parts []string) {
	spec, err := defaultStructMap.getOrLoad(rv.Type())
	if err != nil {
		// TODO this should not happen, but what if spec can't be loaded?
		panic("Error loading struct spec.")
	}

	fieldSpec, ok := spec.fields[parts[0]]
	if !ok {
		// Field doesn't exist.
		return
	}

	parts = parts[1:]
	field := setIndirect(rv.FieldByName(fieldSpec.realName))
	kind := field.Kind()
	if (kind == reflect.Struct || kind == reflect.Map) == (len(parts) == 0) {
		// Last part can't be a struct or map.
		// Other parts must be a struct or map.
		return
	}

	var idx string
	if kind == reflect.Map && len(parts) > 0 {
		// Maps must have an index.
		idx = parts[0]
		parts = parts[1:]
	}

	if len(parts) > 0 {
		if kind == reflect.Map || kind == reflect.Slice {
			// Maps and slices must be last part.
			// This may change in the future if we start to support
			// maps or slices of structs.
			return
		}
		// A struct. Move to next part.
		loadValue(path, values[:], field, parts)
	} else {
		// Last part: set the value.
		switch kind {
			case reflect.Bool,
				reflect.Float32, reflect.Float64,
				reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.String,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				field.Set(coerce(values[0], kind))
			case reflect.Map:
				ekind := field.Type().Elem().Kind()
				if field.IsNil() {
					field.Set(reflect.MakeMap(field.Type()))
				}
				field.SetMapIndex(reflect.ValueOf(idx), coerce(values[0], ekind))
			case reflect.Slice:
				ekind := field.Type().Elem().Kind()
				slice := reflect.MakeSlice(field.Type(), 0, 0)
				for _, value := range values {
					slice = reflect.Append(slice, coerce(value, ekind))
				}
				field.Set(slice)
		}
	}
}

// coerce coerces base types from a string to a reflect.Value of a given kind.
func coerce(value string, kind reflect.Kind) (rv reflect.Value) {
	switch kind {
		case reflect.Bool:
			v, _ := strconv.Atob(value)
			rv = reflect.ValueOf(v)
		case reflect.Float32:
			v, _ := strconv.Atof32(value)
			rv = reflect.ValueOf(v)
		case reflect.Float64:
			v, _ := strconv.Atof64(value)
			rv = reflect.ValueOf(v)
		case reflect.Int:
			v, _ := strconv.Atoi(value)
			rv = reflect.ValueOf(v)
		case reflect.Int8:
			v, _ := strconv.Atoi(value)
			rv = reflect.ValueOf(int8(v))
		case reflect.Int16:
			v, _ := strconv.Atoi(value)
			rv = reflect.ValueOf(int16(v))
		case reflect.Int32:
			v, _ := strconv.Atoi(value)
			rv = reflect.ValueOf(int32(v))
		case reflect.Int64:
			v, _ := strconv.Atoi64(value)
			rv = reflect.ValueOf(v)
		case reflect.String:
			rv = reflect.ValueOf(value)
		case reflect.Uint:
			v, _ := strconv.Atoui(value)
			rv = reflect.ValueOf(v)
		case reflect.Uint8:
			v, _ := strconv.Atoui(value)
			rv = reflect.ValueOf(uint8(v))
		case reflect.Uint16:
			v, _ := strconv.Atoui(value)
			rv = reflect.ValueOf(uint16(v))
		case reflect.Uint32:
			v, _ := strconv.Atoui(value)
			rv = reflect.ValueOf(uint32(v))
		case reflect.Uint64:
			v, _ := strconv.Atoui64(value)
			rv = reflect.ValueOf(v)
	}
	return
}

// ----------------------------------------------------------------------------
// structMap
// ----------------------------------------------------------------------------

// Internal map of cached struct specs.
var defaultStructMap = newStructMap()

// structMap caches parsed structSpec's keyed by package+name.
type structMap struct {
	specs map[string]*structSpec
	mutex sync.RWMutex
}

// newStructMap returns a new structMap instance.
func newStructMap() *structMap {
	return &structMap{
		specs: make(map[string]*structSpec),
	}
}

// getByType returns a cached structSpec given a struct type.
//
// It returns nil if the type argument is not a reflect.Struct.
func (m *structMap) getByType(t reflect.Type) (spec *structSpec) {
	if m.specs != nil && t.Kind() == reflect.Struct {
		m.mutex.RLock()
		spec = m.specs[getStructId(t)]
		m.mutex.RUnlock()
	}
	return
}

// getOrLoad returns a cached structSpec, loading and caching it if needed.
//
// It returns nil if the passed type is not a struct.
func (m *structMap) getOrLoad(t reflect.Type) (spec *structSpec, err os.Error) {
	if spec = m.getByType(t); spec != nil {
		return spec, nil
	}

	// Lock it for writes until the new type is loaded.
	m.mutex.Lock()
	loaded := make([]string, 0)
	if spec, err = m.load(t, &loaded); err != nil {
		// Roll back loaded structs.
		for _, v := range loaded {
			m.specs[v] = nil, false
		}
		return
	}
	m.mutex.Unlock()

	return
}

// load caches parsed struct metadata.
//
// It is an internal function used by getOrLoad and can't be called directly
// because a write lock is required.
//
// The loaded argument is the list of keys to roll back in case of error.
func (m *structMap) load(t reflect.Type, loaded *[]string) (spec *structSpec, err os.Error) {
   defer func() {
		if r := recover(); r != nil {
			err = r.(os.Error)
		}
	}()

	if t.Kind() != reflect.Struct {
		return nil, os.NewError("Not a struct.")
	}

	structId := getStructId(t)
	spec = &structSpec{fields: make(map[string]*structFieldSpec)}
	m.specs[structId] = spec
	*loaded = append(*loaded, structId)

	var toLoad reflect.Type
	uniqueNames := make([]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		ft := field.Type
		if !isSupportedType(ft) {
			continue
		}

		toLoad = nil
		switch ft.Kind() {
			case reflect.Map, reflect.Slice:
				et := ft.Elem()
				if et.Kind() == reflect.Struct {
					toLoad = et
				}
			case reflect.Struct:
				toLoad = ft
		}

		if toLoad != nil {
			// Load nested struct.
			structId = getStructId(toLoad)
			if m.specs[structId] == nil {
				if _, err = m.load(toLoad, loaded); err != nil {
					return nil, err
				}
			}
		}

		// Use the name defined in the tag, if available.
		name := field.Tag.Get("schema-name")
		if name == "" {
			name = field.Name
		}

		// The name must be unique for the struct.
		for _, uniqueName := range uniqueNames {
			if name == uniqueName {
				return nil, os.NewError("Field names and name tags in a struct must be unique.")
			}
		}
		uniqueNames[i] = name

		// Finally, set the field.
		spec.fields[name] = &structFieldSpec{
			name:     name,
			realName: field.Name,
		}
	}
	return
}

// ----------------------------------------------------------------------------
// structSpec
// ----------------------------------------------------------------------------

// structSpec stores information from a parsed struct.
//
// It is used to fill a struct with values from a multi-map, checking if keys
// in dotted notation can be translated to a struct field and executing
// filters and conversions.
type structSpec struct {
	fields map[string]*structFieldSpec
}

// ----------------------------------------------------------------------------
// structFieldSpec
// ----------------------------------------------------------------------------

// structFieldSpec stores information from a parsed struct field.
type structFieldSpec struct {
	// Name defined in the field tag, or the real field name.
	name     string
	// Real field name as defined in the struct.
	realName string
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// getStructId returns an ID for a struct: package name + "." + struct name.
func getStructId(t reflect.Type) string {
	return t.PkgPath() + "." + t.Name()
}

// isSupportedType returns true for supported field types.
func isSupportedType(t reflect.Type) (b bool) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if isSupportedBaseType(t) {
		b = true
	} else {
		switch t.Kind() {
			case reflect.Slice:
				// Only []anyOfTheBaseTypes.
				b = isSupportedBaseType(t.Elem())
			case reflect.Struct:
				b = true
			case reflect.Map:
				// Only map[string]anyOfTheBaseTypes.
				if t.Key().Kind() == reflect.String && isSupportedBaseType(t.Elem()) {
					b = true
				}
		}
	}
	return
}

// isSupportedBaseType returns true for supported base field types.
//
// Only base types can be used in maps/slices values.
func isSupportedBaseType(t reflect.Type) (b bool) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.Kind() {
		case reflect.Bool,
			reflect.Float32, reflect.Float64,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.String,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			b = true
	}
	return
}

// setIndirect resolves a pointer to value, setting it recursivelly if needed.
func setIndirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			ptr := reflect.New(v.Type().Elem())
			v.Set(ptr)
			v = ptr
		}
		v = reflect.Indirect(v)
	}
	return v
}
