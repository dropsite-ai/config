package config

import "testing"

type srcStruct struct {
	FieldA string
	FieldB int
}

type dstStruct struct {
	FieldX string
	FieldY int
}

func TestCopyProperty_Valid(t *testing.T) {
	src := srcStruct{
		FieldA: "hello",
		FieldB: 42,
	}
	dst := dstStruct{}
	if err := CopyProperty(&src, "FieldA", &dst, "FieldX"); err != nil {
		t.Fatalf("CopyProperty returned error: %v", err)
	}
	if dst.FieldX != "hello" {
		t.Errorf("Expected dst.FieldX to be %q, got %q", "hello", dst.FieldX)
	}
	// Also copy an int field.
	if err := CopyProperty(&src, "FieldB", &dst, "FieldY"); err != nil {
		t.Fatalf("CopyProperty returned error: %v", err)
	}
	if dst.FieldY != 42 {
		t.Errorf("Expected dst.FieldY to be 42, got %d", dst.FieldY)
	}
}

func TestCopyProperty_InvalidSourceField(t *testing.T) {
	src := srcStruct{FieldA: "test"}
	dst := dstStruct{}
	err := CopyProperty(&src, "NonExistent", &dst, "FieldX")
	if err == nil {
		t.Error("Expected error when source field does not exist")
	}
}

func TestCopyProperty_InvalidDestinationField(t *testing.T) {
	src := srcStruct{FieldA: "test"}
	dst := dstStruct{}
	err := CopyProperty(&src, "FieldA", &dst, "NonExistent")
	if err == nil {
		t.Error("Expected error when destination field does not exist")
	}
}

func TestCopyProperty_TypeMismatch(t *testing.T) {
	// Attempt to copy an int into a string.
	src := srcStruct{FieldB: 100}
	dst := dstStruct{}
	err := CopyProperty(&src, "FieldB", &dst, "FieldX")
	if err == nil {
		t.Error("Expected error due to type mismatch")
	}
}
