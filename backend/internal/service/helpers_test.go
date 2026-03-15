package service

import "net/http"

// hostRewriteTransport rewrites all outbound requests to target a specific host.
// This lets tests point service code at an httptest.Server without modifying production URLs.
type hostRewriteTransport struct {
	rt   http.RoundTripper
	host string
}

func (t *hostRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r2 := req.Clone(req.Context())
	r2.URL.Scheme = "http"
	r2.URL.Host = t.host
	return t.rt.RoundTrip(r2)
}
