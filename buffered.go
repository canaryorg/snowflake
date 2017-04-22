package snowflake

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	defaultRequestSize = 512
)

// BufferedClient implements a client that internally buffers ids for performance purposes
type BufferedClient struct {
	client Client
	ch     chan int64
	ctx    context.Context
	cancel func()
	wg     *sync.WaitGroup
}

func (c *BufferedClient) Id() int64 {
	return <-c.ch
}

func (c *BufferedClient) spawnN(n int) {
	c.wg.Add(n)
	for i := 0; i < n; i++ {
		go c.spawn()
	}
}

func (c *BufferedClient) spawn() {
	defer c.wg.Done()

	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
		ids, err := c.client.IntN(ctx, defaultRequestSize)
		if err != nil {
			select {
			case <-c.ctx.Done():
				return
			default:
			}

			// if there was a problem, hang out for a little bit before trying again
			fmt.Fprintln(os.Stderr, err)
			time.Sleep(250*time.Millisecond + time.Duration(rand.Intn(100))*time.Millisecond)
			continue
		}

		for _, id := range ids {
			select {
			case c.ch <- id:
			case <-c.ctx.Done():
				return
			}
		}
	}
}

func (c *BufferedClient) Close() {
	c.cancel()
	c.wg.Wait()
}

func NewBufferedClient(c Client, opts ...BufferedClientOption) *BufferedClient {
	ctx, cancel := context.WithCancel(context.Background())
	bc := &BufferedClient{
		ctx:    ctx,
		cancel: cancel,
		client: c,
		ch:     make(chan int64, 4096),
		wg:     &sync.WaitGroup{},
	}

	for _, opt := range opts {
		opt(bc)
	}

	bc.spawnN(8)

	return bc
}

type BufferedClientOption func(client *BufferedClient)

func WithBufferSize(size int64) BufferedClientOption {
	max := int64(65384)

	return func(client *BufferedClient) {
		if size < 0 || size > max {
			panic(fmt.Sprintf("WithBufferSize must be between 0 and %v", max))
		}

		if client.ch != nil {
			close(client.ch)
		}

		client.ch = make(chan int64, size)
	}
}
