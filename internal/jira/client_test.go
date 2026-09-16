package jira

import (
	"net/http"
	"testing"
)

// TestNewClientHonorsHTTPProxyEnv verifies that the Jira client's HTTP
// transport is wired to respect the standard HTTP_PROXY/HTTPS_PROXY/NO_PROXY
// environment variables, so requests to Jira can be routed through a
// corporate outbound proxy.
func TestNewClientHonorsHTTPProxyEnv(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://proxy.example.com:8080")
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("NO_PROXY", "")

	c := NewClient("https://jira.example.com", "token", "PROJ")

	transport, ok := c.httpClient.Transport.(*http.Transport)
	if !ok || transport.Proxy == nil {
		t.Fatalf("expected httpClient.Transport to be a *http.Transport with a Proxy func, got %#v", c.httpClient.Transport)
	}

	req, err := http.NewRequest(http.MethodGet, "http://jira.example.com/rest/api/2/field", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	proxyURL, err := transport.Proxy(req)
	if err != nil {
		t.Fatalf("resolve proxy: %v", err)
	}

	want := "http://proxy.example.com:8080"
	if proxyURL == nil || proxyURL.String() != want {
		t.Fatalf("proxy URL = %v, want %v", proxyURL, want)
	}
}
