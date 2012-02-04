package datastore

import (
	"testing"
)

func TestNamespaceKey(t *testing.T) {
	c := getContext(t)
	defer c.Close()

	k1 := NewNamespaceKey(c, "Test", "foo", 0, nil, "ns1")
	k2 := NewNamespaceKey(c, "Test", "foo", 0, k1, "")
	k3 := NewNamespaceKey(c, "Test", "foo", 0, nil, "")

	if k1.Namespace() != "ns1" {
		t.Fatalf("Wrong namespace %v, expected %v", k1.Namespace(), "ns1")
	}
	if k2.Namespace() != "ns1" {
		t.Fatalf("Wrong namespace %v, expected %v", k2.Namespace(), "ns1")
	}
	if k3.Namespace() != "" {
		t.Fatalf("Wrong namespace %v, expected %v", k3.Namespace(), "")
	}
}

func TestNamespaceKeyEquality(t *testing.T) {
	c := getContext(t)
	defer c.Close()

	k1 := NewNamespaceKey(c, "Test", "foo", 0, nil, "ns1")
	k2 := NewNamespaceKey(c, "Test", "foo", 0, nil, "ns1")
	k3 := NewNamespaceKey(c, "Test", "foo", 0, k1, "")
	k4 := NewNamespaceKey(c, "Test", "foo", 0, k2, "")
	k5 := NewNamespaceKey(c, "Test", "foo", 0, nil, "")
	k6 := NewNamespaceKey(c, "Test", "foo", 0, nil, "")
	k7 := NewNamespaceKey(c, "Test", "foo", 0, nil, "ns2")

	if !k1.Eq(k2) {
		t.Fatalf("These keys are equal: %v, %v", k1, k2)
	}
	if !k3.Eq(k4) {
		t.Fatalf("These keys are equal: %v, %v", k3, k4)
	}
	if !k5.Eq(k6) {
		t.Fatalf("These keys are equal: %v, %v", k5, k6)
	}
	if k1.Eq(k3) {
		t.Fatalf("These keys are not equal: %v, %v", k1, k3)
	}
	if k1.Eq(k5) {
		t.Fatalf("These keys are not equal: %v, %v", k1, k5)
	}
	if k1.Eq(k7) {
		t.Fatalf("These keys are not equal: %v, %v", k1, k7)
	}
}
