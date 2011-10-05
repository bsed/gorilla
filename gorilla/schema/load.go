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

// SchemaError stores global errors and validation errors for field values.
type SchemaError struct {
}

func (e *SchemaError) SetGlobalError(msg string) {
}

func (e *SchemaError) SetFieldError(key string, index int, msg string) {
}

func (e *SchemaError) String() string {
	return ""
}

// Load fills a struct with form values.
//
// The first parameter must be a pointer to a struct. The second is a map,
// typically url.Values, http.Request.Form or http.Request.MultipartForm.
//
// This function is capable of filling nested structs recursivelly using map
// keys as "paths" in dotted notation.
//
// See the package documentation for a full explanation of the mechanics.
func Load(i interface{}, data map[string][]string) *SchemaError {
	err := &SchemaError{}
	val := reflect.ValueOf(i)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		err.SetGlobalError("Interface must be a pointer to struct.")
	} else {
		rv := val.Elem()
		for path, values := range data {
			parts := strings.Split(path, ".")
			loadValue(rv, values, parts, path, err)
		}
	}
	return err
}

// ----------------------------------------------------------------------------
// loader
// ----------------------------------------------------------------------------

// loadValue sets the value for a path in a struct.
//
// - rv is the current struct being walked.
//
// - values are the ummodified values to be set.
//
// - parts are the remaining path parts to be walked.
//
// TODO support struct values in maps and slices at some point.
// Currently maps and slices can be of the basic types only.
func loadValue(rv reflect.Value, values, parts []string, key string,
	err *SchemaError) {
	spec, error := defaultStructMap.getOrLoad(rv.Type())
	if error != nil {
		// Struct spec could not be loaded.
		err.SetGlobalError(error.String())
		return
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
		// Last part can't be a struct or map. Others must be a struct or map.
		return
	}

	var idx string
	if kind == reflect.Map {
		// Get map index.
		idx = parts[0]
		parts = parts[1:]
		if len(parts) > 0 {
			// Last part must be the map index.
			return
		}
	}

	if len(parts) > 0 {
		// A struct. Move to next part.
		loadValue(field, values, parts, key, err)
		return
	}

	// Last part: set the value.
	var value reflect.Value
	switch kind {
	case reflect.Bool,
		reflect.Float32, reflect.Float64,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.String,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		value = coerce(kind, values[0], key, 0, err)
		if value.IsValid() {
			field.Set(value)
		}
	case reflect.Map:
		ekind := field.Type().Elem().Kind()
		if field.IsNil() {
			field.Set(reflect.MakeMap(field.Type()))
		}
		value = coerce(ekind, values[0], key, 0, err)
		if value.IsValid() {
			field.SetMapIndex(reflect.ValueOf(idx), value)
		}
	case reflect.Slice:
		ekind := field.Type().Elem().Kind()
		slice := reflect.MakeSlice(field.Type(), 0, 0)
		for k, v := range values {
			value = coerce(ekind, v, key, k, err)
			if value.IsValid() {
				slice = reflect.Append(slice, value)
			}
		}
		field.Set(slice)
	}
	return
}

// coerce coerces basic types from a string to a reflect.Value of a given kind.
func coerce(kind reflect.Kind, value, key string, index int,
	err *SchemaError) (rv reflect.Value) {
	var error os.Error
	switch kind {
	case reflect.Bool:
		if v, e := strconv.Atob(value); e == nil {
			rv = reflect.ValueOf(v)
		} else {
			error = e
		}
	case reflect.Float32:
		if v, e := strconv.Atof32(value); e == nil {
			rv = reflect.ValueOf(v)
		} else {
			error = e
		}
	case reflect.Float64:
		if v, e := strconv.Atof64(value); e == nil {
			rv = reflect.ValueOf(v)
		} else {
			error = e
		}
	case reflect.Int:
		if v, e := strconv.Atoi(value); e == nil {
			rv = reflect.ValueOf(v)
		} else {
			error = e
		}
	case reflect.Int8:
		if v, e := strconv.Atoi(value); e == nil {
			rv = reflect.ValueOf(int8(v))
		} else {
			error = e
		}
	case reflect.Int16:
		if v, e := strconv.Atoi(value); e == nil {
			rv = reflect.ValueOf(int16(v))
		} else {
			error = e
		}
	case reflect.Int32:
		if v, e := strconv.Atoi(value); e == nil {
			rv = reflect.ValueOf(int32(v))
		} else {
			error = e
		}
	case reflect.Int64:
		if v, e := strconv.Atoi64(value); e == nil {
			rv = reflect.ValueOf(v)
		} else {
			error = e
		}
	case reflect.String:
		rv = reflect.ValueOf(value)
	case reflect.Uint:
		if v, e := strconv.Atoui(value); e == nil {
			rv = reflect.ValueOf(v)
		} else {
			error = e
		}
	case reflect.Uint8:
		if v, e := strconv.Atoui(value); e == nil {
			rv = reflect.ValueOf(uint8(v))
		} else {
			error = e
		}
	case reflect.Uint16:
		if v, e := strconv.Atoui(value); e == nil {
			rv = reflect.ValueOf(uint16(v))
		} else {
			error = e
		}
	case reflect.Uint32:
		if v, e := strconv.Atoui(value); e == nil {
			rv = reflect.ValueOf(uint32(v))
		} else {
			error = e
		}
	case reflect.Uint64:
		if v, e := strconv.Atoui64(value); e == nil {
			rv = reflect.ValueOf(v)
		} else {
			error = e
		}
	}
	if error != nil {
		err.SetFieldError(key, index, error.String())
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
func (m *structMap) getOrLoad(t reflect.Type) (spec *structSpec,
	err os.Error) {
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
func (m *structMap) load(t reflect.Type, loaded *[]string) (spec *structSpec,
	err os.Error) {
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
		if !isSupportedType(field.Type) {
			continue
		}

		toLoad = nil
		switch field.Type.Kind() {
			case reflect.Map, reflect.Slice:
				et := field.Type.Elem()
				if et.Kind() == reflect.Struct {
					toLoad = et
				}
			case reflect.Struct:
				toLoad = field.Type
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
				return nil, os.NewError("Field names and name tags in a " +
										"struct must be unique.")
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
			case reflect.Struct:
				return true
			case reflect.Map:
				// Only map[string]anyOfTheBaseTypes.
				stringKey := t.Key().Kind() == reflect.String
				if stringKey &&	isSupportedBasicType(t.Elem()) {
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
			reflect.Uint64:
			return true
	}
	return false
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
