package utils

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/sha3"
)

func GetFunctionSelector(signature string) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(signature))
	hash := hasher.Sum(nil)
	return hash[:4]
}

func DecodeJSON(r *http.Request, dst any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Prevent unknown fields
	return decoder.Decode(dst)
}

func RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func RespondWithError(w http.ResponseWriter, code int, message string) {
	RespondWithJSON(w, code, map[string]string{"error": message})
}
