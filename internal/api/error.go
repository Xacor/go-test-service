package api

import (
	"net/http"

	"github.com/Xacor/go-test-service/internal/gen"
	"github.com/labstack/echo/v4"
)

func sendErrorResponse(ctx echo.Context, code int, details string) error {
	return ctx.JSON(
		code,
		&gen.ErrorResponseData{
			Error: gen.Error{
				Code:    code,
				Message: http.StatusText(code),
				Details: &details,
			},
		},
	)
}

func sendItemNotFound(ctx echo.Context) error {
	return sendErrorResponse(ctx, http.StatusNotFound, "Resource Not Found")
}
