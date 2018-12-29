package core

type List struct {
	data []interface{}
}

func (s *List) Reset() {
	s.data = make([]interface{}, 0, 100)
}

func (s *List) Add(value interface{}) {
	if s.data == nil {
		s.data = make([]interface{}, 0, 100)
	}
	if cap(s.data) < len(s.data)+1 {
		newSlice := make([]interface{}, len(s.data), cap(s.data)*2)
		copy(newSlice, s.data)
		s.data = newSlice
	}
	s.data = append(s.data, value)
}
func (s *List) Get(i int) interface{} {
	if i < len(s.data) {
		return s.data[i]
	} else {
		return nil
	}
}
func (s *List) Size() int {
	return len(s.data)
}
func (s *List) GetAsArray() []interface{} {
	if s.data == nil {
		s.data = make([]interface{}, 0, 100)
	}
	return s.data
}
