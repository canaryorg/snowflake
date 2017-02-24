package main

import (
	"fmt"

	"github.com/savaki/snowflake"
)

func main() {
	client, _ := snowflake.NewClient(snowflake.WithHosts("your-host"))
	buffered := snowflake.NewBufferedClient(client)
	fmt.Println("id:", buffered.Id())
}
