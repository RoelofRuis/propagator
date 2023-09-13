package propagator

type SetQueue[T comparable] struct {
	Elements []T
	set      Set[T]
}

func NewSetQueue[T comparable]() *SetQueue[T] {
	return &SetQueue[T]{set: make(Set[T])}
}

func (q *SetQueue[T]) Length() int {
	return len(q.Elements)
}

func (q *SetQueue[T]) Reset() {
	q.Elements = []T{}
	q.set = make(Set[T])
}

func (q *SetQueue[T]) Enqueue(elems ...T) {
	for _, elem := range elems {
		if q.set.Contains(elem) {
			continue
		}
		q.Elements = append(q.Elements, elem)
		q.set.Insert(elem)
	}
}

func (q *SetQueue[T]) IsEmpty() bool {
	return len(q.Elements) == 0
}

func (q *SetQueue[T]) Peek() T {
	if q.IsEmpty() {
		panic("empty SetQueue")
	}
	return q.Elements[0]
}

func (q *SetQueue[T]) Dequeue() T {
	element := q.Peek()
	delete(q.set, element)
	if len(q.Elements) == 1 {
		q.Elements = nil
	} else {
		q.Elements = q.Elements[1:]
	}
	return element
}
