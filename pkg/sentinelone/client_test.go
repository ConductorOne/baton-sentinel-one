package sentinelone

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetAccounts_HTMLResponse401(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`<html><body><h1>401 Unauthorized</h1></body></html>`))
	}))
	defer server.Close()

	client := NewClient(http.DefaultClient, server.URL+"/", "bad-token")
	_, _, err := client.GetAccounts(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if strings.Contains(err.Error(), "invalid character '<'") {
		t.Errorf("got cryptic JSON parse error instead of descriptive HTTP error: %v", err)
	}

	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected HTTP status 401 in error message, got: %v", err)
	}
}

func TestGetAccounts_HTMLResponse500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`<html><body><h1>500 Internal Server Error</h1></body></html>`))
	}))
	defer server.Close()

	client := NewClient(http.DefaultClient, server.URL+"/", "test-token")
	_, _, err := client.GetAccounts(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if strings.Contains(err.Error(), "invalid character '<'") {
		t.Errorf("got cryptic JSON parse error instead of descriptive HTTP error: %v", err)
	}

	if !strings.Contains(err.Error(), "500") {
		t.Errorf("expected HTTP status 500 in error message, got: %v", err)
	}
}

func TestGetAccounts_ValidJSON(t *testing.T) {
	accounts := Response[Account]{
		Data: []Account{
			{ID: "acc-1", Name: "Test Account", AccountType: "Trial", IsDefault: true},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(accounts)
	}))
	defer server.Close()

	client := NewClient(http.DefaultClient, server.URL+"/", "valid-token")
	result, cursor, err := client.GetAccounts(context.Background(), nil)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 account, got %d", len(result))
	}

	if result[0].ID != "acc-1" {
		t.Errorf("expected account ID 'acc-1', got %q", result[0].ID)
	}

	if cursor != "" {
		t.Errorf("expected empty cursor, got %q", cursor)
	}
}
