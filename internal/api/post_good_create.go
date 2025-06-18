package api

import (
	"net/http"

	"github.com/Xacor/go-test-service/internal/gen"
	"github.com/Xacor/go-test-service/internal/model/mto"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (s *Server) PostGoodCreate(ctx echo.Context, params gen.PostGoodCreateParams) error {
	var req gen.CreateGood
	if err := ctx.Bind(&req); err != nil {
		return sendErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := ctx.Validate(req); err != nil {
		return sendErrorResponse(ctx, http.StatusBadRequest, err.Error())
	}

	good, err := s.mdl.CreateGood(ctx.Request().Context(), &mto.CreateGood{
		Name:      req.Name,
		ProjectID: params.ProjectId,
	})
	if err != nil {
		zap.L().Error("model create good", zap.Error(err))
		return sendErrorResponse(ctx, http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, gen.CreateGoodResponseData{
		Id:          &good.ID,
		ProjectId:   &good.ProjectID,
		Name:        &good.Name,
		Description: good.Description,
		Priority:    &good.Priority,
		Removed:     &good.Removed,
		CreatedAt:   &good.CreatedAt,
	})
}
