package snowstorm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
)

type RemoteFactory interface {
	IntN(ctx context.Context, n int) ([]int64, error)
}

type httpFactory struct {
	hosts     []string
	hostCount int32
	offset    int32
	doFunc    func(r *http.Request) (*http.Response, error)
}

func (h *httpFactory) IntN(ctx context.Context, n int) ([]int64, error) {
	host := h.hosts[int(h.offset%h.hostCount)]
	atomic.AddInt32(&h.offset, 1)
	if h.offset > h.hostCount {
		atomic.StoreInt32(&h.offset, 0)
	}

	url := host + "?n=" + strconv.Itoa(n)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := h.doFunc(req)
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
		hosts:  []string{"http://snowflake.altairsix.com/10/13"},
		doFunc: http.DefaultClient.Do,
	}

	for _, opt := range opts {
		opt(h)
	}

	h.hostCount = int32(len(h.hosts))

	// ensure the hosts are all valid
	for _, host := range h.hosts {
		_, err := http.NewRequest("GET", host, nil)
		if err != nil {
			return nil, err
		}
	}

	return h, nil
}

type Option func(*httpFactory)

func DoFunc(fn func(r *http.Request) (*http.Response, error)) Option {
	return func(h *httpFactory) {
		h.doFunc = fn
	}
}

func Hosts(hosts ...string) Option {
	return func(h *httpFactory) {
		h.hosts = hosts
	}
}
