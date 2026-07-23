package http

import "net/http"

func NewRouter(auth *AuthHandler, stream *StreamHandler) http.Handler {
	mux := http.NewServeMux()
	auth.Routes(mux)
	stream.Routes(mux)
	return mux
}
