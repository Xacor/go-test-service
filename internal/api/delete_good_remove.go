package api

import (
	"net/http"

	"github.com/Xacor/go-test-service/internal/gen"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (s *Server) DeleteGoodRemove(ctx echo.Context, params gen.DeleteGoodRemoveParams) error {
	deleted, err := s.mdl.DeleteGood(ctx.Request().Context(), params.Id, params.ProjectId)
	if err != nil {
		zap.L().Error("model delete good", zap.Error(err))
		return sendErrorResponse(ctx, http.StatusInternalServerError, err.Error())
	}

	if deleted == nil {
		return sendItemNotFound(ctx)
	}

	return ctx.JSON(http.StatusOK, gen.RemoveGoodResponseData{
		Id:        &deleted.ID,
		ProjectId: &deleted.ProjectID,
		Removed:   &deleted.Removed,
	})
}
