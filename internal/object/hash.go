// langur/object/hash.go

/*
NOTE: To ensure things go smoothly, other packages should call these functions ...
... to create new hashes, add to them, or remove from them.

Iterating over pairs in Go should be done in the following manner.

for _, kv := range h.Pairs {
	... kv.Key
	... kv.Value
}

The Pairs slice is exportable so it can be iterated over by built-in functions.

Hashes might be considered *unordered* in case the implementation changes.
*/

package object

import (
	"fmt"
	"langur/common"
	"langur/str"
	"strings"
)

type keyValuePair struct {
	Key   Object
	Value Object
}

func (kv keyValuePair) Copy() keyValuePair {
	return keyValuePair{Key: kv.Key.Copy(), Value: kv.Value.Copy()}
}

type Hash struct {
	Pairs []keyValuePair
}

var EmptyHash = &Hash{}

func (d *Hash) HasImpureEffects() bool {
	for _, kv := range d.Pairs {
		if SliceHasImpureEffects(kv.Value) {
			return true
		}
	}
	return false
}

func (d *Hash) keyIndex(key Object) int {
	for i, kv := range d.Pairs {
		if compareHashKeys(key, kv.Key) {
			return i
		}
	}
	return -1
}

func compareHashKeys(k1, k2 Object) bool {
	return Equal(k1, k2)
}

func IsValidForHashKey(obj Object) bool {
	_, ok := obj.(IHashKey)
	return ok
}

func MakeHashKey(obj Object) Object {
	left, ok := obj.(IHashKey)
	if ok {
		return left.HashKey()
	}
	// nil: not a valid hash key
	return nil
}

func (d *Hash) Copy() Object {
	newHash := &Hash{Pairs: make([]keyValuePair, len(d.Pairs))}
	for i, kv := range d.Pairs {
		newHash.Pairs[i] = kv.Copy()
	}
	return newHash
}
func (d *Hash) CopyRefs() Object {
	newHash := &Hash{Pairs: make([]keyValuePair, len(d.Pairs))}
	for i, kv := range d.Pairs {
		newHash.Pairs[i] = kv
	}
	return newHash
}

func (l *Hash) Equal(o2 Object) bool {
	r, ok := o2.(*Hash)
	if !ok {
		return false
	}
	if len(l.Pairs) != len(r.Pairs) {
		return false
	}
	// compare keys and values
	for _, kv := range l.Pairs {
		test, err := r.GetValue(kv.Key)
		if err != nil {
			return false
		}
		// keys in both; values the same?
		if !Equal(kv.Value, test) {
			return false
		}
	}
	return true
}

func (d *Hash) Type() ObjectType {
	return HASH_OBJ
}
func (d *Hash) TypeString() string {
	return common.HashTypeName
}

func (d *Hash) IsTruthy() bool {
	return len(d.Pairs) != 0
}

func (d *Hash) String() string {
	if len(d.Pairs) == 0 {
		return "{:}"
	}

	pairs := make([]string, 0, len(d.Pairs))
	for _, kv := range d.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			ComposedOrRegularString(kv.Key), ComposedOrRegularString(kv.Value)))
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

func (d *Hash) ReplString() string {
	if len(d.Pairs) == 0 {
		return common.HashTypeName + " {:}"
	}

	pairs := make([]string, 0, len(d.Pairs))
	for _, kv := range d.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", kv.Key.ReplString(), kv.Value.ReplString()))
	}
	return common.HashTypeName + " {" + strings.Join(pairs, ", ") + "}"
}

func (d *Hash) KeyExists(key Object) bool {
	k := MakeHashKey(key)
	if k == nil {
		return false
	}
	return d.keyIndex(k) != -1
}

func (d *Hash) WritePairIfKeyNotPresent(key, value Object) error {
	if d.KeyExists(key) {
		return nil
	} else {
		return d.WritePair(key, value)
	}
}

func (d *Hash) WritePair(key, value Object) error {
	// overwrites if key already exists in hash
	var pair keyValuePair

	k := MakeHashKey(key)
	if k == nil {
		return fmt.Errorf("Unusable hash key (%s)", hashUserValueOutput(key))
	}

	pair = keyValuePair{Key: k, Value: value}

	idx := d.keyIndex(key)
	if idx == -1 {
		d.Pairs = append(d.Pairs, pair)
	} else {
		d.Pairs[idx] = pair
	}

	return nil
}

func (d *Hash) GetValue(key Object) (Object, error) {
	if !IsValidForHashKey(key) {
		return nil, fmt.Errorf("Expected hashable key, received %s", hashUserValueOutput(key))
	}
	k := MakeHashKey(key)
	idx := d.keyIndex(k)
	if idx == -1 {
		return nil, fmt.Errorf("Key %s not found in hash", hashUserValueOutput(key))
	}
	return d.Pairs[idx].Value, nil
}

