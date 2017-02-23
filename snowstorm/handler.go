package snowstorm

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/savaki/snowflake"
)

func writeErr(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func Handler(factory *snowflake.Factory, maxN int) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		req.ParseForm()

		n := 1
		if v := req.FormValue("n"); v != "" {
			var err error
			n, err = strconv.Atoi(v)
			if err != nil {
				writeErr(w, err)
				return
			}
			if n > maxN {
				writeErr(w, errors.New(fmt.Sprintf("exceeded the maximum number per request, %v", maxN)))
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(factory.IdN(n))
	}
}

type Statistics struct {
	ServerId int `json:"server-id"`
}

func Stats(serverId int) http.Handler {
	var handlerFunc http.HandlerFunc = func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Statistics{
			ServerId: serverId,
		})
	}

	return handlerFunc
}

func Multi(serverID, nMax int) http.HandlerFunc {
	handlers := map[string]http.Handler{}

	for srv := 1; srv <= 13; srv++ {
		for seq := 0; seq <= 13; seq++ {
			if srv+seq+41 > 64 {
				continue
			}

			func(server, sequence int) {
				path := fmt.Sprintf("/%v/%v", srv, seq)
				factory := snowflake.New(snowflake.Options{
					ServerID:     int64(serverID),
					ServerBits:   uint(server),
					SequenceBits: uint(sequence),
				})
				handlers[path] = Handler(factory, nMax)
			}(srv, seq)
		}
	}

	factory := snowflake.New(snowflake.Options{
		ServerID: int64(serverID),
	})
	handlers["/"] = Handler(factory, nMax)
	handlers["/internal/stats"] = Stats(serverID)

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
