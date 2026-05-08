package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsAtlassianNetURL(t *testing.T) {
	tests := []struct {
		name       string
		rawURL     string
		wantDomain string
		wantOK     bool
	}{
		{
			name:       "atlassian.net with wiki path",
			rawURL:     "https://mycompany.atlassian.net/wiki",
			wantDomain: "mycompany",
			wantOK:     true,
		},
		{
			name:       "atlassian.net without path",
			rawURL:     "https://mycompany.atlassian.net",
			wantDomain: "mycompany",
			wantOK:     true,
		},
		{
			name:       "atlassian.net with trailing slash",
			rawURL:     "https://mycompany.atlassian.net/wiki/",
			wantDomain: "mycompany",
			wantOK:     true,
		},
		{
			name:       "api.atlassian.com URL",
			rawURL:     "https://api.atlassian.com/ex/confluence/abc123/wiki",
			wantDomain: "",
			wantOK:     false,
		},
		{
			name:       "non-atlassian URL",
			rawURL:     "https://example.com/wiki",
			wantDomain: "",
			wantOK:     false,
		},
		{
			name:       "empty string",
			rawURL:     "",
			wantDomain: "",
			wantOK:     false,
		},
		{
			name:       "bare atlassian.net without subdomain",
			rawURL:     "https://atlassian.net",
			wantDomain: "",
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, ok := IsAtlassianNetURL(tt.rawURL)
			if ok != tt.wantOK {
				t.Errorf("IsAtlassianNetURL(%q) ok = %v, want %v", tt.rawURL, ok, tt.wantOK)
			}
			if domain != tt.wantDomain {
				t.Errorf("IsAtlassianNetURL(%q) domain = %q, want %q", tt.rawURL, domain, tt.wantDomain)
			}
		})
	}
}

func TestResolveCloudID(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		wantID     string
		wantErr    bool
	}{
		{
			name:       "successful response",
			statusCode: http.StatusOK,
			body:       `{"cloudId":"abc123"}`,
			wantID:     "abc123",
			wantErr:    false,
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			body:       `{"error":"not found"}`,
			wantID:     "",
			wantErr:    true,
		},
		{
			name:       "invalid JSON",
			statusCode: http.StatusOK,
			body:       `not json`,
			wantID:     "",
			wantErr:    true,
		},
		{
			name:       "empty cloudId",
			statusCode: http.StatusOK,
			body:       `{"cloudId":""}`,
			wantID:     "",
			wantErr:    true,
		},
		{
			name:       "missing cloudId field",
			statusCode: http.StatusOK,
			body:       `{"other":"value"}`,
			wantID:     "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/_edge/tenant_info" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			// Use server URL as the domain base (ResolveCloudID builds the full URL)
			id, err := ResolveCloudID(server.Client(), server.URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveCloudID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if id != tt.wantID {
				t.Errorf("ResolveCloudID() = %q, want %q", id, tt.wantID)
			}
		})
	}
}

func TestResolveBaseURL(t *testing.T) {
	t.Run("atlassian.net URL resolves to api.atlassian.com", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"cloudId":"abc123"}`))
		}))
		defer server.Close()

		// ResolveBaseURL needs to call tenant_info on the atlassian.net domain.
		// In tests, we override the tenant info URL via the provided httpClient + tenantInfoURL helper.
		// For this test, we use resolveBaseURLWithTenantFetcher to inject the test server.
		got, err := resolveBaseURLWithTenantFetcher("https://mycompany.atlassian.net/wiki", func(domain string) (string, error) {
			return fetchCloudIDFromURL(server.Client(), server.URL+"/_edge/tenant_info")
		})
		if err != nil {
			t.Fatalf("ResolveBaseURL() unexpected error: %v", err)
		}
		want := "https://api.atlassian.com/ex/confluence/abc123/wiki"
		if got != want {
			t.Errorf("ResolveBaseURL() = %q, want %q", got, want)
		}
	})

	t.Run("non-atlassian URL returns unchanged without HTTP call", func(t *testing.T) {
		called := false
		got, err := resolveBaseURLWithTenantFetcher("https://example.com/wiki", func(domain string) (string, error) {
			called = true
			return "", nil
		})
		if err != nil {
			t.Fatalf("ResolveBaseURL() unexpected error: %v", err)
		}
		if got != "https://example.com/wiki" {
			t.Errorf("ResolveBaseURL() = %q, want %q", got, "https://example.com/wiki")
		}
		if called {
			t.Error("ResolveBaseURL() should not call tenant fetcher for non-atlassian URL")
		}
	})

	t.Run("api.atlassian.com URL returns unchanged without HTTP call", func(t *testing.T) {
		called := false
		got, err := resolveBaseURLWithTenantFetcher("https://api.atlassian.com/ex/confluence/abc/wiki", func(domain string) (string, error) {
			called = true
			return "", nil
		})
		if err != nil {
			t.Fatalf("ResolveBaseURL() unexpected error: %v", err)
		}
		if got != "https://api.atlassian.com/ex/confluence/abc/wiki" {
			t.Errorf("ResolveBaseURL() = %q, want unchanged", got)
		}
		if called {
			t.Error("ResolveBaseURL() should not call tenant fetcher for api.atlassian.com URL")
		}
	})

	t.Run("tenant_info failure returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`not found`))
		}))
		defer server.Close()

		_, err := resolveBaseURLWithTenantFetcher("https://mycompany.atlassian.net/wiki", func(domain string) (string, error) {
			return fetchCloudIDFromURL(server.Client(), server.URL+"/_edge/tenant_info")
		})
		if err == nil {
			t.Fatal("ResolveBaseURL() expected error for failed tenant_info, got nil")
		}
	})
}
