package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/savaki/snowflake/snowstorm"
)

type UserData struct {
	ServerId int `json:"server-id"`
}

type Options struct {
	ServerId int
	Port     int
	AWS      bool
}

func main() {
	var (
		flagServerID = flag.Int("id", 0, "unique server id")
		flagPort     = flag.Int("port", 7006, "port to list on")
		flagNoAWS    = flag.Bool("no-aws", false, "do not attempt to retrieve server id from user data json")
	)
	flag.Parse()

	opts := Options{
		ServerId: *flagServerID,
		Port:     *flagPort,
		AWS:      !*flagNoAWS,
	}

	Run(opts)
}

func Run(opts Options) {
	serverID := opts.ServerId

	if opts.AWS {
		ctx, _ := context.WithTimeout(context.Background(), time.Second*3)
		req, err := http.NewRequest("GET", "http://169.254.169.254/latest/user-data", nil)
		if err != nil {
			log.Fatalln(err)
		}
		req = req.WithContext(ctx)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			config := UserData{}
			if err = json.NewDecoder(resp.Body).Decode(&config); err == nil {
				serverID = config.ServerId
			}
		}
	}

	handler := snowstorm.Multi(serverID, 512)
	http.ListenAndServe(fmt.Sprintf(":%v", opts.Port), handler)
}
