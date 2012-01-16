// Copyright 2011 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package datastore

import (
	"fmt"
	"os"
	"reflect"
	"time"

	"appengine"
	"goprotobuf.googlecode.com/hg/proto"

	pb "appengine_internal/datastore"
)

// Time is the number of microseconds since the Unix epoch,
// January 1, 1970 00:00:00 UTC.
//
// It is a distinct type so that loading and saving fields of type Time are
// displayed correctly in App Engine tools like the Admin Console.
type Time int64

// Time returns a *time.Time from a datastore time.
func (t Time) Time() *time.Time {
	// TODO: once App Engine has release.r60 or later,
	// support subseconds here.  Currently we just drop them.
	return time.SecondsToUTC(int64(t) / 1e6)
}

// SecondsToTime converts an int64 number of seconds since to Unix epoch
// to a Time value.
func SecondsToTime(n int64) Time {
	return Time(n * 1e6)
}

var (
	// ErrInvalidEntityType is returned when an invalid destination entity type
	// is passed to Get, GetAll, GetMulti or Next.
	ErrInvalidEntityType = os.NewError("datastore: invalid entity type")
	// ErrInvalidKey is returned when an invalid key is presented.
	ErrInvalidKey = os.NewError("datastore: invalid key")
	// ErrNoSuchEntity is returned when no entity was found for a given key.
	ErrNoSuchEntity = os.NewError("datastore: no such entity")
)

// ErrFieldMismatch is returned when a field is to be loaded into a different
// type than the one it was stored from, or when a field is missing or
// unexported in the destination struct.
// StructType is the type of the struct pointed to by the destination argument
// passed to Get or to Iterator.Next.
type ErrFieldMismatch struct {
	StructType reflect.Type
	FieldName  string
	Reason     string
}

// String returns a string representation of the error.
func (e *ErrFieldMismatch) String() string {
	return fmt.Sprintf("datastore: cannot load field %q into a %q: %s",
		e.FieldName, e.StructType, e.Reason)
}

// valueToProto converts a named value to a newly allocated Property.
// The returned error string is empty on success.
func valueToProto(name string, v reflect.Value, multiple bool) (p *pb.Property, errStr string) {
	var (
		pv          pb.PropertyValue
		unsupported bool
	)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		pv.Int64Value = proto.Int64(v.Int())
	case reflect.Bool:
		pv.BooleanValue = proto.Bool(v.Bool())
	case reflect.String:
		pv.StringValue = proto.String(v.String())
	case reflect.Float32, reflect.Float64:
		pv.DoubleValue = proto.Float64(v.Float())
	case reflect.Ptr:
		if k, ok := v.Interface().(*Key); ok {
			if k == nil {
				return nil, nilKeyErrStr
			}
			pv.Referencevalue = k.toReferenceValue()
		} else {
			unsupported = true
		}
	case reflect.Slice:
		if b, ok := v.Interface().([]byte); ok {
			pv.StringValue = proto.String(string(b))
		} else {
			// nvToProto should already catch slice values.
			// If we get here, we have a slice of slice values.
			unsupported = true
		}
	default:
		unsupported = true
	}
	if unsupported {
		return nil, "unsupported datastore value type: " + v.Type().String()
	}
	p = &pb.Property{
		Name:     proto.String(name),
		Value:    &pv,
		Multiple: proto.Bool(multiple),
	}
	switch v.Interface().(type) {
	case []byte:
		p.Meaning = pb.NewProperty_Meaning(pb.Property_BLOB)
	case appengine.BlobKey:
		p.Meaning = pb.NewProperty_Meaning(pb.Property_BLOBKEY)
	case Time:
		p.Meaning = pb.NewProperty_Meaning(pb.Property_GD_WHEN)
	}
	return p, ""
}
