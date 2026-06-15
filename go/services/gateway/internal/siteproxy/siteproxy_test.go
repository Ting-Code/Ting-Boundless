package siteproxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsAPIPath(t *testing.T) {
	if !IsAPIPath("/v1/business/ping") {
		t.Fatal("expected api path")
	}
	if !IsAPIPath("/internal/webhooks/logto") {
		t.Fatal("expected internal path")
	}
	if IsAPIPath("/") || IsAPIPath("/product/1") {
		t.Fatal("expected non-api path")
	}
}

func TestComposeAPIAndSite_RoutesAPI(t *testing.T) {
	apiCalled := false
	siteCalled := false
	api := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { apiCalled = true })
	site := &Handler{upstream: http.HandlerFunc(func(http.ResponseWriter, *http.Request) { siteCalled = true })}

	h := ComposeAPIAndSite(api, site)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/users/me", nil))

	if !apiCalled || siteCalled {
		t.Fatalf("api=%v site=%v", apiCalled, siteCalled)
	}
}

func TestComposeAPIAndSite_RoutesSite(t *testing.T) {
	apiCalled := false
	siteCalled := false
	api := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { apiCalled = true })
	site := &Handler{upstream: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		siteCalled = true
		w.WriteHeader(http.StatusOK)
	})}

	h := ComposeAPIAndSite(api, site)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	if apiCalled || !siteCalled {
		t.Fatalf("api=%v site=%v", apiCalled, siteCalled)
	}
}