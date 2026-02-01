package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func New(target string) http.Handler {
	u, _ := url.Parse(target)
	p := httputil.NewSingleHostReverseProxy(u)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.ServeHTTP(w, r)
	})
}
