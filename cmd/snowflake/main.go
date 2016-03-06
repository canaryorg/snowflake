package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/codegangsta/cli"
	"github.com/savaki/snowflake/snowstorm"
)

type Options struct {
	ServerId int
	Port     int
}

var opts Options

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.IntFlag{"id", 0, "unique server id", "SERVER_ID", &opts.ServerId},
		cli.IntFlag{"port", 7006, "port to list on", "PORT", &opts.Port},
	}
	app.Action = run
	app.Run(os.Args)
}

func run(c *cli.Context) {
	handler := snowstorm.Multi(opts.ServerId, 512)
	http.ListenAndServe(fmt.Sprintf(":%v", opts.Port), handler)
}
