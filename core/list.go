package core

type List struct {
	data []interface{}
}

func (list *List) Reset() {
	list.data = make([]interface{}, 0, 100)
}

func (list *List) Add(value interface{}) {
	if list.data == nil {
		list.data = make([]interface{}, 0, 100)
	}
	if cap(list.data) < len(list.data)+1 {
		newSlice := make([]interface{}, len(list.data), cap(list.data)*2)
		copy(newSlice, list.data)
		list.data = newSlice
	}
	list.data = append(list.data, value)
}

func (list *List) Get(i int) interface{} {
	if i < len(list.data) {
		return list.data[i]
	} else {
		return nil
	}
}

func (list *List) Size() int {
	return len(list.data)
}

func (list *List) GetAsArray() []interface{} {
	if list.data == nil {
		list.data = make([]interface{}, 0, 100)
	}
	return list.data
}
