package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type tenantKey string

const TenantIDKey tenantKey = "tenant_id"

func TenantFromHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantIDStr := r.Header.Get("X-Tenant-ID")
		if tenantIDStr == "" {
			http.Error(w, `{"success":false,"error":"X-Tenant-ID header is required"}`, http.StatusBadRequest)
			return
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			http.Error(w, `{"success":false,"error":"invalid X-Tenant-ID format"}`, http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func TenantIDFromContext(ctx context.Context) uuid.UUID {
	id, _ := ctx.Value(TenantIDKey).(uuid.UUID)
	return id
}
