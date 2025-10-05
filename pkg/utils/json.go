package utils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func ParseJSON(r *http.Request, maxBytes int64, input any) error {
	if maxBytes == 0 {
		maxBytes = 100_000
	}

	data, err := io.ReadAll(io.LimitReader(r.Body, maxBytes))
	if err != nil {
		return err
	}

	return json.Unmarshal(data, input)
}

func WriteJSON(w http.ResponseWriter, statusCode int, message any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	switch value := message.(type) {
	case error:
		message = value.Error()
	}

	if err := json.NewEncoder(w).Encode(message); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fallback := map[string]string{
			"error": "failed to encode the response",
		}

		if err := json.NewEncoder(w).Encode(fallback); err != nil {
			log.Println(err)
		}
	}
}
