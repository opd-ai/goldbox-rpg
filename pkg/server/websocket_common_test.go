package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSessionFromContext(t *testing.T) {
	s := &RPCServer{}

	t.Run("no session in context", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/ws", nil)

		session, ok := s.getSessionFromContext(w, r, "TestHandler")

		assert.False(t, ok)
		assert.Nil(t, session)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("nil session value", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/ws", nil)
		ctx := context.WithValue(r.Context(), sessionKey, nil)
		r = r.WithContext(ctx)

		session, ok := s.getSessionFromContext(w, r, "TestHandler")

		assert.False(t, ok)
		assert.Nil(t, session)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("wrong type in context", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/ws", nil)
		ctx := context.WithValue(r.Context(), sessionKey, "not a session")
		r = r.WithContext(ctx)

		session, ok := s.getSessionFromContext(w, r, "TestHandler")

		assert.False(t, ok)
		assert.Nil(t, session)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid session", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/ws", nil)
		expectedSession := &PlayerSession{SessionID: "test-session"}
		ctx := context.WithValue(r.Context(), sessionKey, expectedSession)
		r = r.WithContext(ctx)

		session, ok := s.getSessionFromContext(w, r, "TestHandler")

		assert.True(t, ok)
		assert.Equal(t, expectedSession, session)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
