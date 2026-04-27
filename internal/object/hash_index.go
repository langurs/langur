// langur/object/hash_index.go

package object

import (
	"fmt"
)

func (d *Hash) IndexCount() int {
	return len(d.Pairs)
}

func (left *Hash) IndexKeys() *List {
	list := &List{}
	for _, kv := range left.Pairs {
		list.Elements = append(list.Elements, kv.Key)
	}
	return list
}

// returnOtherObjType: unused on hash
func (left *Hash) Index(index Object, returnOtherObjType bool) (result Object, err error) {
	result, err, _ = left.index(index, returnOtherObjType)
	if err != nil {
		return left, fmt.Errorf("Index out of range")
	}
	return
}

func (left *Hash) index(index Object, returnOtherObjType bool) (result Object, err error, isPoly bool) {
	if returnOtherObjType {
		return left, fmt.Errorf("No alternate return type for Hash"), false
	}

	switch idx := index.(type) {
	case *List:
		list := &List{}
		for _, v := range idx.Elements {
			e, err, poly := left.index(v, returnOtherObjType)
			if err != nil {
				return left, err, poly
			}
			if poly {
				for _, e2 := range e.(*List).Elements {
					list.Elements = append(list.Elements, e2)
				}
			} else {
				list.Elements = append(list.Elements, e)
			}
		}
		return list, nil, true
		
	case *Range:
		list, err := idx.ToList(One, true)
		if err != nil {
			return left, fmt.Errorf("Error generating list from range: %s", err.Error()), true
		}
		// call self with list...
		return left.index(list, returnOtherObjType)
	}

	if !IsValidForHashKey(index) {
		return left, fmt.Errorf("Unusable hash key (%s)", index.TypeString()), false
	}

	value, err := left.GetValue(index)
	if err != nil {
		return left, fmt.Errorf("Hash key not found"), false
	}

	return value, nil, false
}

func (d *Hash) IndexInverse(index Object, returnOtherObjType bool) (
	result Object, err error) {

	var keySlc []Object

	switch idx := index.(type) {
	case *List:
		keySlc = idx.Elements

	case *Range:
		list, err := idx.ToList(One, true)
		if err != nil {
			return idx, fmt.Errorf("Error generating list from range: %s", err.Error())
		}
		keySlc = list.Elements

	default:
		keySlc = []Object{idx}
	}

	hash := &Hash{}
	hash.Pairs = make([]keyValuePair, 0, len(d.Pairs)-len(keySlc))

	for _, kv := range d.Pairs {
		addThis := true
		for _, k := range keySlc {
			if d.keyIndex(k) == -1 {
				return d, fmt.Errorf("Invalid index on hash")
			}
			if compareHashKeys(kv.Key, MakeHashKey(k)) {
				addThis = false
				break
			}
		}
		if addThis {
			hash.Pairs = append(hash.Pairs, kv)
		}
	}

	return hash, nil
}

func (left *Hash) IndexValid(index Object) bool {
	switch idx := index.(type) {
	case *List:
		for _, v := range idx.Elements {
			if !left.IndexValid(v) {
				return false
			}
		}
		return true

	case *Range:
		list, err := idx.ToList(One, true)
		if err != nil {
			return false
		}
		return left.IndexValid(list)
	}

	_, err := left.GetValue(index)
	return err == nil
}

// assumes mutability of an object (checked elsewhere)
func (left *Hash) SetIndex(index, setTo Object) (Object, error) {
	// any valid hash key setable
	// can create new value if doesn't exist
	if !IsValidForHashKey(index) {
		return left, fmt.Errorf("Cannot set hash value from invalid index (not hashable)")
	}
	left.WritePair(index, setTo)
	return left, nil
}
