// langur/object/object_test.go
package object

import (
	"runtime"
	"testing"
)

func TestObjectCopyRefSlice(t *testing.T) {
	slc := []Object{NumberFromInt(7), NewString("no foo bars, please")}
	slc2 := CopyRefSlice(slc)
	slc[0] = NumberFromInt(21)
	// slc[1] = NewString("fooey")
	runtime.GC()

	if len(slc2) != 2 {
		t.Fatalf("Copy Object Ref Slice failed; expected 2 objects, received %d", len(slc2))
	}
	n, ok := slc2[0].(*Number)
	if !ok {
		t.Fatalf("Copy Object Ref Slice failed; first value not *Number")
	}
	i, ok := NumberToInt(n)
	if !ok {
		t.Fatalf("Copy Object Ref Slice failed; first value not an integer")
	}
	if i != 7 {
		t.Fatalf("Copy Object Ref Slice failed; first value not 7")
	}

	s, ok := slc2[1].(*String)
	if !ok {
		t.Fatalf("Copy Object Ref Slice failed; second value not *String")
	}
	if s.String() != "no foo bars, please" {
		t.Fatalf("Copy Object Ref Slice failed; string == %q", s.String())
	}
}

func TestObjectCopySlice(t *testing.T) {
	slc := []Object{NumberFromInt(7), NewString("no foo bars, please")}
	slc2 := CopySlice(slc)
	slc[0] = NumberFromInt(21)
	// slc[1] = NewString("fooey")
	runtime.GC()

	if len(slc2) != 2 {
		t.Fatalf("Copy Object Slice failed; expected 2 objects, received %d", len(slc2))
	}
	n, ok := slc2[0].(*Number)
	if !ok {
		t.Fatalf("Copy Object Slice failed; first value not *Number")
	}
	i, ok := NumberToInt(n)
	if !ok {
		t.Fatalf("Copy Object Slice failed; first value not an integer")
	}
	if i != 7 {
		t.Fatalf("Copy Object Slice failed; first value not 7")
	}

	s, ok := slc2[1].(*String)
	if !ok {
		t.Fatalf("Copy Object Slice failed; second value not *String")
	}
	if s.String() != "no foo bars, please" {
		t.Fatalf("Copy Object Slice failed; string == %q", s.String())
	}
}

func TestObjectCopyAndReverseEvenSlice(t *testing.T) {
	slc := []Object{NewString("no foo bars, please"), NumberFromInt(7)}
	slc2 := CopyAndReverseSlice(slc)
	slc[0] = NumberFromInt(21)
	runtime.GC()

	if len(slc2) != 2 {
		t.Fatalf("Copy Object Slice failed; expected 2 objects, received %d", len(slc2))
	}
	n, ok := slc2[0].(*Number)
	if !ok {
		t.Fatalf("Copy Object Slice failed; first value not *Number")
	}
	i, ok := NumberToInt(n)
	if !ok {
		t.Fatalf("Copy Object Slice failed; first value not an integer")
	}
	if i != 7 {
		t.Fatalf("Copy Object Slice failed; first value not 7")
	}

	s, ok := slc2[1].(*String)
	if !ok {
		t.Fatalf("Copy Object Slice failed; second value not *String")
	}
	if s.String() != "no foo bars, please" {
		t.Fatalf("Copy Object Slice failed; string == %q", s.String())
	}
}

func TestObjectCopyAndReverseUnevenSlice(t *testing.T) {
	slc := []Object{NewString("no foo bars, please"), NumberFromInt(14), NumberFromInt(7)}
	slc2 := CopyAndReverseSlice(slc)
	slc = []Object{NewString("fooey"), NumberFromInt(42), NumberFromInt(21)}
	runtime.GC()

	if len(slc2) != 3 {
		t.Fatalf("Copy Object Slice failed; expected 2 objects, received %d", len(slc2))
	}
	n, ok := slc2[0].(*Number)
	if !ok {
		t.Fatalf("Copy Object Slice failed; first value not *Number")
	}
	i, ok := NumberToInt(n)
	if !ok {
		t.Fatalf("Copy Object Slice failed; first value not an integer")
	}
	if i != 7 {
		t.Fatalf("Copy Object Slice failed; first value not 7")
	}

	n, ok = slc2[1].(*Number)
	if !ok {
		t.Fatalf("Copy Object Slice failed; second value not *Number")
	}
	i, ok = NumberToInt(n)
	if !ok {
		t.Fatalf("Copy Object Slice failed; second value not an integer")
	}
	if i != 14 {
		t.Fatalf("Copy Object Slice failed; second value not 14")
	}

	s, ok := slc2[2].(*String)
	if !ok {
		t.Fatalf("Copy Object Slice failed; third value not *String")
	}
	if s.String() != "no foo bars, please" {
		t.Fatalf("Copy Object Slice failed; string == %q", s.String())
	}
}
