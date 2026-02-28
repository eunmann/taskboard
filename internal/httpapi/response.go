package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/OWNER/PROJECT_NAME/internal/platform/logging"
)

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
	}
}

// WriteError writes a 400 error response.
func WriteError(w http.ResponseWriter, message string) {
	WriteErrorWithStatus(w, http.StatusBadRequest, message)
}

// WriteErrorWithStatus writes an error response with a custom status code.
func WriteErrorWithStatus(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		_, _ = w.Write([]byte(`{"error":"internal error"}`))
	}
}

// WriteInternalError logs the server error and returns a generic message.
func WriteInternalError(w http.ResponseWriter, r *http.Request, err error, clientMessage string) {
	logger := logging.LoggerFrom(r.Context())
	logger.Error().Err(err).Msg(clientMessage)
	WriteErrorWithStatus(w, http.StatusInternalServerError, clientMessage)
}
