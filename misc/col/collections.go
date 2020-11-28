package col

type Set map[string]interface{}

func NewSet(n int) Set {
	return make(Set, n)
}

func NewSetBySlice(ss ...string) Set {
	set := NewSet(len(ss))
	for _, s := range ss {
		set.Add(s)
	}
	return set
}

func (s Set) Add(key string) {
	s[key] = struct{}{}
}

func (s Set) Exists(key string) bool {
	_, ok := s[key]
	return ok
}

func (s Set) Traverse(f func(key string) bool) {
	for key := range s {
		if ok := f(key); !ok {
			break
		}
	}
}

func (s Set) Slice() []string {
	if s == nil {
		return nil
	}
	slice := make([]string, 0, len(s))
	for key := range s {
		slice = append(slice, key)
	}
	return slice
}
