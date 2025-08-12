package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"wb_tech_level_zero/internal/dto"
	"wb_tech_level_zero/pkg/logger"

	"go.uber.org/zap"
)

func (h *Handlers) writeJSONResponse(ctx context.Context, w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log := logger.GetLoggerFromCtx(ctx)
		log.Error(ctx, "Failed to encode and write JSON response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handlers) writeErrorResponse(ctx context.Context, w http.ResponseWriter, statusCode int, message string) {
	// log := logger.GetLoggerFromCtx(ctx)
	// log.Info(ctx, message)

	response := dto.ErrorResponse{Message: message}
	h.writeJSONResponse(ctx, w, statusCode, response)
}
