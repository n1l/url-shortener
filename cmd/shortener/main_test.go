package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestGetURLByHash(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
		expectedURL  string
	}{
		{
			method:       http.MethodGet,
			expectedCode: http.StatusBadRequest,
			expectedURL:  "",
		},
		{
			method:       http.MethodGet,
			expectedCode: http.StatusTemporaryRedirect,
			expectedURL:  "http://google.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			hashID := getHashOfURL(tc.expectedURL)
			if tc.expectedURL != "" {
				shortedUrls[hashID] = tc.expectedURL
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/{id}", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", hashID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			GetURLByHashHandler(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, fmt.Sprintf("Код ответа не совпадает с ожидаемым - %d", w.Code))

			if tc.expectedURL != "" {
				url := w.Header().Get("Location")
				assert.Equal(t, tc.expectedURL, url, fmt.Sprintf("URL в 'location' не равен ожидаемому - %s", tc.expectedURL))
			}
		})
	}
}

func TestGetURLByHashStatusCodes(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
	}{
		{
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
		},
		{
			method:       http.MethodPut,
			expectedCode: http.StatusBadRequest,
		},
		{
			method:       http.MethodDelete,
			expectedCode: http.StatusBadRequest,
		},
		{
			method:       http.MethodHead,
			expectedCode: http.StatusBadRequest,
		},
		{
			method:       http.MethodPatch,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/someId", nil)
			w := httptest.NewRecorder()

			GetURLByHashHandler(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestCreateShortedUrl(t *testing.T) {
	options.PublicHost = "http://example.com"

	testCases := []struct {
		method       string
		body         string
		expectedCode int
		expectedBody string
	}{
		{
			method:       http.MethodGet,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			method:       http.MethodPut,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			method:       http.MethodDelete,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			method:       http.MethodHead,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			method:       http.MethodPatch,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
		},
		{
			method:       http.MethodPost,
			body:         "http://google.com",
			expectedCode: http.StatusCreated,
			expectedBody: "http://example.com/x7kg9X5V",
		},
		{
			method:       http.MethodPost,
			body:         "http://eynt73dlmnjj3b.biz/t0pwb",
			expectedCode: http.StatusCreated,
			expectedBody: "http://example.com/HppQetTZ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			body := strings.NewReader(tc.body)
			r := httptest.NewRequest(tc.method, "/", body)
			w := httptest.NewRecorder()

			CreateShortedURLHandler(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым"+" "+w.Body.String())
			}
		})
	}
}
