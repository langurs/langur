// langur/object/object_interfaces.go

package object

// THE GENERAL OBJECT INTERFACE
type Object interface {
	Type() ObjectType
	TypeString() string

	Copy() Object
	Equal(Object) bool
	IsTruthy() bool

	// 0.13.1 changed the String() method to ReplString() for clarity
	// 0.13.3 changed the LangurString() method to String() and added the IComposableString interface
	String() string
	ReplString() string
}

// NOTE: not following Go convention for naming the following interfaces

// IComposableString: for including things like quote marks around strings, re// around regex, etc.
type IComposableString interface {
	Object
	ComposedString() string
}

type IHashKey interface {
	Object
	HashKey() Object
}

type INumericNegation interface {
	Object
	Negate() Object
	Abs() Object
}

type IGreaterThan interface {
	Object
	GreaterThan(Object) (bool, bool)
	// The second Boolean indicates whether they were actually comparable or not.
}

type IDefilableEffects interface {
	Object
	HasImpureEffects() bool
}

type IContains interface {
	Object
	Contains(Object) (bool, bool)
	// The second Boolean indicates whether they were actually comparable or not.
}

// NOTE: With the following interfaces, ...
// ... the methods return nil if the second object cannot be used with the first object.
// ... They may return an Error Object if there was an error.

type IIndex interface {
	Object
	Index(index Object, returnOtherObjType bool) (Object, error)
	// Index on error ...
	// return original Object if valid operation, but failed index
	// return nil for Object if not a valid operation

	IndexKeys() *List
	IndexCount() int

	IndexValid(index Object) bool
}

type IIndexSet interface {
	IIndex
	SetIndex(index, setTo Object) (Object, error)
}

type IIndexInverse interface {
	IIndex
	IndexInverse(index Object, returnOtherObjType bool) (Object, error)
}

type IIndexNativeInt interface {
	IIndex
	IndexNativeInt(Object) (int, bool)
	indexNativeInt(int) (int, bool)
}

// type IIterable interface {
// 	Object
// }

// type IEnumerable interface {
// 	IIndex
// 	IndexNativeInt(Object) (int, bool)
// }

type IForward interface {
	Object
	Forward(Object) Object
}

type IAppend interface {
	Object
	Append(Object) Object
}

type IAppendToNone interface {
	Object
	AppendToNone() Object
}

type IAdd interface {
	Object
	Add(Object) Object
}

type ISubtract interface {
	Object
	Subtract(Object) Object
}

type IMultiply interface {
	Object
	Multiply(Object) Object
}

type IDivide interface {
	Object
	Divide(Object) Object
}

type IDivideTruncate interface {
	Object
	DivideTruncate(Object) Object
}

type IDivideFloor interface {
	Object
	DivideFloor(Object) Object
}

type IRemainder interface {
	Object
	Remainder(Object) Object
}

type IModulus interface {
	Object
	Modulus(Object) Object
}

type IDivisibleBy interface {
	Object
	DivisibleBy(Object) (bool, bool)
	// The second Boolean indicates whether they were actually comparable or not.
}

type IPower interface {
	Object
	Power(Object) Object
}

type IRoot interface {
	Object
	Root(Object) Object
}

type ISimplify interface {
	Object
	Simplify() Object
}
