package snowstorm_test

import (
	"net/http"
	"testing"

	"github.com/savaki/mockhttp"
	"github.com/savaki/snowflake"
	"github.com/savaki/snowflake/snowstorm"
	"golang.org/x/net/context"
)

func TestHttpFactory(t *testing.T) {
	handler := snowstorm.Handler(snowflake.Default, 16)
	router := http.NewServeMux()
	router.Handle("/ids", handler)

	app := mockhttp.New(handler)

	remoteFactory, err := snowstorm.HttpFactory("http://localhost/ids")
	if err != nil {
		t.Error("Unable to create HttpFactory")
	}
	remoteFactory = snowstorm.WithGetFunc(remoteFactory, func(ctx context.Context, client *http.Client, url string) (*http.Response, error) {
		return app.GET(url)
	})

	client := snowstorm.New(16, remoteFactory)
	defer client.Close()

	uniques := map[int64]int64{}
	iterations := 100
	for i := 0; i < iterations; i++ {
		id := client.Id()
		uniques[id] = id
	}

	if v := len(uniques); v != iterations {
		t.Errorf("expected %v; got %v\n", iterations, v)
	}
}