func (d *Hash) AppendWithOverWrite(d2 *Hash) (hash *Hash) {
	hash = &Hash{Pairs: make([]keyValuePair, 0, len(d.Pairs)+len(d2.Pairs))}

	for _, kv := range d.Pairs {
		hash.Pairs = append(hash.Pairs, kv)
	}

	for _, kv := range d2.Pairs {
		idx := d.keyIndex(kv.Key)
		if idx == -1 {
			hash.Pairs = append(hash.Pairs, kv)
		} else {
			hash.Pairs[idx] = kv
		}
	}
	return hash
}

func NewHashFromSlice(elements []Object, overwriteIfDupeKeys bool) (*Hash, error) {
	if len(elements)%2 == 1 {
		return nil, fmt.Errorf("Wrong list length for building hash; expected list with even number of elements")
	}

	hash := &Hash{Pairs: make([]keyValuePair, 0, len(elements)*2)}

	for i := 0; i < len(elements); i += 2 {
		key := MakeHashKey(elements[i])
		if key == nil {
			return nil, fmt.Errorf("Unusable hash key (%s)", hashUserValueOutput(elements[i]))
		}

		value := elements[i+1]
		pair := keyValuePair{Key: key, Value: value}

		idx := hash.keyIndex(key)
		if idx == -1 {
			hash.Pairs = append(hash.Pairs, pair)
		} else {
			if !overwriteIfDupeKeys {
				return nil, fmt.Errorf("New hash contains duplicate key (%s)", hashUserValueOutput(key))
			}
			hash.Pairs[idx] = pair
		}
	}

	return hash, nil
}

func NewHashFromParallelSlices(keys, values []Object, overwriteIfDupeKeys bool) (*Hash, error) {
	if len(keys) != len(values) {
		return nil, fmt.Errorf("Mismatched list lengths for building hash; expected parallel lists")
	}

	hash := &Hash{Pairs: make([]keyValuePair, 0, len(keys))}

	for i := 0; i < len(keys); i++ {
		key := MakeHashKey(keys[i])
		if key == nil {
			return nil, fmt.Errorf("Unusable hash key (%s)", hashUserValueOutput(keys[i]))
		}

		value := values[i]
		pair := keyValuePair{Key: key, Value: value}

		idx := hash.keyIndex(key)
		if idx == -1 {
			hash.Pairs = append(hash.Pairs, pair)
		} else {
			if !overwriteIfDupeKeys {
				return nil, fmt.Errorf("New hash contains duplicate key (%s)", hashUserValueOutput(key))
			}
			hash.Pairs[idx] = pair
		}
	}

	return hash, nil
}

func getHashStringValue(h *Hash, key Object) (string, error) {
	strObj, err := h.GetValue(key)
	if err != nil {
		return "", fmt.Errorf("Could not find hash key %s of type %s", hashUserValueOutput(key), key.TypeString())
	}
	s, ok := ExpectString(strObj)
	if !ok {
		return "", fmt.Errorf("Hash value %s of wrong type (not STRING)", hashUserValueOutput(key))
	}
	return s, nil
}

func getHashIntegerValue(h *Hash, key Object) (int, error) {
	intObj, err := h.GetValue(key)
	if err != nil {
		return 0, fmt.Errorf("Could not find hash key %s of type %s", hashUserValueOutput(key), key.TypeString())
	}
	i, ok := NumberToInt(intObj)
	if !ok {
		return 0, fmt.Errorf("Hash value %s of wrong type (not INTEGER)", hashUserValueOutput(key))
	}
	return i, nil
}

func getHashInteger64Value(h *Hash, key Object) (int64, error) {
	intObj, err := h.GetValue(key)
	if err != nil {
		return 0, fmt.Errorf("Could not find hash key %s of type %s", hashUserValueOutput(key), key.TypeString())
	}
	switch intObj := intObj.(type) {
	case *Number:
		i, err := intObj.ToInt64()
		if err == nil {
			return i, nil
		}
	}
	return 0, fmt.Errorf("Hash value %s of wrong type (not INTEGER)", hashUserValueOutput(key))
}

func hashUserValueOutput(obj Object) string {
	return str.ReformatInput(ComposedOrRegularString(obj))
}

func (d *Hash) Reverse() (hash *Hash, err error) {
	// reverse keys/values of hash if possible
	hash = &Hash{}
	for _, v := range d.Pairs {
		newKey := v.Value
		newValue := v.Key

		if !IsValidForHashKey(newKey) {
			err = fmt.Errorf("Could not reverse hash; not all values of hash hashable")
			return
		}

		if hash.KeyExists(newKey) {
			err = fmt.Errorf("Could not reverse hash; not all values of hash uniquely hashable")
			return
		}

		err = hash.WritePair(newKey, newValue)
	}
	return hash, err
}
