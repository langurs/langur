// langur/object/hash_test.go
package object

// tests with using slice internally

import (
	"testing"
)

func TestHashStringKey(t *testing.T) {
	one1 := NewString("yo, what's up?")
	one2 := NewString("yo, what's up?")
	two1 := NewString("great")
	two2 := NewString("great")

	o1, o2, t1, t2 := &Hash{}, &Hash{}, &Hash{}, &Hash{}
	o1.WritePair(one1, FALSE)
	o2.WritePair(one2, FALSE)
	t1.WritePair(two1, FALSE)
	t2.WritePair(two2, FALSE)

	if o1.Pairs[0].Key.String() != o2.Pairs[0].Key.String() {
		t.Errorf("strings with same content have different hash keys")
	}

	if t1.Pairs[0].Key.String() != t2.Pairs[0].Key.String() {
		t.Errorf("strings with same content have different hash keys")
	}

	if o1.Pairs[0].Key.String() == t1.Pairs[0].Key.String() {
		t.Errorf("strings with different content have same hash keys")
	}
}

func TestHashNumberKey(t *testing.T) {
	one1 := NumberFromInt(1)
	one2 := NumberFromInt(1)
	two1 := NumberFromInt(2)
	two2 := NumberFromInt(2)

	o1, o2, t1, t2 := &Hash{}, &Hash{}, &Hash{}, &Hash{}
	o1.WritePair(one1, NULL)
	o2.WritePair(one2, NULL)
	t1.WritePair(two1, NULL)
	t2.WritePair(two2, NULL)

	if !o1.Pairs[0].Key.(*Number).Equal(o2.Pairs[0].Key.(*Number)) {
		t.Errorf("numbers with same content have different hash keys")
	}

	if !t1.Pairs[0].Key.(*Number).Equal(t2.Pairs[0].Key.(*Number)) {
		t.Errorf("numbers with same content have different hash keys")
	}

	if o1.Pairs[0].Key.(*Number).Equal(t1.Pairs[0].Key.(*Number)) {
		t.Errorf("numbers with different content have same hash keys")
	}
}

func TestHashStringValues(t *testing.T) {
	o1, o2, t1, t2 := &Hash{}, &Hash{}, &Hash{}, &Hash{}

	o1.WritePair(NumberFromInt(1), NewString("abc"))
	o2.WritePair(NumberFromInt(2), NewString("abc"))
	t1.WritePair(NumberFromInt(1), NewString("def"))
	t2.WritePair(NumberFromInt(2), NewString("def"))

	if o1.Pairs[0].Value.String() != o2.Pairs[0].Value.String() {
		t.Errorf("different hashes, same key, different value (should be same value)")
	}

	if t1.Pairs[0].Value.String() != t2.Pairs[0].Value.String() {
		t.Errorf("different hashes, same key, different value (should be same value)")
	}

	if o1.Pairs[0].Value.String() == t1.Pairs[0].Value.String() {
		t.Errorf("different hashes, same key, same values (should be different values)")
	}
}

func TestHashNumberValues(t *testing.T) {
	o1, o2, t1, t2 := &Hash{}, &Hash{}, &Hash{}, &Hash{}

	o1.WritePair(NumberFromInt(1), NumberFromInt(1))
	o2.WritePair(NumberFromInt(2), NumberFromInt(1))
	t1.WritePair(NumberFromInt(1), NumberFromInt(2))
	t2.WritePair(NumberFromInt(2), NumberFromInt(2))

	if !o1.Pairs[0].Value.(*Number).Equal(o2.Pairs[0].Value.(*Number)) {
		t.Errorf("different hashes, same key, different value (should be same value)")
	}

	if !t1.Pairs[0].Value.(*Number).Equal(t2.Pairs[0].Value.(*Number)) {
		t.Errorf("different hashes, same key, different value (should be same value)")
	}

	if o1.Pairs[0].Value.(*Number).Equal(t1.Pairs[0].Value.(*Number)) {
		t.Errorf("different hashes, same key, same values (should be different values)")
	}
}

