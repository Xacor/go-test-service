package api

import (
	"net/http"

	"github.com/Xacor/go-test-service/internal/gen"
	"github.com/Xacor/go-test-service/internal/model/mto"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (s *Server) PatchGoodReprioritize(ctx echo.Context, params gen.PatchGoodReprioritizeParams) error {
	var req gen.PatchGoodReprioritizeJSONBody
	if err := ctx.Bind(&req); err != nil {
		return sendErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := ctx.Validate(req); err != nil {
		return err
	}

	updatedPriorities, err := s.mdl.ReprioritizeGood(ctx.Request().Context(), params.Id, params.ProjectId, req.NewPriority)
	if err != nil {
		zap.L().Error("model reprioritize good", zap.Error(err))
		return sendErrorResponse(ctx, http.StatusInternalServerError, err.Error())
	}

	if updatedPriorities == nil {
		return sendItemNotFound(ctx)
	}

	return ctx.JSON(http.StatusOK, mto.ApiPrioritiesFromPriorities(updatedPriorities))
}
