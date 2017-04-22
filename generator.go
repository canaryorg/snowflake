package snowflake

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"sync"
	"time"
)

var (
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
)

type source interface {
	IntN(ctx context.Context, n int) ([]int64, error)
}

type config struct {
	wg         *sync.WaitGroup
	bufferSize int
	workers    int
	sources    []source
	spawnFunc  func(c *config, ids chan int64) io.Closer
}

type Option func(*config)

func spawn(ctx context.Context, fetchN int, ch chan<- int64, sources ...source) {
	for {
		source := sources[r.Intn(len(sources))]
		v, err := source.IntN(ctx, fetchN)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			time.Sleep(250*time.Millisecond + time.Duration(r.Intn(100))*time.Millisecond)
		}

		for _, id := range v {
			select {
			case <-ctx.Done():
				return
			case ch <- id:
			}
		}
	}
}

type closerFunc func()

func (fn closerFunc) Close() error {
	fn()
	return nil
}

func spawnN(c *config, ids chan int64) io.Closer {
	ctx, cancel := context.WithCancel(context.Background())

	wg := &sync.WaitGroup{}
	wg.Add(c.workers)

	fetchN := 512

	for i := 0; i < c.workers; i++ {
		go func(ctx context.Context) {
			defer wg.Done()
			spawn(ctx, fetchN, ids, c.sources...)
		}(ctx)
	}

	closer := closerFunc(func() {
		cancel()
		wg.Wait()
	})

	return closer
}

func New(opts ...Option) *Generator {
	c := &config{
		wg:         &sync.WaitGroup{},
		bufferSize: 4096,
		workers:    8,
		spawnFunc:  spawnN,
	}

	for _, opt := range opts {
		opt(c)
	}

	ids := make(chan int64, c.bufferSize)
	closer := c.spawnFunc(c, ids)

	return &Generator{
		closer: closer,
		ids:    ids,
	}
}

type Generator struct {
	closer io.Closer
	ids    chan int64
}

func (g *Generator) Close() error {
	return g.closer.Close()
}

func (g *Generator) ID() int64 {
	return <-g.ids
}

type factorySource struct {
	factory *Factory
}

func (f *factorySource) IntN(_ context.Context, n int) ([]int64, error) {
	return f.factory.IdN(n), nil
}

func WithFactory(factory *Factory) Option {
	source := &factorySource{factory: factory}

	return func(c *config) {
		c.sources = append(c.sources, source)
	}
}

func WithServers(servers ...string) Option {
	return func(g *config) {
		for _, server := range servers {
			client, err := NewClient(WithHosts(server))
			if err != nil {
				panic(fmt.Sprintf("Invalid URL, %v, provided to WithServers - %v", server, err))
			}

			g.sources = append(g.sources, client)
		}
	}
}

// WithBufferSize specifies the number of ids that may be buffered locally; beware, the larger you make this, the longer
// the startup will take
func WithBufferSize(n int) Option {
	max := 65384

	return func(c *config) {
		if n < 1 || n > max {
			panic(fmt.Sprintf("WithBufferSize must be between 1 and %v", max))
		}

		c.bufferSize = n
	}
}

// WithWorkers specifies the number of concurrent goroutines that will be fetching ids
func WithWorkers(n int) Option {
	max := 100

	return func(c *config) {
		if n < 1 || n > max {
			panic(fmt.Sprintf("WithBufferSize must be between 1 and %v", max))
		}
		c.workers = n
	}
}

func WithMonotonic() Option {
	return func(c *config) {
		c.spawnFunc = monotonic
	}
}

func monotonic(c *config, ids chan int64) io.Closer {
	ctx, cancel := context.WithCancel(context.Background())
	sourceCount := len(c.sources)

	// Logic:
	// 1. Create a bucket for each source/worker combination e.g. 3 workers per source would mean 3 buckets
	// 2. Spawn goroutines to pull for each worker independently
	// 3. Return the lowest value from each column
	buckets := c.workers * sourceCount

	wg := &sync.WaitGroup{}
	wg.Add(buckets)

	all := make([]chan int64, 0, buckets)

	for index, source := range c.sources {
		for i := 0; i < c.workers; i++ {
			bucket := index*c.workers + i
			all[bucket] = make(chan int64, 128)

			go func() {
				defer wg.Done()
				spawn(ctx, 128, all[bucket], source)
			}()
		}
	}

	closer := closerFunc(func() {
		cancel()
		wg.Wait()
	})
	return closer
}
