package mto

import (
	"time"

	"github.com/Xacor/go-test-service/internal/gen"
)

type GetGoodResponseData struct {
	Goods []Good
	Meta  MetaData
}

type Good struct {
	ID          int
	ProjectID   int
	Name        string
	Description *string
	Priority    int
	Removed     bool
	CreatedAt   time.Time
}

type MetaData struct {
	Limit   int
	Offset  int
	Removed int
	Total   int
}

func ApiGoodsFromGoods(goods []Good) *[]gen.Good {
	res := make([]gen.Good, len(goods))
	for i := range goods {
		res[i] = gen.Good{
			Id:          &goods[i].ID,
			ProjectId:   &goods[i].ProjectID,
			Name:        &goods[i].Name,
			Description: goods[i].Description,
			Priority:    &goods[i].Priority,
			Removed:     &goods[i].Removed,
			CreatedAt:   &goods[i].CreatedAt,
		}
	}

	return &res
}

func ApiGetGoodResponseDataFromGetGoodResponseData(data *GetGoodResponseData) *gen.GetGoodResponseData {
	return &gen.GetGoodResponseData{
		Meta: &gen.MetaData{
			Limit:   &data.Meta.Limit,
			Offset:  &data.Meta.Offset,
			Removed: &data.Meta.Removed,
			Total:   &data.Meta.Total,
		},
		Goods: ApiGoodsFromGoods(data.Goods),
	}
}

type CreateGood struct {
	Name      string
	ProjectID int
}

type UpdateGood struct {
	ID          int
	ProjectID   int
	Name        string
	Description *string
}

type DeleteGoodResponse struct {
	ID        int
	ProjectID int
	Removed   bool
}

type Priority struct {
	ID       int
	Priority int
}

func ApiPrioritiesFromPriorities(p []Priority) gen.ReprioritizeGoodResponseData {
	res := make(gen.ReprioritizeGoodResponseData, len(p))
	for i := range p {
		res[i].Id = &p[i].ID
		res[i].Priority = &p[i].Priority
	}

	return res
}
