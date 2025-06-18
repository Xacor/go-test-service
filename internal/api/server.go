package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Xacor/go-test-service/internal/gen"
	"github.com/Xacor/go-test-service/internal/model"
	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Server struct {
	e   *echo.Echo
	mdl model.GoodsModel
}

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	if err := v.validator.Struct(i); err != nil {
		return err
	}
	return nil
}

func NewServer(mdl model.GoodsModel) *Server {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Validator = &Validator{validator: validator.New()}

	s := &Server{
		e:   e,
		mdl: mdl,
	}

	gen.RegisterHandlers(e, s)

	return s
}

func (s *Server) Listen() error {
	s.e.Server.ReadTimeout = time.Second * 10
	s.e.Server.WriteTimeout = time.Second * 10
	s.e.Server.Addr = ":8080"
	go func() {
		if err := s.e.StartServer(s.e.Server); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.e.Logger.Fatal("start", zap.Error(err))
		}
	}()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.e.Shutdown(ctx)
}
