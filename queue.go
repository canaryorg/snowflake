package snowflake

type queue struct {
	ch   chan int64
	head int
}

func newQueue(bufferSize int) *queue {
	return &queue{
		ch:   make(chan int64, bufferSize),
		head: -1,
	}
}
