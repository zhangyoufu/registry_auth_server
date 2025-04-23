package main

import "net/http"

// Cloudflare middleware replace http.Request.RemoteAddr with CF-Connecting-IP
func cloudflareRemoteAddrMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// assume CF-Connecting-IP request header is trustworthy
		ip := r.Header.Get("CF-Connecting-IP")
		if ip != "" {
			r.RemoteAddr = ip
		}
		next.ServeHTTP(w, r)
	})
}
