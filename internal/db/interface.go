package db

import (
	"context"

	"github.com/Xacor/go-test-service/internal/model/mto"
)

type DB interface {
	CreateGood(ctx context.Context, req *mto.CreateGood) (*mto.Good, error)
	GetGood(ctx context.Context, id int) (*mto.Good, error)
	UpdateGood(ctx context.Context, req *mto.UpdateGood) (*mto.Good, error)
	DeleteGood(ctx context.Context, id, projectID int) (*mto.DeleteGoodResponse, error)
	ListGoodsWithMeta(ctx context.Context, limit, offset int) ([]mto.Good, *mto.MetaData, error)
	ReprioritizeGood(ctx context.Context, id, projectID, newPriority int) ([]mto.Priority, error)
}
