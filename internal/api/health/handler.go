package health

import (
	"encoding/json"
	"net/http"
)

// Handler returns an HTTP handler for the health check endpoint.
// This endpoint is used by orchestrators (Kubernetes, Docker) for liveness/readiness probes.
func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(map[string]string{
				"status": "ok",
			})
		}
	}
}
