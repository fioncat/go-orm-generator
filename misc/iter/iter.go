package iter

import "reflect"

// implements of iterator.

// Iter represents an iterator. It receives slice as input,
// and each time Next is called, it will access the elements
// in the slice in order and assign values.
// The usual usage is:
//   slice := []string{"a", "b", "c"}
//   iter := iter.New(slice)
//   var s string
//   for iter.Next(&s) {
//      ...
//   }
type Iter struct {
	idx int

	slen  int
	slice reflect.Value
}

// New create a new Iterator, the "slice" must be a slice type value,
// otherwise, this function will panic.
func New(slice interface{}) *Iter {
	slicev := reflect.ValueOf(slice)
	if slicev.Kind() != reflect.Slice {
		panic("passed a non-slice param to iter.New")
	}

	iter := new(Iter)
	iter.idx = 0
	iter.slice = slicev
	iter.slen = slicev.Len()

	return iter
}

// NextP returns the current element and increments the counter.
// It returns the index of the current element. If the iteration
// ends, return -1.
// Because the counter is incremented, the next call will take
// out the next element, thus achieving the effect of iterating
// in order.
// The parameter "v" must be a pointer, and the type pointed to
// must be equal to the type of the slice element. Otherwise, the
// function will panic.
func (iter *Iter) NextP(v interface{}) int {
	tar := reflect.ValueOf(v)
	if tar.Kind() != reflect.Ptr {
		panic("passed a non-pointer param to Next")
	}

	if iter.idx >= iter.slen {
		// the iteration is end
		return -1
	}

	curidx := iter.idx
	cur := iter.slice.Index(curidx)
	iter.idx += 1

	tar.Elem().Set(cur)

	return curidx
}

// Next is same as NextP, except that it returns
// whether the iteration is over.
func (iter *Iter) Next(v interface{}) bool {
	return iter.NextP(v) >= 0
}

// Pick returns the current element, but does not change the counter.
func (iter *Iter) Pick() interface{} {
	if iter.idx >= iter.slen {
		return nil
	}
	return iter.slice.Index(iter.idx).Interface()
}

// Reset set the counter to zero.
func (iter *Iter) Reset() {
	iter.idx = 0
}
