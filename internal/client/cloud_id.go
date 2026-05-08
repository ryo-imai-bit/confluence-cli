package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// HTTPDoer abstracts HTTP request execution for testability.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// tenantInfoResponse represents the response from the _edge/tenant_info endpoint.
type tenantInfoResponse struct {
	CloudID string `json:"cloudId"`
}

// IsAtlassianNetURL checks if the given URL is a traditional *.atlassian.net URL.
// Returns the subdomain (e.g., "mycompany") and true if it matches.
func IsAtlassianNetURL(rawURL string) (string, bool) {
	if rawURL == "" {
		return "", false
	}

	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "", false
	}

	host := parsed.Hostname()
	if !strings.HasSuffix(host, ".atlassian.net") {
		return "", false
	}

	// Extract subdomain: "mycompany.atlassian.net" -> "mycompany"
	subdomain := strings.TrimSuffix(host, ".atlassian.net")
	if subdomain == "" || strings.Contains(subdomain, ".") {
		return "", false
	}

	return subdomain, true
}

// ResolveCloudID fetches the Cloud ID for the given Atlassian site base URL.
// The baseURL should be like "https://mycompany.atlassian.net".
func ResolveCloudID(httpClient HTTPDoer, baseURL string) (string, error) {
	return fetchCloudIDFromURL(httpClient, baseURL+"/_edge/tenant_info")
}

// fetchCloudIDFromURL fetches the Cloud ID from the given full URL.
func fetchCloudIDFromURL(httpClient HTTPDoer, tenantInfoURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, tenantInfoURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create tenant info request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch tenant info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tenant info returned status %d: verify that the Atlassian domain is correct", resp.StatusCode)
	}

	var info tenantInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("failed to decode tenant info response: %w", err)
	}

	if info.CloudID == "" {
		return "", fmt.Errorf("tenant info response did not contain a cloudId")
	}

	return info.CloudID, nil
}

// cloudIDFetcher is a function type that fetches Cloud ID for a given domain.
type cloudIDFetcher func(domain string) (string, error)

// ResolveBaseURL resolves a base URL for use with the Confluence API.
// If the URL is a *.atlassian.net URL, it fetches the Cloud ID and returns
// the equivalent api.atlassian.com gateway URL.
// Non-atlassian.net URLs are returned unchanged without any HTTP calls.
func ResolveBaseURL(httpClient HTTPDoer, baseURL string) (string, error) {
	return resolveBaseURLWithTenantFetcher(baseURL, func(domain string) (string, error) {
		return ResolveCloudID(httpClient, "https://"+domain+".atlassian.net")
	})
}

// resolveBaseURLWithTenantFetcher is the internal implementation that accepts
// a fetcher function for testability.
func resolveBaseURLWithTenantFetcher(baseURL string, fetcher cloudIDFetcher) (string, error) {
	domain, ok := IsAtlassianNetURL(baseURL)
	if !ok {
		return baseURL, nil
	}

	cloudID, err := fetcher(domain)
	if err != nil {
		return "", fmt.Errorf("failed to resolve Cloud ID for %s.atlassian.net: %w", domain, err)
	}

	return "https://api.atlassian.com/ex/confluence/" + cloudID + "/wiki", nil
}
