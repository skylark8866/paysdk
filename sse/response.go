package sse

import (
	"encoding/json"
	"net/http"
)

func setSSEHeaders(w http.ResponseWriter) {
	w.Header().Set(HeaderContentType, SSEContentType)
	w.Header().Set(HeaderCacheControl, SSECacheControl)
	w.Header().Set(HeaderConnection, SSEConnection)
	w.Header().Set(HeaderACAO, SSEAllowOrigin)
}

func writeJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set(HeaderContentType, JSONContentType)
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{RespFieldError: message})
}
