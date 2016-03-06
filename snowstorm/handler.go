package snowstorm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/savaki/snowflake"
)

type server struct {
	factory *snowflake.Factory
	nMax    int
}

func writeErr(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func (s *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()

	n := 1
	if v := req.FormValue("n"); v != "" {
		var err error
		n, err = strconv.Atoi(v)
		if err != nil {
			writeErr(w, err)
			return
		}
		if n > s.nMax {
			writeErr(w, errors.New(fmt.Sprintf("exceeded the maximum number per request, %v", s.nMax)))
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(s.factory.IdN(n))
}

func Handler(factory *snowflake.Factory, nMax int) http.Handler {
	return &server{
		factory: factory,
		nMax:    nMax,
	}
}

func Multi(serverId, nMax int) http.HandlerFunc {
	handlers := map[string]http.Handler{}

	for srv := 1; srv <= 13; srv++ {
		for seq := 0; seq <= 13; seq++ {
			if srv+seq+41 > 64 {
				continue
			}

			func(server, sequence int) {
				path := fmt.Sprintf("/%v/%v", srv, seq)
				factory := snowflake.New(snowflake.Options{
					ServerId:     int64(serverId),
					ServerBits:   uint(server),
					SequenceBits: uint(sequence),
				})
				handlers[path] = Handler(factory, nMax)
			}(srv, seq)
		}
	}

	factory := snowflake.New(snowflake.Options{
		ServerId: int64(serverId),
	})
	handlers["/"] = Handler(factory, nMax)

	var handler http.HandlerFunc = func(w http.ResponseWriter, req *http.Request) {
		handler, ok := handlers[req.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		handler.ServeHTTP(w, req)
	}
	return handler
}
