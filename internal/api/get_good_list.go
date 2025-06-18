package api

import (
	"net/http"

	"github.com/Xacor/go-test-service/internal/gen"
	"github.com/Xacor/go-test-service/internal/model/mto"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func (s *Server) GetGoodList(ctx echo.Context, params gen.GetGoodListParams) error {
	const (
		defaultLimit  = 10
		defaultOffset = 0
	)
	var limit, offset int
	if params.Limit != nil {
		limit = *params.Limit
	} else {
		limit = defaultLimit
	}

	if params.Offset != nil {
		offset = *params.Offset
	} else {
		offset = defaultOffset
	}

	goodsData, err := s.mdl.ListGoods(ctx.Request().Context(), limit, offset)
	if err != nil {
		zap.L().Error("model get goods", zap.Error(err))
		return sendErrorResponse(ctx, http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, mto.ApiGetGoodResponseDataFromGetGoodResponseData(goodsData))
}
