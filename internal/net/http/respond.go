package http

import (
	"encoding/json"
	"net/http"

	"go.adoublef.dev/runtime/debug"
)

func respond[V any](w http.ResponseWriter, _ *http.Request, v V, code int) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(code)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		debug.Printf(`%v = json.NewEncoder(w).Encode(%#v)`, err, v)
	}
}
