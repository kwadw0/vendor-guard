package utils

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type EmptyData struct{}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, message string, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func WriteJSONWithMeta(w http.ResponseWriter, status int, message string, data any, meta any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_576 // 1MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(data)
}

func ErrorJSON(w http.ResponseWriter, status int, err error, code ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errCode := "INTERNAL_ERROR"
	if len(code) > 0 {
		errCode = code[0]
	} else {
		switch status {
		case http.StatusBadRequest:
			errCode = "BAD_REQUEST"
		case http.StatusNotFound:
			errCode = "NOT_FOUND"
		case http.StatusUnauthorized:
			errCode = "UNAUTHORIZED"
		case http.StatusForbidden:
			errCode = "FORBIDDEN"
		}
	}

	resp := ErrorResponse{
		Success: false,
		Message: err.Error(),
		Code:    errCode,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}
