package snowstorm

import (
	"math/rand"
	"sync"
	"time"

	"golang.org/x/net/context"
)

type Client struct {
	factory RemoteFactory
	ch      chan int64
	ctx     context.Context
	cancel  func()
	wg      *sync.WaitGroup
	n       int
}

func (c *Client) Id() int64 {
	return <-c.ch
}

func (c *Client) start() {
	defer c.wg.Done()

	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
		ids, err := c.factory.IntN(ctx, c.n)
		if err != nil {
			// if there was a problem, hang out for a little bit before trying again
			time.Sleep(250*time.Millisecond + time.Duration(rand.Intn(100))*time.Millisecond)
			continue
		}

		for _, id := range ids {
			c.ch <- id
		}

		select {
		case <-c.ctx.Done():
			return
		default:
		}
	}
}

func (c *Client) Close() {
	c.cancel()
	c.wg.Wait()
}

func New(buffer int, factory RemoteFactory) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		ctx:     ctx,
		cancel:  cancel,
		factory: factory,
		ch:      make(chan int64, buffer),
		wg:      &sync.WaitGroup{},
		n:       buffer / 3,
	}

	// start background job to pull content
	c.wg.Add(1)
	go c.start()

	return c
}
