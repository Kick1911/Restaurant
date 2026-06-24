package response

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Page    int         `json:"page"`
	Limit   int         `json:"limit"`
	Total   int         `json:"total"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{Success: true, Data: data})
}

func Error(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{Success: false, Error: message})
}

func ValidationError(w http.ResponseWriter, err error) {
	var errors []string
	if verrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range verrs {
			errors = append(errors, e.Field()+": "+e.Tag())
		}
	}
	if len(errors) == 0 {
		errors = append(errors, err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(APIResponse{Success: false, Error: "validation failed", Data: errors})
}

func Paginated(w http.ResponseWriter, data interface{}, page, limit, total int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(PaginatedResponse{
		Success: true,
		Data:    data,
		Page:    page,
		Limit:   limit,
		Total:   total,
	})
}
