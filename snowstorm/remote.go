package snowstorm

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

type RemoteFactory interface {
	IntN(ctx context.Context, n int) ([]int64, error)
}

type httpFactory struct {
	hosts   []string
	rand    *rand.Rand
	getFunc func(ctx context.Context, client *http.Client, url string) (*http.Response, error)
}

func (h *httpFactory) IntN(ctx context.Context, n int) ([]int64, error) {
	host := h.hosts[h.rand.Intn(len(h.hosts))]
	url := host + "?n=" + strconv.Itoa(n)
	resp, err := h.getFunc(ctx, http.DefaultClient, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var ids []int64
	err = json.NewDecoder(resp.Body).Decode(&ids)
	return ids, err
}

func HttpFactory(opts ...Option) (RemoteFactory, error) {
	h := &httpFactory{
		hosts:   []string{"http://snowflake.altairsix.com/4/13"},
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
		getFunc: ctxhttp.Get,
	}

	for _, opt := range opts {
		opt(h)
	}

	for _, host := range h.hosts {
		_, err := http.NewRequest("GET", host, nil)
		if err != nil {
			return nil, err
		}
	}

	return h, nil
}

type Option func(*httpFactory)

func GetFunc(fn func(ctx context.Context, client *http.Client, url string) (*http.Response, error)) Option {
	return func(h *httpFactory) {
		h.getFunc = fn
	}
}

func Hosts(hosts ...string) Option {
	return func(h *httpFactory) {
		h.hosts = hosts
	}
}
