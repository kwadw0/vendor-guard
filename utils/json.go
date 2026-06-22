package utils

import (
	"encoding/json"
	"net/http"
)

type EmptyData struct{}

type SuccessResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func WriteJSON(w http.ResponseWriter, status int, message string, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := SuccessResponse[any]{
		Success: true,
		Message: message,
		Data:    data,
	}

	return json.NewEncoder(w).Encode(resp)
}

func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_576 // 1MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(data)
}

func ErrorJSON(w http.ResponseWriter, status int, err error, code ...string) error {
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
	return json.NewEncoder(w).Encode(resp)
}
