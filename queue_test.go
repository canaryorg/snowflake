package snowflake

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueuePush(t *testing.T) {
	id := int64(123)
	q := newQueue(32, nil)
	ok := q.add(context.Background(), id)
	assert.True(t, ok)

	v, ok := q.peek()
	assert.True(t, ok)
	assert.Equal(t, id, v)

	v, ok = q.remove()
	assert.True(t, ok)
	assert.Equal(t, id, v)

	_, ok = q.peek()
	assert.False(t, ok)
}

func TestQueueWake(t *testing.T) {
	wake := make(chan struct{}, 1)
	q := newQueue(32, wake)
	q.add(context.Background(), 123)
	<-wake
}

func TestQueues_Next(t *testing.T) {
	ctx := context.Background()
	q1 := newQueue(32, nil)
	q2 := newQueue(32, nil)
	q3 := newQueue(32, nil)

	q1.add(ctx, 1)
	q2.add(ctx, 2)
	q3.add(ctx, 3)
	q3.add(ctx, 4)
	q2.add(ctx, 5)
	q1.add(ctx, 6)

	q := queues{q1, q2, q3}
	for i := 1; i <= 6; i++ {
		v, ok := q.Next()
		assert.True(t, ok)
		assert.Equal(t, int64(i), v, "expected %v; got %v", i, v)
	}
	_, ok := q.Next()
	assert.False(t, ok)
}

func BenchmarkQueues_Next(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bufferSize := b.N / 2
	if bufferSize < 10 {
		bufferSize = 10
	}
	q1 := newQueue(bufferSize, nil)
	q2 := newQueue(bufferSize, nil)
	q3 := newQueue(bufferSize, nil)

	q := queues{q1, q2, q3}
	previous := int64(0)
	for i := 0; i < b.N; i++ {
		if i < bufferSize {
			q1.add(ctx, int64(i)*3)
			q2.add(ctx, int64(i)*3+1)
			q3.add(ctx, int64(i)*3+2)
		}

		v, ok := q.Next()
		if !ok {
			b.Error("expected next to return true")
			break
		}

		if v <= previous {
			b.Errorf("expected v > %v; got %v", previous, v)
			break
		}
		previous = v
	}
}

func BenchmarkQueue(b *testing.B) {
	ctx := context.Background()
	q := newQueue(32, nil)
	id := int64(0)

	for i := 0; i < b.N; i++ {
		id++
		q.add(ctx, id)
		v, ok := q.remove()
		if !ok || v != id {
			b.Errorf("added %v; removed %v", id, v)
			break
		}
	}
}
