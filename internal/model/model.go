package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Xacor/go-test-service/internal/db"
	"github.com/Xacor/go-test-service/internal/model/mto"
	"github.com/Xacor/go-test-service/pkg/log"
	"github.com/jackc/pgx/v5"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type goodsModelImpl struct {
	redis *redis.Client
	js    jetstream.JetStream
	repo  db.DB
}

func New(redis *redis.Client, js jetstream.JetStream, db db.DB) *goodsModelImpl {
	return &goodsModelImpl{
		redis: redis,
		js:    js,
		repo:  db,
	}
}

func (m *goodsModelImpl) CreateGood(ctx context.Context, req *mto.CreateGood) (*mto.Good, error) {
	iter := m.redis.Scan(ctx, 0, "goods:list:*", 0).Iterator()
	for iter.Next(ctx) {
		_ = m.redis.Del(ctx, iter.Val()).Err()
	}

	good, err := m.repo.CreateGood(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed go create good")
	}

	err = m.sendLogEntry(ctx, log.LogEntry{
		ID:          good.ID,
		ProjectID:   good.ProjectID,
		Name:        good.Name,
		Description: good.Description,
		Priority:    good.Priority,
		Removed:     good.Removed,
		EventTime:   time.Now().UTC(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "publish log")
	}

	return good, nil
}

func (m *goodsModelImpl) UpdateGood(ctx context.Context, req *mto.UpdateGood) (*mto.Good, error) {
	iter := m.redis.Scan(ctx, 0, "goods:list:*", 0).Iterator()
	for iter.Next(ctx) {
		_ = m.redis.Del(ctx, iter.Val()).Err()
	}

	good, err := m.repo.UpdateGood(ctx, req)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	err = m.sendLogEntry(ctx, log.LogEntry{
		ID:          good.ID,
		ProjectID:   good.ProjectID,
		Name:        good.Name,
		Description: good.Description,
		Priority:    good.Priority,
		Removed:     good.Removed,
		EventTime:   time.Now().UTC(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "publish log")
	}

	return good, nil
}

func (m *goodsModelImpl) DeleteGood(ctx context.Context, id, projectID int) (*mto.DeleteGoodResponse, error) {
	iter := m.redis.Scan(ctx, 0, "goods:list:*", 0).Iterator()
	for iter.Next(ctx) {
		_ = m.redis.Del(ctx, iter.Val()).Err()
	}

	good, err := m.repo.GetGood(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	deleted, err := m.repo.DeleteGood(ctx, id, projectID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	err = m.sendLogEntry(ctx, log.LogEntry{
		ID:          good.ID,
		ProjectID:   good.ProjectID,
		Name:        good.Name,
		Description: good.Description,
		Priority:    good.Priority,
		Removed:     deleted.Removed,
		EventTime:   time.Now().UTC(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "publish log")
	}

	return deleted, nil
}

func (m *goodsModelImpl) ListGoods(ctx context.Context, limit, offset int) (*mto.GetGoodResponseData, error) {
	cacheKey := listGoodsCacheKey(limit, offset)
	cached, err := m.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var resp *mto.GetGoodResponseData
		if err := json.Unmarshal([]byte(cached), resp); err == nil {
			return resp, nil
		}
	}

	goods, meta, err := m.repo.ListGoodsWithMeta(ctx, limit, offset)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	resp := &mto.GetGoodResponseData{
		Meta:  *meta,
		Goods: goods,
	}

	data, err := json.Marshal(resp)
	if err == nil {
		_ = m.redis.Set(ctx, cacheKey, data, time.Minute).Err()
	}

	return resp, nil
}

func (m *goodsModelImpl) ReprioritizeGood(ctx context.Context, id, projectID, newPriority int) ([]mto.Priority, error) {
	iter := m.redis.Scan(ctx, 0, "goods:list:*", 0).Iterator()
	for iter.Next(ctx) {
		_ = m.redis.Del(ctx, iter.Val()).Err()
	}

	updated, err := m.repo.ReprioritizeGood(ctx, id, projectID, newPriority)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	} else if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}

	for i := range updated {
		good, err := m.repo.GetGood(ctx, updated[i].ID)
		if err != nil {
			return nil, err
		}

		err = m.sendLogEntry(ctx, log.LogEntry{
			ID:          good.ID,
			ProjectID:   good.ProjectID,
			Name:        good.Name,
			Description: good.Description,
			Priority:    good.Priority,
			Removed:     good.Removed,
			EventTime:   time.Now().UTC(),
		})
		if err != nil {
			return nil, errors.Wrap(err, "publish log")
		}

	}

	return updated, nil
}

func listGoodsCacheKey(limit, offset int) string {
	return fmt.Sprintf("goods:list:limit=%d:offset=%d", limit, offset)
}

func (m *goodsModelImpl) sendLogEntry(ctx context.Context, entry log.LogEntry) error {
	const subject = "logs.goods"
	if m.js == nil {
		return fmt.Errorf("JetStream not initialized")
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	_, err = m.js.Publish(ctx, subject, data)
	return err
}
