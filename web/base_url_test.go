//nolint:testpackage // verifies the configured server handler and existing test repository.
package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHostedBaseURL(t *testing.T) {
	t.Setenv("GOOGLE_MAPS_BASE_URL", "https://tools.complexelite.com/google-maps-scraper/")
	jobs := paginatedJobs(1)
	jobs[0].Status = StatusOK
	srv := newTestServerWithRepo(t, &mockJobRepo{jobs: jobs})
	prefix := "/google-maps-scraper"
	cases := []struct {
		path string
		want []string
	}{
		{prefix + "/", []string{`href="` + prefix + `/api/docs"`, `href="` + prefix + `/static/css/main.css"`, `hx-post="` + prefix + `/scrape"`, `hx-get="` + prefix + `/jobs"`}},
		{prefix + "/jobs?page=1", []string{prefix + "/download?id=", prefix + "/delete?id=", prefix + "/view?id=", prefix + "/jobs?page="}},
		{prefix + "/api/docs", []string{`spec-url="` + prefix + `/static/spec/spec.yaml"`, `type="module"`}},
		{prefix + "/static/spec/spec.yaml", []string{`url: "https://tools.complexelite.com/google-maps-scraper"`, `https://tools.complexelite.com/google-maps-scraper/api/v1/jobs`}},
		{prefix + "/static/css/main.css", []string{"{"}},
		{prefix + "/api/v1/jobs", []string{"job-1"}},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			srv.srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, http.NoBody))
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d: %s", rec.Code, rec.Body.String())
			}
			for _, want := range tc.want {
				if !strings.Contains(rec.Body.String(), want) {
					t.Errorf("response missing %q", want)
				}
			}
		})
	}
	rec := httptest.NewRecorder()
	srv.srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, prefix+"?page=2", http.NoBody))
	if rec.Code != http.StatusPermanentRedirect || rec.Header().Get("Location") != prefix+"/?page=2" {
		t.Fatalf("canonical redirect lost prefix or query: %d %s", rec.Code, rec.Header().Get("Location"))
	}
}

func TestInvalidBaseURL(t *testing.T) {
	for _, value := range []string{"/relative", "ftp://tools.complexelite.com/app", "https://tools.complexelite.com/app?query=1"} {
		t.Setenv("GOOGLE_MAPS_BASE_URL", value)
		if _, err := New(nil, ":0"); err == nil {
			t.Errorf("accepted invalid base URL %q", value)
		}
	}
}
