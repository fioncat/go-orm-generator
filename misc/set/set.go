package set

// implements of the set collection.

// Set is a collection that does not allow duplicate elements.
// The current set only supports storing string types.
// This struct is concurrent-unsafe.
type Set struct {
	slice []string
	m     map[string]int
}

// New creates a new Set using the specified elements.
func New(eles ...string) *Set {
	set := new(Set)
	for _, ele := range eles {
		set.Append(ele)
	}
	return set
}

// Append an element to the Set. If the element exists,
// it will overwrite the original one.
func (s *Set) Append(ele string) {
	if idx, ok := s.m[ele]; ok {
		s.slice[idx] = ele
		return
	}
	s.m[ele] = len(s.slice)
	s.slice = append(s.slice, ele)
}

// Slice returns the set data as a slice.
func (s *Set) Slice() []string {
	return s.slice
}
