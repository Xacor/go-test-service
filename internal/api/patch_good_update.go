package api

import (
	"net/http"

	"github.com/Xacor/go-test-service/internal/gen"
	"github.com/Xacor/go-test-service/internal/model/mto"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (s *Server) PatchGoodUpdate(ctx echo.Context, params gen.PatchGoodUpdateParams) error {
	var req gen.UpdateGood
	if err := ctx.Bind(&req); err != nil {
		return sendErrorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	if err := ctx.Validate(req); err != nil {
		return err
	}

	good, err := s.mdl.UpdateGood(ctx.Request().Context(), &mto.UpdateGood{
		ID:          params.Id,
		ProjectID:   params.ProjectId,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		zap.L().Error("model update good", zap.Error(err))
		return sendErrorResponse(ctx, http.StatusInternalServerError, err.Error())
	}

	if good == nil {
		return sendItemNotFound(ctx)
	}

	return ctx.JSON(http.StatusOK, gen.UpdateGoodResponseData{
		Id:          &good.ID,
		ProjectId:   &good.ProjectID,
		Name:        &good.Name,
		Description: good.Description,
		Priority:    &good.Priority,
		Removed:     &good.Removed,
		CreatedAt:   &good.CreatedAt,
	})
}
