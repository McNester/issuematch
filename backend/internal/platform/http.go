package platform

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

func WriteServerError(w http.ResponseWriter, err error) {
	log.Println("error:", err)
	WriteError(w, http.StatusInternalServerError, "internal error")
}

func ReadJSON(w http.ResponseWriter, r *http.Request, into any) bool {
	if err := json.NewDecoder(r.Body).Decode(into); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	return true
}
