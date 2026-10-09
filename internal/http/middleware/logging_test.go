package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func newTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.Use(RequestLogger())
	return r
}

func TestRequestIDIsSetAndEchoedInHeader(t *testing.T) {
	var captured string

	r := newTestEngine()
	r.GET("/x", func(c *gin.Context) {
		captured = RequestIDFrom(c)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))

	if captured == "" || captured == "-" {
		t.Fatalf("RequestIDFrom = %q, want a real id", captured)
	}
	if got := w.Header().Get(RequestIDHeader); got != captured {
		t.Errorf("header %s = %q, want %q", RequestIDHeader, got, captured)
	}
}

func TestRequestIDRespectsIncomingHeader(t *testing.T) {
	var captured string

	r := newTestEngine()
	r.GET("/x", func(c *gin.Context) { captured = RequestIDFrom(c) })

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set(RequestIDHeader, "retry-id-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Un retry del mismo cliente tiene que conservar la correlación.
	if captured != "retry-id-123" {
		t.Errorf("RequestIDFrom = %q, want retry-id-123", captured)
	}
}

func TestUserIDFromFallsBackToAnonymous(t *testing.T) {
	r := gin.New()

	var got string
	r.GET("/x", func(c *gin.Context) { got = UserIDFrom(c) })
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))

	if got != "-" {
		t.Errorf("UserIDFrom = %q, want -", got)
	}

	r2 := gin.New()
	r2.GET("/x", func(c *gin.Context) {
		c.Set("user_id", "u9")
		got = UserIDFrom(c)
	})
	r2.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))

	if got != "u9" {
		t.Errorf("UserIDFrom = %q, want u9", got)
	}
}
