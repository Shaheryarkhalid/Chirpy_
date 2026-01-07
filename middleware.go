package main
import "net/http"


func middlewareMetricsInc(apiConfig *ApiConfig, next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiConfig.fileserverHits.Add(1)
		next.ServeHTTP(w, r);
	})
}

