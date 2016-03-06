package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
	"encoding/json"

	"github.com/codegangsta/cli"
	"github.com/savaki/snowflake/snowstorm"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

type UserData struct {
	ServerId int `json:"server-id"`
}

type Options struct {
	ServerId int
	Port     int
	AWS      bool
}

var opts Options

func main() {
	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.IntFlag{"id", 0, "unique server id", "SERVER_ID", &opts.ServerId},
		cli.IntFlag{"port", 7006, "port to list on", "PORT", &opts.Port},
		cli.BoolTFlag{"no-aws", "do not attempt to retrieve server id from user data json", "NO_AWS", &opts.AWS},
	}
	app.Action = run
	app.Run(os.Args)
}

func run(c *cli.Context) {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
	resp, err := ctxhttp.Get(ctx, http.DefaultClient, "http://169.254.169.254/latest/user-data")
	if err == nil {
		config := UserData{}
		if err = json.NewDecoder(resp.Body).Decode(&config); err == nil {
			opts.ServerId = config.ServerId
		}
	}

	handler := snowstorm.Multi(opts.ServerId, 512)
	http.ListenAndServe(fmt.Sprintf(":%v", opts.Port), handler)
}
