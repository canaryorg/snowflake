package snowflake_test

import (
	"context"
	"testing"
	"time"

	"github.com/savaki/snowflake"
)

type Remote struct {
	factory *snowflake.Factory
}

func (t *Remote) IntN(ctx context.Context, n int) ([]int64, error) {
	return t.factory.IdN(n), nil
}

func TestGenerateIdStream(t *testing.T) {
	buffer := 4
	client := snowflake.NewBufferedClient(&Remote{snowflake.Mock})

	uniques := map[int64]int64{}
	iterations := buffer * 10
	for i := 0; i < iterations; i++ {
		id := client.Id()
		uniques[id] = id
	}
	client.Close()

	if v := len(uniques); v != iterations {
		t.Errorf("expected %v; got %v\n", iterations, v)
	}
}

type Mock struct {
	count int
}

func (m *Mock) IntN(ctx context.Context, n int) ([]int64, error) {
	now := time.Now().UnixNano()

	ids := make([]int64, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, now+int64(i))
	}

	m.count += n
	return ids, nil
}
