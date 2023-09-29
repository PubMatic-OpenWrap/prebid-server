package xmlparser

// A simple stack
type stack[T any] struct {
	data []T
}

func (s *stack[T]) push(value T) {
	s.data = append(s.data, value)
}

func (s *stack[T]) pop() *T {
	if len(s.data) == 0 {
		return nil
	}
	value := s.data[len(s.data)-1]
	s.data = s.data[:len(s.data)-1]
	return &value
}

func (s *stack[T]) peek() *T {
	if len(s.data) == 0 {
		return nil
	}
	return &s.data[len(s.data)-1]
}

func (s *stack[T]) len() int {
	return len(s.data)
}
