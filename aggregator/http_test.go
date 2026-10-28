package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMakeHTTPHandlerFunc(t *testing.T) {
	tests := []struct {
		name     string
		fn       HTTPFunc
		wantCode int
		wantErr  string
	}{
		{
			name: "api error keeps its status code",
			fn: func(w http.ResponseWriter, r *http.Request) error {
				return APIError{Code: http.StatusBadRequest, Err: errors.New("bad request")}
			},
			wantCode: http.StatusBadRequest,
			wantErr:  "bad request",
		}, {
			name: "internal server error",
			fn: func(w http.ResponseWriter, r *http.Request) error {
				return errors.New("boom")
			},
			wantCode: http.StatusInternalServerError,
			wantErr:  "internal server error",
		}, {
			name: "no errors",
			fn: func(w http.ResponseWriter, r *http.Request) error {
				return writeJSON(w, http.StatusOK, map[string]string{"msg": "ok"})
			},
			wantCode: http.StatusOK,
			wantErr:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/invoice", nil)

			makeHTTPHandlerFunc(tt.fn)(rec, req)

			if tt.wantCode != rec.Code {
				t.Errorf("want %v code got %v code", tt.wantCode, rec.Code)
			}

			got := rec.Result().Header.Get("Content-Type")
			if got != "application/json" {
				t.Errorf("Content-Type = %q, want %q", got, "application/json")
			}

			var body map[string]any
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode body %v", err)
			}

			gotErr, _ := body["error"].(string)
			if gotErr != tt.wantErr {
				t.Errorf("want %s error, got %s", tt.wantErr, gotErr)
			}
		})
	}
}
