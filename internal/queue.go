package internal

// generics Queue
type Queue[T comparable] struct {
	elements []T
	unique   map[T]bool
}

func (q *Queue[T]) Push(element T) {
	q.elements = append(q.elements, element)
}
func (q *Queue[T]) PushUnique(element T) {
	if q.unique == nil {
		q.unique = make(map[T]bool)
	}
	if _, ok := q.unique[element]; !ok {
		q.elements = append(q.elements, element)
		q.unique[element] = true
	}
}

func (q *Queue[T]) Len() int {
	return len(q.elements)
}

func (q *Queue[T]) Pop() T {
	element := q.elements[q.Len()-1]
	q.elements = q.elements[:q.Len()-1]
	return element
}

func (q *Queue[T]) Top() T {
	return q.elements[q.Len()-1]
}
