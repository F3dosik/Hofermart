package gzip

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWithCompression(t *testing.T) {
	tests := []struct {
		name            string
		acceptEncoding  string
		contentEncoding string
		body            []byte
		handler         http.Handler
		wantBody        string
		wantGzip        bool
	}{
		{
			name:           "compress response when client supports gzip",
			acceptEncoding: "gzip",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("hello gzip"))
			}),
			wantBody: "hello gzip",
			wantGzip: true,
		},
		{
			name: "no compression when client does not support gzip",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("plain response"))
			}),
			wantBody: "plain response",
			wantGzip: false,
		},
		{
			name:            "decompress gzip request body",
			contentEncoding: "gzip",
			body:            []byte("compressed request"),
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				b, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				w.Write(b)
			}),
			wantBody: "compressed request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop().Sugar()

			var reqBody io.Reader

			if tt.contentEncoding == "gzip" {
				var buf bytes.Buffer
				gw := gzip.NewWriter(&buf)
				_, err := gw.Write(tt.body)
				require.NoError(t, err)
				require.NoError(t, gw.Close())
				reqBody = &buf
			} else {
				reqBody = bytes.NewReader(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/test", reqBody)

			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}

			handler := WithCompression(logger)(tt.handler)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			respBody := rr.Body.Bytes()

			if tt.wantGzip {
				assert.Equal(t, "gzip", rr.Header().Get("Content-Encoding"))

				gr, err := gzip.NewReader(bytes.NewReader(respBody))
				require.NoError(t, err)

				decompressed, err := io.ReadAll(gr)
				require.NoError(t, err)

				assert.Equal(t, tt.wantBody, string(decompressed))
			} else {
				assert.Equal(t, tt.wantBody, rr.Body.String())
			}
		})
	}
}
