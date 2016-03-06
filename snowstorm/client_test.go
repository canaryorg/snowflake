package snowstorm_test

import (
	"testing"

	"github.com/savaki/snowflake"
	"github.com/savaki/snowflake/snowstorm"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
)

type Remote struct {
	factory *snowflake.Factory
}

func (t *Remote) IntN(ctx context.Context, n int) ([]int64, error) {
	return t.factory.IdN(n), nil
}

func TestClient(t *testing.T) {
	Convey("Verify client can generate stream of ids", t, func() {
		buffer := 4
		client := snowstorm.New(buffer, &Remote{snowflake.Default})

		uniques := map[int64]int64{}
		iterations := buffer * 10
		for i := 0; i < iterations; i++ {
			id := client.Id()
			uniques[id] = id
		}
		client.Close()

		So(len(uniques), ShouldEqual, iterations)
	})
}
