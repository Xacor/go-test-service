package dto

import (
	"database/sql"
	"time"

	"github.com/Xacor/go-test-service/internal/model/mto"
)

type Good struct {
	ID          int
	ProjectID   int
	Name        string
	Description sql.NullString
	Priority    int
	Removed     bool
	CreatedAt   time.Time
}

func DbGoodFromGood(good *mto.Good) *Good {
	return &Good{
		ID:          good.ID,
		ProjectID:   good.ProjectID,
		Name:        good.Name,
		Description: sql.NullString{String: *good.Description, Valid: true},
		Priority:    good.Priority,
		Removed:     good.Removed,
		CreatedAt:   good.CreatedAt,
	}
}

func GoodFromDbGood(good Good) *mto.Good {
	var desc *string
	if good.Description.Valid {
		desc = &good.Description.String
	}
	return &mto.Good{
		ID:          good.ID,
		ProjectID:   good.ProjectID,
		Name:        good.Name,
		Description: desc,
		Priority:    good.Priority,
		Removed:     good.Removed,
		CreatedAt:   good.CreatedAt,
	}
}

func GoodsFromDbGoods(goods []Good) *[]mto.Good {
	res := make([]mto.Good, len(goods))
	for i := range goods {
		res[i] = *GoodFromDbGood(goods[i])
	}
	return &res
}
