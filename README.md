# snowflake

Snowflake is a go implementation of Twitter's Snowflake service.  Like its namespace, Snowflake is also a network
service for generating unique ID numbers at high scale with some simple guarantees.

[https://blog.twitter.com/2010/announcing-snowflake](https://blog.twitter.com/2010/announcing-snowflake)

## Concepts

Snowflake generates unique int64 ids that (unlike uuids) are loosely time sorted.  Each id consists of:

| Bits  | Field | Notes |
| :--- | :--- | :--- |
| 41 | Timestamp in MS | ~70yrs |
| 10 | Server ID | Unique Server ID |
| 13 | Sequence ID| sequence to disambiguate requests in the same ms |

## Server Usage

The simplest way to run snowflake is via docker:

```
docker run -p 80:80 -it --rm savaki/snowflake:1.3
```

To retrieve the a single id:

```
curl http://your-host-name?n=4
[152193159915372544]
```

To retrieve the N ids:

```
curl http://your-host-name?n=8
[152193295848570880,152193295848570881,152193295848570882,152193295848570883]
```

## Client Usage

Snowflake implements two clients, a low level client, and a high level buffered client.  In most cases, you'll want to 
use the buffered client.  The buffered client maintains and replenishes an internal queue of ids so there should always 
be one available when you need it.

```
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
```

