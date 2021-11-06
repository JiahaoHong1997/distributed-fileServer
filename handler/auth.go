package handler

import "net/http"

// HTTPInterceptor: http请求拦截器
func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			userName := r.Form.Get("username")
			token := r.Form.Get("token")

			if len(userName)<3 || !IsTokenValid(token) {
				w.WriteHeader(http.StatusForbidden)
			}
			h(w,r)
		})
}
