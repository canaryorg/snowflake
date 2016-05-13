package snowstorm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"

	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

type RemoteFactory interface {
	IntN(ctx context.Context, n int) ([]int64, error)
}

type httpFactory struct {
	hosts     []string
	hostCount int32
	offset    int32
	getFunc   func(ctx context.Context, client *http.Client, url string) (*http.Response, error)
}

func (h *httpFactory) IntN(ctx context.Context, n int) ([]int64, error) {
	host := h.hosts[int(h.offset%h.hostCount)]
	atomic.AddInt32(&h.offset, 1)
	if h.offset > h.hostCount {
		atomic.StoreInt32(&h.offset, 0)
	}

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
		getFunc: ctxhttp.Get,
	}

	for _, opt := range opts {
		opt(h)
	}

	h.hostCount = int32(len(h.hosts))

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
