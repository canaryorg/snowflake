package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/savaki/snowflake/snowstorm"
)

type UserData struct {
	ServerId int `json:"server-id"`
}

type Options struct {
	ServerID int
	Port     int
}

func main() {
	var (
		flagServerID = flag.Int("id", 0, "unique server id")
		flagPort     = flag.Int("port", 7006, "port to list on")
	)
	flag.Parse()

	opts := Options{
		ServerID: *flagServerID,
		Port:     *flagPort,
	}

	Run(opts)
}

func Run(opts Options) {
	serverID := opts.ServerID

	handler := snowstorm.Multi(serverID, 512)
	http.ListenAndServe(fmt.Sprintf(":%v", opts.Port), handler)
}
