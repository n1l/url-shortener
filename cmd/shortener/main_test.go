package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/n1l/url-shortener/internal/config"
	"github.com/n1l/url-shortener/internal/di"
	"github.com/n1l/url-shortener/internal/hasher"
	"github.com/n1l/url-shortener/internal/models"
	"github.com/n1l/url-shortener/internal/storage"
	"github.com/n1l/url-shortener/internal/zipper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetURLByHashHandler(t *testing.T) {
	options := config.Options{
		PublicHost: "http://example.com",
	}

	storage := storage.NewInMemoryStorage()
	services := di.NewServices(&options, storage, storage)

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
			hashID := hasher.GetHashOfURL(tc.expectedURL)
			if tc.expectedURL != "" {
				storage.Save(&models.URLRecord{
					ShortURL:    hashID,
					OriginalURL: tc.expectedURL,
				})
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/{id}", nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", hashID)

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			services.GetURLByHashHandler(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, fmt.Sprintf("Код ответа не совпадает с ожидаемым - %d", w.Code))

			if tc.expectedURL != "" {
				url := w.Header().Get("Location")
				assert.Equal(t, tc.expectedURL, url, fmt.Sprintf("URL в 'location' не равен ожидаемому - %s", tc.expectedURL))
			}
		})
	}
}

func TestGetURLByHashHandlerStatusCodes(t *testing.T) {
	options := config.Options{
		PublicHost: "http://example.com",
	}

	storage := storage.NewInMemoryStorage()
	services := di.NewServices(&options, storage, storage)

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

			services.GetURLByHashHandler(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestCreateShortedUrl(t *testing.T) {
	options := config.Options{
		PublicHost: "http://example.com",
	}

	storage := storage.NewInMemoryStorage()
	services := di.NewServices(&options, storage, storage)

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

			services.CreateShortedURLHandler(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if tc.expectedBody != "" {
				assert.Equal(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым"+" "+w.Body.String())
			}
		})
	}
}

func TestCreateShortedUrlJSON(t *testing.T) {
	options := config.Options{
		PublicHost: "http://example.com",
	}

	storage := storage.NewInMemoryStorage()
	services := di.NewServices(&options, storage, storage)

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
			body:         `{ "url" : "http://google.com" }`,
			expectedCode: http.StatusCreated,
			expectedBody: `{ "result" : "http://example.com/x7kg9X5V" }`,
		},
		{
			method:       http.MethodPost,
			body:         `{ "url" : "http://eynt73dlmnjj3b.biz/t0pwb" }`,
			expectedCode: http.StatusCreated,
			expectedBody: `{ "result" : "http://example.com/HppQetTZ" }`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			body := strings.NewReader(tc.body)
			r := httptest.NewRequest(tc.method, "/api/shorten", body)
			w := httptest.NewRecorder()

			services.CreateShortedURLfromJSONHandler(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if tc.expectedBody != "" {
				bodyStr := w.Body.String()
				assert.JSONEq(t, tc.expectedBody, bodyStr, "Тело ответа не совпадает с ожидаемым"+" "+bodyStr)
			}
		})
	}
}

func TestGzipCompression(t *testing.T) {
	options := config.Options{
		PublicHost: "http://example.com",
	}

	storage := storage.NewInMemoryStorage()
	services := di.NewServices(&options, storage, storage)

	handler := http.HandlerFunc(services.CreateShortedURLfromJSONHandler)

	srv := httptest.NewServer(zipper.GzipMiddleware(handler))
	defer srv.Close()

	requestBody := `{ "url" : "http://google.com" }`

	successBody := `{ "result" : "http://example.com/x7kg9X5V" }`

	t.Run("sends_gzip", func(t *testing.T) {
		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err := zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		buf := bytes.NewBufferString(requestBody)
		r := httptest.NewRequest("POST", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.JSONEq(t, successBody, string(b))
	})
}
