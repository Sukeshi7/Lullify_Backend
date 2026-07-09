package http

import "net/http"

func NewRouter(auth *AuthHandler) http.Handler {
	mux := http.NewServeMux()
	auth.Routes(mux)
	return mux
}
