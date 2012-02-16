package gdt

import (
	"reflect"

	pb "appengine_internal/datastore"
)

// TODO
func valueToProto(property string, value interface{}) (*pb.Property, error) {
	v := reflect.ValueOf(value)
	if v.IsValid() {}
	return nil, nil
}
