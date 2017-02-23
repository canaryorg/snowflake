package snowstorm

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	RequestSize = 512
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

func (c *Client) spawnN(n int) {
	c.wg.Add(n)
	for i := 0; i < n; i++ {
		go c.spawn()
	}
}

func (c *Client) spawn() {
	defer c.wg.Done()

	for {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
		ids, err := c.factory.IntN(ctx, RequestSize)
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

func (c *Client) Close() {
	c.cancel()
	c.wg.Wait()
}

func New(factory RemoteFactory) *Client {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		ctx:     ctx,
		cancel:  cancel,
		factory: factory,
		ch:      make(chan int64, 4096),
		wg:      &sync.WaitGroup{},
	}

	c.spawnN(8)

	return c
}
