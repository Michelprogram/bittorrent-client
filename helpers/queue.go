package helpers

import (
	"errors"
	"sync"
)

type Queue[T any] struct {
	Items chan T
	mu    sync.Mutex
	con   *sync.Cond
}

func NewQueue[T any](items ...T) *Queue[T] {

	data := make(chan T, len(items))

	for _, item := range items {
		data <- item
	}

	queue := &Queue[T]{
		Items: data,
	}

	queue.con = sync.NewCond(&queue.mu)

	return queue
}

func (q *Queue[T]) Add(value T) error {

	q.mu.Lock()

	defer q.mu.Unlock()

	select {
	case q.Items <- value:
		q.con.Signal()
		return nil
	default:
		return errors.New("can't add value to channel, full")
	}

}

func (q *Queue[T]) Pop() (*T, error) {

	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.Items) == 0 {
		q.con.Wait()
	}

	val, ok := <-q.Items
	if !ok {
		return nil, errors.New("queue is closed")
	}

	return &val, nil
}