func TestHashWithOverWrite(t *testing.T) {
	o1 := &Hash{}

	o1.WritePair(NumberFromInt(1), NumberFromInt(1))
	o1.WritePair(NumberFromInt(2), NumberFromInt(2))
	o1.WritePair(NumberFromInt(1), NewString("asdf"))

	if len(o1.Pairs) != 2 {
		t.Fatalf("expected 2 elements in hash; received=%d", len(o1.Pairs))
	}

	if o1.Pairs[0].Key.(*Number).String() != "1" || o1.Pairs[1].Key.(*Number).String() != "2" {
		t.Errorf("hash keys not as expected")
	}

	if o1.Pairs[0].Value.Type() != STRING_OBJ {
		t.Errorf("expected string object for hash value")
	}
	if o1.Pairs[1].Value.Type() != NUMBER_OBJ {
		t.Errorf("expected number object for hash value")
	}
}

func TestHashWithoutOverWrite(t *testing.T) {
	o1 := &Hash{}

	o1.WritePairIfKeyNotPresent(NumberFromInt(1), NumberFromInt(1))
	o1.WritePairIfKeyNotPresent(NumberFromInt(2), NumberFromInt(2))
	o1.WritePairIfKeyNotPresent(NumberFromInt(1), NewString("asdf"))

	if len(o1.Pairs) != 2 {
		t.Fatalf("expected 2 elements in hash; received=%d", len(o1.Pairs))
	}

	if o1.Pairs[0].Key.(*Number).String() != "1" || o1.Pairs[1].Key.(*Number).String() != "2" {
		t.Errorf("hash keys not as expected")
	}

	if o1.Pairs[0].Value.Type() != NUMBER_OBJ {
		t.Errorf("expected number object for hash value")
	}
	if o1.Pairs[1].Value.Type() != NUMBER_OBJ {
		t.Errorf("expected number object for hash value")
	}
}

func TestHashGetValue(t *testing.T) {
	o1 := &Hash{}
	key1 := NumberFromInt(1)
	key2 := NumberFromInt(2)
	key3 := NumberFromInt(0)

	o1.WritePair(key1, NewString("asdf"))
	o1.WritePair(key2, NumberFromInt(21))

	value1, err := o1.GetValue(key1)
	if err != nil {
		t.Errorf("GetValue failed: %s", err)
	}
	if value1.Type() != STRING_OBJ {
		t.Fatalf("expected string object, received=%T", value1)
	}
	if value1.String() != "asdf" {
		t.Errorf(`expected boring string "asdf", received=%q`, value1.String())
	}

	value2, err := o1.GetValue(key2)
	if err != nil {
		t.Errorf("GetValue failed: %s", err)
	}
	if value2.Type() != NUMBER_OBJ {
		t.Fatalf("expected number object, received=%T", value1)
	}
	if !value2.(*Number).Equal(NumberFromInt(21)) {
		t.Errorf("expected 21, received=%s", value2.String())
	}

	_, err = o1.GetValue(key3)
	if err == nil {
		t.Errorf("expected no value for invalid key")
	}
}

func TestHashKeyIndex(t *testing.T) {
	o1 := &Hash{}
	key1 := NumberFromInt(1)
	key2 := NumberFromInt(2)
	key3 := NumberFromInt(0)

	o1.WritePair(key1, NewString("asdf"))
	o1.WritePair(key2, NumberFromInt(7))

	idx := o1.keyIndex(key1)
	if idx != 0 {
		t.Errorf("expected key index 0, received=%d", idx)
	}
	idx = o1.keyIndex(key2)
	if idx != 1 {
		t.Errorf("expected key index 1, received=%d", idx)
	}
	idx = o1.keyIndex(key3)
	if idx != -1 {
		t.Errorf("expected key index -1, received=%d", idx)
	}
}
