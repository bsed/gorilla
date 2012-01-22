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
	"appengine_internal"
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
			pv.Referencevalue = keyToReferenceValue(k)
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

// ----------------------------------------------------------------------------
// Get
// ----------------------------------------------------------------------------

// Get loads the entity stored for k into dst, which may be either a struct
// pointer, a PropertyLoadSaver or a Map (although Maps are deprecated). If
// there is no such entity for the key, Get returns ErrNoSuchEntity.
//
// The values of dst's unmatched struct fields or Map entries are not modified.
// In particular, it is recommended to pass either a pointer to a zero valued
// struct or an empty Map on each Get call.
//
// ErrFieldMismatch is returned when a field is to be loaded into a different
// type than the one it was stored from, or when a field is missing or
// unexported in the destination struct. ErrFieldMismatch is only returned if
// dst is a struct pointer.
func Get(c appengine.Context, key *Key, dst interface{}) os.Error {
	err := GetMulti(c, []*Key{key}, []interface{}{dst})
	if errMulti, ok := err.(ErrMulti); ok {
		return errMulti[0]
	}
	return err
}

// GetMulti is a batch version of Get.
func GetMulti(c appengine.Context, key []*Key, dst []interface{}) os.Error {
	if len(key) != len(dst) {
		return os.NewError("datastore: key and dst slices have different length")
	}
	if len(key) == 0 {
		return nil
	}
	if err := multiValid(key); err != nil {
		return err
	}
	req := &pb.GetRequest{
		Key: multiKeyToProto(c.FullyQualifiedAppID(), key),
	}
	res := &pb.GetResponse{}
	err := c.Call("datastore_v3", "Get", req, res, nil)
	if err != nil {
		return err
	}
	if len(key) != len(res.Entity) {
		return os.NewError("datastore: internal error: server returned the wrong number of entities")
	}
	errMulti := make(ErrMulti, len(key))
	for i, e := range res.Entity {
		if e.Entity == nil {
			errMulti[i] = ErrNoSuchEntity
			continue
		}
		errMulti[i] = loadEntity(dst[i], e.Entity)
	}
	for _, e := range errMulti {
		if e != nil {
			return errMulti
		}
	}
	return nil
}

// ----------------------------------------------------------------------------
// Put
// ----------------------------------------------------------------------------

// Put saves the entity src into the datastore with key k. src may be either a
// struct pointer, a PropertyLoadSaver or a Map (although Maps are deprecated);
// if the former then any unexported fields of that struct will be skipped.
// If k is an incomplete key, the returned key will be a unique key
// generated by the datastore.
func Put(c appengine.Context, key *Key, src interface{}) (*Key, os.Error) {
	k, err := PutMulti(c, []*Key{key}, []interface{}{src})
	if err != nil {
		if errMulti, ok := err.(ErrMulti); ok {
			return nil, errMulti[0]
		}
		return nil, err
	}
	return k[0], nil
}

// PutMulti is a batch version of Put.
func PutMulti(c appengine.Context, key []*Key, src []interface{}) ([]*Key, os.Error) {
	if len(key) != len(src) {
		return nil, os.NewError("datastore: key and src slices have different length")
	}
	if len(key) == 0 {
		return nil, nil
	}
	appID := c.FullyQualifiedAppID()
	if err := multiValid(key); err != nil {
		return nil, err
	}
	req := &pb.PutRequest{}
	for i := range src {
		sProto, err := saveEntity(appID, key[i], src[i])
		if err != nil {
			return nil, err
		}
		req.Entity = append(req.Entity, sProto)
	}
	res := &pb.PutResponse{}
	err := c.Call("datastore_v3", "Put", req, res, nil)
	if err != nil {
		return nil, err
	}
	if len(key) != len(res.Key) {
		return nil, os.NewError("datastore: internal error: server returned the wrong number of keys")
	}
	ret := make([]*Key, len(key))
	for i := range ret {
		ret[i], err = protoToKey(res.Key[i])
		if err != nil || ret[i].Incomplete() {
			// TODO: improve this error message. use the result from
			// Key.valid(). Improve that one too.
			return nil, os.NewError("datastore: internal error: server returned an invalid key")
		}
	}
	return ret, nil
}

// ----------------------------------------------------------------------------
// Delete
// ----------------------------------------------------------------------------

// Delete deletes the entity for the given key.
func Delete(c appengine.Context, key *Key) os.Error {
	err := DeleteMulti(c, []*Key{key})
	if errMulti, ok := err.(ErrMulti); ok {
		return errMulti[0]
	}
	return err
}

// DeleteMulti is a batch version of Delete.
func DeleteMulti(c appengine.Context, key []*Key) os.Error {
	if len(key) == 0 {
		return nil
	}
	if err := multiValid(key); err != nil {
		return err
	}
	req := &pb.DeleteRequest{
		Key: multiKeyToProto(c.FullyQualifiedAppID(), key),
	}
	res := &pb.DeleteResponse{}
	return c.Call("datastore_v3", "Delete", req, res, nil)
}

// asStructValue converts a struct pointer to a reflect.Value.
// TODO: this is not used anywhere
func asStructValue(x interface{}) (reflect.Value, os.Error) {
	pv := reflect.ValueOf(x)
	if pv.Kind() != reflect.Ptr || pv.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, ErrInvalidEntityType
	}
	return pv.Elem(), nil
}

func init() {
	appengine_internal.RegisterErrorCodeMap("datastore_v3", pb.Error_ErrorCode_name)
}
