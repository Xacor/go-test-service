package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/Xacor/go-test-service/internal/db/dto"
	"github.com/Xacor/go-test-service/internal/model/mto"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type dbImpl struct {
	pool *pgxpool.Pool
}

func NewDB(pool *pgxpool.Pool) DB {
	return &dbImpl{pool: pool}
}

func (d *dbImpl) GetGood(ctx context.Context, id int) (*mto.Good, error) {
	row := d.pool.QueryRow(ctx,
		`SELECT id, project_id, name, description, priority, removed, created_at
         FROM goods WHERE id = $1 AND removed = false`, id)

	var g dto.Good
	err := row.Scan(&g.ID, &g.ProjectID, &g.Name, &g.Description, &g.Priority, &g.Removed, &g.CreatedAt)
	if err != nil {
		return nil, errors.New("not found")
	}
	return dto.GoodFromDbGood(g), nil
}

func (d *dbImpl) CreateGood(ctx context.Context, input *mto.CreateGood) (*mto.Good, error) {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var priority int
	err = tx.QueryRow(ctx, `
        SELECT COALESCE(MAX(priority), 0) + 1 FROM goods
    `).Scan(&priority)
	if err != nil {
		return nil, err
	}

	var g dto.Good
	err = tx.QueryRow(ctx, `
        INSERT INTO goods (project_id, name, priority, removed, created_at)
        VALUES ($1, $2, $3, false, NOW())
        RETURNING id, project_id, name, description, priority, removed, created_at
    `,
		input.ProjectID, input.Name, priority,
	).Scan(&g.ID, &g.ProjectID, &g.Name, &g.Description, &g.Priority, &g.Removed, &g.CreatedAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return dto.GoodFromDbGood(g), nil
}

func (d *dbImpl) UpdateGood(ctx context.Context, req *mto.UpdateGood) (*mto.Good, error) {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, "SELECT id FROM goods WHERE id = $1 AND project_id = $2 AND removed = false FOR UPDATE", req.ID, req.ProjectID)
	var id int
	if err := row.Scan(&id); err != nil {
		return nil, errors.Wrap(err, "not found")
	}

	setParts := []string{"name = @name"}
	setArgs := pgx.NamedArgs{"name": req.Name}

	if req.Description != nil || *req.Description != "" {
		setParts = append(setParts, "description = @description")
		setArgs["description"] = req.Description
	}

	q := `UPDATE goods SET ` + strings.Join(setParts, ", ")
	q += ` WHERE id = @id RETURNING id, project_id, name, description, priority, removed, created_at`
	setArgs["id"] = req.ID

	var g dto.Good
	err = tx.QueryRow(ctx, q, setArgs).Scan(&g.ID, &g.ProjectID, &g.Name, &g.Description, &g.Priority, &g.Removed, &g.CreatedAt)
	if err != nil {
		return nil, errors.Wrap(err, "update good")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return dto.GoodFromDbGood(g), nil
}

func (d *dbImpl) DeleteGood(ctx context.Context, id, projectID int) (*mto.DeleteGoodResponse, error) {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, "SELECT id FROM goods WHERE id = $1 AND project_id = $2 AND removed = false FOR UPDATE", id, projectID)
	var check int
	if err := row.Scan(&check); err != nil {
		return nil, errors.Wrap(err, "not found")
	}

	var deleted mto.DeleteGoodResponse
	err = tx.QueryRow(ctx, `UPDATE goods SET removed = true WHERE id = $1 RETURNING id, project_id, removed`, id).
		Scan(&deleted.ID, &deleted.ProjectID, &deleted.Removed)
	if err != nil {
		return nil, errors.Wrap(err, "soft delete")
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &deleted, nil
}

func (d *dbImpl) ListGoodsWithMeta(ctx context.Context, limit, offset int) ([]mto.Good, *mto.MetaData, error) {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)

	var meta mto.MetaData
	err = tx.QueryRow(ctx, `
        SELECT COUNT(*) AS total,
               COUNT(*) FILTER (WHERE removed = true) AS removed
        FROM goods
    `).Scan(&meta.Total, &meta.Removed)
	if err != nil {
		return nil, nil, err
	}
	meta.Limit = limit
	meta.Offset = offset

	rows, err := tx.Query(ctx, `
        SELECT id, project_id, name, description, priority, removed, created_at
        FROM goods
        WHERE removed = false
        ORDER BY priority
        LIMIT $1 OFFSET $2
    `, limit, offset)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	goods := make([]mto.Good, 0, meta.Total)
	for rows.Next() {
		var g dto.Good
		err := rows.Scan(&g.ID, &g.ProjectID, &g.Name, &g.Description, &g.Priority, &g.Removed, &g.CreatedAt)
		if err != nil {
			return nil, nil, err
		}
		goods = append(goods, *dto.GoodFromDbGood(g))
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, err
	}

	return goods, &meta, nil
}

func (d *dbImpl) ReprioritizeGood(ctx context.Context, id, projectID, newPriority int) ([]mto.Priority, error) {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var currentPriority int
	err = tx.QueryRow(ctx, `
        SELECT priority FROM goods
        WHERE id = $1 AND project_id = $2 AND removed = false
        FOR UPDATE
    `, id, projectID).Scan(&currentPriority)
	if err == pgx.ErrNoRows {
		return nil, errors.Wrap(err, "not found")
	}
	if err != nil {
		return nil, err
	}

	if currentPriority == newPriority {
		return nil, tx.Commit(ctx)
	}

	if newPriority < currentPriority {
		_, err = tx.Exec(ctx, `
            UPDATE goods
            SET priority = priority + 1
            WHERE project_id = $1 AND removed = false AND priority >= $2 AND priority < $3
        `, projectID, newPriority, currentPriority)
	} else {
		_, err = tx.Exec(ctx, `
            UPDATE goods
            SET priority = priority - 1
            WHERE project_id = $1 AND removed = false AND priority > $2 AND priority <= $3
        `, projectID, currentPriority, newPriority)
	}
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
        UPDATE goods SET priority = $1
        WHERE id = $2 AND project_id = $3
    `, newPriority, id, projectID)
	if err != nil {
		return nil, err
	}

	fmt.Printf("currentPriority: %v (%T)\n", currentPriority, currentPriority)
	fmt.Printf("newPriority: %v (%T)\n", newPriority, newPriority)

	q := "SELECT id, priority FROM goods WHERE project_id = $1 AND removed = false "
	q += "AND priority BETWEEN LEAST($2::int, $3::int) AND GREATEST($2::int, $3::int) ORDER BY priority"
	rows, err := tx.Query(ctx, q, projectID, currentPriority, newPriority)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []mto.Priority
	for rows.Next() {
		var p mto.Priority
		if err := rows.Scan(&p.ID, &p.Priority); err != nil {
			return nil, err
		}
		result = append(result, p)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return result, nil
}
