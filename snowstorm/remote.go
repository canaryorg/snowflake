package snowstorm

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
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
	resp, err := h.getFunc(ctx, http.DefaultClient, host)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	var ids []int64
	err = json.NewDecoder(resp.Body).Decode(&ids)
	return ids, err
}

func HttpFactory(hosts ...string) (RemoteFactory, error) {
	for _, host := range hosts {
		_, err := http.NewRequest("GET", host, nil)
		if err != nil {
			return nil, err
		}
	}

	return &httpFactory{
		hosts:   hosts,
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
		getFunc: ctxhttp.Get,
	}, nil
}

func WithGetFunc(remote RemoteFactory, getFunc func(ctx context.Context, client *http.Client, url string) (*http.Response, error)) RemoteFactory {
	switch v := remote.(type) {
	case *httpFactory:
		v.getFunc = getFunc
		return v
	default:
		return remote
	}
}
