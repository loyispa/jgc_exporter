package health

type RingBuffer[T any] struct {
	data  []T
	cap   int
	start int
	size  int
}

func NewRingBuffer[T any](capacity int) *RingBuffer[T] {
	return &RingBuffer[T]{
		data: make([]T, capacity),
		cap:  capacity,
	}
}

func (rb *RingBuffer[T]) Push(item T) {
	idx := (rb.start + rb.size) % rb.cap
	rb.data[idx] = item
	if rb.size < rb.cap {
		rb.size++
	} else {
		rb.start = (rb.start + 1) % rb.cap
	}
}

func (rb *RingBuffer[T]) Slice() []T {
	if rb.size == 0 {
		return nil
	}
	result := make([]T, rb.size)
	for i := 0; i < rb.size; i++ {
		result[i] = rb.data[(rb.start+i)%rb.cap]
	}
	return result
}

func (rb *RingBuffer[T]) Len() int {
	return rb.size
}
