package aggregator

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/mdoroguncev/toll-calculator/internal/types"
)

func TestPostgresStore(t *testing.T) {
	got := os.Getenv("TEST_DB_DSN")
	if got == "" {
		t.Skip("TEST_DB_DSN not set")
	}

	store, err := NewPostgresStore(t.Context(), got)
	if err != nil {
		t.Fatalf("failed to create postgres store %v", err)
	}
	t.Cleanup(store.Close)

	testOBUIDs := []int{101, 102, 103, 104, 105}

	cleanup := func(t *testing.T) {
		t.Helper()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := store.pool.Exec(ctx,
			`DELETE FROM distance WHERE obu_id = ANY($1)`, testOBUIDs); err != nil {
			t.Fatalf("failed to clean test rows: %v", err)
		}
	}
	cleanup(t)
	t.Cleanup(func() { cleanup(t) })

	tests := []struct {
		name    string
		inserts []types.Distance
		obuID   int
		wantKm  float64
		wantErr error
	}{
		{
			name: "sum distances for one OBU",
			inserts: []types.Distance{
				{OBUID: 101, Value: 1.5, Unix: 1700000000},
				{OBUID: 101, Value: 2.5, Unix: 1700000000},
				{OBUID: 102, Value: 9, Unix: 1700000000},
			},
			obuID:   101,
			wantKm:  4.0,
			wantErr: nil,
		},
		{
			name: "zero distance for known OBU",
			inserts: []types.Distance{
				{OBUID: 103, Value: 0, Unix: 1700000000},
			},
			obuID:   103,
			wantKm:  0,
			wantErr: nil,
		},
		{
			name:    "unknown OBU",
			obuID:   104,
			wantErr: ErrNotFound,
		},
	}

	for _, tt := range tests {
		ctx := t.Context()
		t.Run(tt.name, func(t *testing.T) {
			for _, dist := range tt.inserts {
				if err := store.Insert(ctx, dist); err != nil {
					t.Fatalf("failed to insert distance %v", err)
				}
			}

			gotKm, err := store.Get(ctx, tt.obuID)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("want error %v, got %v", tt.wantErr, err)
			}

			if tt.wantErr == nil && gotKm != tt.wantKm {
				t.Errorf("want km %v, got %v", tt.wantKm, gotKm)
			}
		})
	}

	t.Run("store Unix timestamp", func(t *testing.T) {
		ctx := t.Context()

		dist := types.Distance{
			OBUID: 105,
			Value: 1.5,
			Unix:  1700000000,
		}

		if err := store.Insert(ctx, dist); err != nil {
			t.Fatalf("failed to insert distance %v", err)
		}

		var gotUnix int64
		query := "SELECT unix FROM distance WHERE obu_id = $1"
		if err := store.pool.QueryRow(ctx, query, dist.OBUID).Scan(&gotUnix); err != nil {
			t.Fatalf("failed to read unix timestamp: %v", err)
		}

		if gotUnix != int64(dist.Unix) {
			t.Errorf("want unix %d, got %d", dist.Unix, gotUnix)
		}
	})
}
