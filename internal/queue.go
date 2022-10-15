package internal

// generics Queue
type Queue[T any] struct {
	elements []T
}

func (q *Queue[T]) Push(element T) {
	q.elements = append(q.elements, element)
}

func (q *Queue[T]) Len() int {
	return len(q.elements)
}

func (q *Queue[T]) Pop() T {
	element := q.elements[0]
	q.elements = q.elements[1:]
	return element
}

func (q *Queue[T]) Top() T {
	return q.elements[q.Len()-1]
}
