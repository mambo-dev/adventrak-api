package main

// func secureHeader(next http.Handler) http.Handler {

// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("x-content-type-options", "nosniff")
// 		w.Header().Set("x-frame-options", "SAMEORIGIN")
// 		w.Header().Set("Referrer-policy", "no-referrer")

// 		next.ServeHTTP(w, r)

// 	})
// }
