package aggregator

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mdoroguncev/toll-calculator/internal/types"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

var _ Store = (*PostgresStore)(nil)

func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStore{
		pool: pool,
	}, nil
}

func (s *PostgresStore) Insert(ctx context.Context, dist types.Distance) error {
	query := "INSERT INTO distance (obu_id, value, unix) VALUES ($1,$2,$3)"

	if _, err := s.pool.Exec(ctx, query, dist.OBUID, dist.Value, dist.Unix); err != nil {
		return fmt.Errorf("insert distance error %w", err)
	}

	return nil
}

func (s *PostgresStore) Get(ctx context.Context, id int) (float64, error) {
	var total float64
	var count int64
	query := "SELECT COALESCE(SUM(value), 0), COUNT(*) FROM distance WHERE obu_id = $1"

	if err := s.pool.QueryRow(ctx, query, id).Scan(&total, &count); err != nil {
		return 0, fmt.Errorf("get distance error %w", err)
	}
	if count == 0 {
		return 0, fmt.Errorf("obu %d: %w", id, ErrNotFound)
	}

	return total, nil
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}
