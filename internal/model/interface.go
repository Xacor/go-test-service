package model

import (
	"context"

	"github.com/Xacor/go-test-service/internal/model/mto"
)

type GoodsModel interface {
	CreateGood(ctx context.Context, req *mto.CreateGood) (*mto.Good, error)
	UpdateGood(ctx context.Context, req *mto.UpdateGood) (*mto.Good, error)
	DeleteGood(ctx context.Context, id, projectID int) (*mto.DeleteGoodResponse, error)
	ListGoods(ctx context.Context, limit, offset int) (*mto.GetGoodResponseData, error)
	ReprioritizeGood(ctx context.Context, id, projectID, newPriority int) ([]mto.Priority, error)
}
