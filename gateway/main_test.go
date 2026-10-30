package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/mdoroguncev/toll-calculator/aggregator/client"
	"github.com/mdoroguncev/toll-calculator/types"
)

type fakeClient struct {
	inv    *types.Invoice
	err    error
	gotID  int
	called bool
}

var _ client.Client = (*fakeClient)(nil)

func (f *fakeClient) GetInvoice(_ context.Context, id int) (*types.Invoice, error) {
	f.gotID = id
	f.called = true

	return f.inv, f.err
}

func (f *fakeClient) Aggregate(_ context.Context, req *types.AggregateRequest) error {
	return nil
}

func TestHandleGetInvoice(t *testing.T) {
	tests := []struct {
		name         string
		method, path string
		fake         *fakeClient
		wantCode     int
		wantErr      string
		wantInv      *types.Invoice
		wantCalled   bool
	}{
		{
			name:   "success",
			method: http.MethodGet,
			path:   "/invoice/7",
			fake: &fakeClient{
				inv: &types.Invoice{OBUID: 7, TotalDistance: 100, TotalAmount: 315},
			},
			wantCode:   200,
			wantErr:    "",
			wantInv:    &types.Invoice{OBUID: 7, TotalDistance: 100, TotalAmount: 315},
			wantCalled: true,
		}, {
			name:       "invalid id",
			method:     http.MethodGet,
			path:       "/invoice/abc",
			fake:       &fakeClient{},
			wantCode:   400,
			wantErr:    `invalid obu id "abc"`,
			wantInv:    nil,
			wantCalled: false,
		}, {
			name:   "aggregator error",
			method: http.MethodGet,
			path:   "/invoice/7",
			fake: &fakeClient{
				err: errors.New("dial tcp: connection refused"),
			},
			wantCode:   500,
			wantErr:    "internal server error",
			wantInv:    nil,
			wantCalled: true,
		}, {
			name:       "wrong method",
			method:     http.MethodPost,
			path:       "/invoice/7",
			fake:       &fakeClient{},
			wantErr:    "",
			wantCode:   405,
			wantInv:    nil,
			wantCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)

			newRouter(tt.fake).ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Fatalf("want code %v, got %v", tt.wantCode, rec.Code)
			}
			if tt.fake.called != tt.wantCalled {
				t.Fatalf("want client called=%v, got %v", tt.wantCalled, tt.fake.called)
			}
			if tt.wantCalled && tt.fake.gotID != 7 {
				t.Errorf("want client got id=7, got %d", tt.fake.gotID)
			}
			if rec.Code == http.StatusMethodNotAllowed {
				return
			}

			if tt.wantInv != nil {
				var gotInv types.Invoice

				if err := json.NewDecoder(rec.Body).Decode(&gotInv); err != nil {
					t.Fatalf("failed to decode body: %v", err)
				}
				if !reflect.DeepEqual(gotInv, *tt.wantInv) {
					t.Errorf("want invoice %+v, got %+v", *tt.wantInv, gotInv)
				}
			} else {
				var resp map[string]string

				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode error body: %v", err)
				}
				if resp["error"] != tt.wantErr {
					t.Errorf("want error %q, got %q", tt.wantErr, resp["error"])
				}
			}
		})
	}
}
