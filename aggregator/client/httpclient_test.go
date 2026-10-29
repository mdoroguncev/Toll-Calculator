package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mdoroguncev/toll-calculator/types"
)

func TestGetInvoice(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		obuID   int
		want    *types.Invoice
		wantErr bool
	}{
		{
			name: "success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("want method %s, got %s", http.MethodGet, r.Method)
				}
				if r.URL.Path != "/invoice" {
					t.Errorf("want path /invoice, got %s", r.URL.Path)
				}
				if got := r.URL.Query().Get("obu"); got != "12345" {
					t.Errorf("want obuID 12345, got %s", got)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(types.Invoice{OBUID: 12345, TotalDistance: 1000, TotalAmount: 10})
			},
			obuID:   12345,
			want:    &types.Invoice{OBUID: 12345, TotalDistance: 1000, TotalAmount: 10},
			wantErr: false,
		}, {
			name: "server returns 500",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			obuID:   123456,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			c := NewHTTPClient(srv.URL)
			defer srv.Close()

			got, err := c.GetInvoice(context.Background(), tt.obuID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("want err %v, got %v", tt.wantErr, err)
			}

			if tt.wantErr {
				if got != nil {
					t.Errorf("want nil invoice, got %v", got)
				}
				return
			}

			if *got != *tt.want {
				t.Errorf("want %v, got %v", tt.want, got)
			}
		})
	}
}

func TestAggregate(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		req     *types.AggregateRequest
		wantErr bool
	}{
		{
			name: "success",
			req:  &types.AggregateRequest{ObuID: 12345, Value: 1000.10, Unix: 10},
			handler: func(w http.ResponseWriter, r *http.Request) {
				want := types.AggregateRequest{ObuID: 12345, Value: 1000.10, Unix: 10}

				if r.Method != http.MethodPost {
					t.Errorf("want method POST, got %s", r.Method)
				}
				if r.URL.Path != "/aggregate" {
					t.Errorf("want path /aggregate, got %s", r.URL.Path)
				}

				var got types.AggregateRequest
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Errorf("failed to decode request body %v", err)
				}

				if got.ObuID != want.ObuID || want.Value != got.Value || want.Unix != got.Unix {
					t.Errorf("want id %v, got %v. want value %v, got %v. want unix %v, got %v.", want.ObuID, got.ObuID, want.Value, got.Value, want.Unix, got.Unix)
				}
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		}, {
			name: "server returns 500",
			req:  &types.AggregateRequest{ObuID: 1, Value: 1, Unix: 1},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(tt.handler)
			defer srv.Close()

			c := NewHTTPClient(srv.URL)
			err := c.Aggregate(context.Background(), tt.req)

			if tt.wantErr && err == nil {
				t.Fatalf("want error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error %v", err)
			}
		})
	}
}
