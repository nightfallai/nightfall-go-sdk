package nightfall

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestScanFile(t *testing.T) {
	uuidStr := "430d42aa-1e1f-405d-8799-7f5f26486a0d"
	reqUUID := uuid.MustParse(uuidStr)
	var chunkCount int
	defaultInitUploadHandler := func(w http.ResponseWriter, r *http.Request) {
		resp := fileUploadResponse{
			ID:            reqUUID,
			FileSizeBytes: 15,
			ChunkSize:     5,
		}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}
	defaultUploadHandler := func(w http.ResponseWriter, r *http.Request) {
		chunkCount++
		w.WriteHeader(http.StatusOK)
	}
	defaultFinishHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	defaultScanHandler := func(w http.ResponseWriter, r *http.Request) {
		resp := ScanFileResponse{
			ID:      uuidStr,
			Message: "scan initiated",
		}
		b, _ := json.Marshal(resp)
		_, _ = w.Write(b)
	}

	tests := []struct {
		name          string
		handlers      map[string]http.HandlerFunc
		clientTimeOut time.Duration
		expChunks     int
		wantErr       bool
	}{
		{
			name: "happy path - 1 chunk",
			handlers: map[string]http.HandlerFunc{
				"/v3/upload": func(w http.ResponseWriter, r *http.Request) {
					resp := fileUploadResponse{
						ID:            reqUUID,
						FileSizeBytes: 15,
						ChunkSize:     15,
					}
					b, _ := json.Marshal(resp)
					_, _ = w.Write(b)
				},
				"/v3/upload/" + uuidStr:             defaultUploadHandler,
				"/v3/upload/" + uuidStr + "/finish": defaultFinishHandler,
				"/v3/upload/" + uuidStr + "/scan":   defaultScanHandler,
			},
			expChunks: 1,
			wantErr:   false,
		},
		{
			name: "happy path - 3 chunks",
			handlers: map[string]http.HandlerFunc{
				"/v3/upload":                        defaultInitUploadHandler,
				"/v3/upload/" + uuidStr:             defaultUploadHandler,
				"/v3/upload/" + uuidStr + "/finish": defaultFinishHandler,
				"/v3/upload/" + uuidStr + "/scan":   defaultScanHandler,
			},
			expChunks: 3,
			wantErr:   false,
		},
		{
			name: "upload timed out",
			handlers: map[string]http.HandlerFunc{
				"/v3/upload": defaultInitUploadHandler,
				"/v3/upload/" + uuidStr: func(w http.ResponseWriter, r *http.Request) {
					if chunkCount == 1 {
						time.Sleep(2 * time.Second)
						w.WriteHeader(http.StatusOK)
						return
					}
					chunkCount++
					w.WriteHeader(http.StatusOK)
				},
				"/v3/upload/" + uuidStr + "/finish": defaultFinishHandler,
				"/v3/upload/" + uuidStr + "/scan":   defaultScanHandler,
			},
			clientTimeOut: 1 * time.Second,
			expChunks:     1,
			wantErr:       true,
		},
		{
			name: "upload init failed",
			handlers: map[string]http.HandlerFunc{
				"/v3/upload": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				},
			},
			wantErr: true,
		},
		{
			name: "upload failed",
			handlers: map[string]http.HandlerFunc{
				"/v3/upload": defaultInitUploadHandler,
				"/v3/upload/" + uuid.UUID{}.String(): func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				},
			},
			wantErr: true,
		},
		{
			name: "upload finish failed",
			handlers: map[string]http.HandlerFunc{
				"/v3/upload":            defaultInitUploadHandler,
				"/v3/upload/" + uuidStr: defaultUploadHandler,
				"/v3/upload/" + uuidStr + "/finish": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				},
			},
			expChunks: 3,
			wantErr:   true,
		},
		{
			name: "upload init failed",
			handlers: map[string]http.HandlerFunc{
				"/v3/upload":                        defaultInitUploadHandler,
				"/v3/upload/" + uuidStr:             defaultUploadHandler,
				"/v3/upload/" + uuidStr + "/finish": defaultFinishHandler,
				"/v3/upload/" + uuidStr + "/scan": func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				},
			},
			expChunks: 3,
			wantErr:   true,
		},
	}

	for _, test := range tests {
		chunkCount = 0

		t.Run(test.name, func(t *testing.T) {
			mux := http.NewServeMux()
			for pattern, handler := range test.handlers {
				mux.HandleFunc(pattern, handler)
			}
			s := httptest.NewServer(mux)
			defer s.Close()

			client, err := NewClient(OptionAPIKey("some key"), OptionFileUploadConcurrency(2))
			if err != nil {
				t.Fatal("Error initializing client")
			}
			client.baseURL = s.URL + "/"

			_, err = client.ScanFile(context.Background(), &ScanFileRequest{
				Content:          strings.NewReader("4242 4242 4242 4242"),
				ContentSizeBytes: 15,
				Timeout:          test.clientTimeOut,
			})
			if !test.wantErr && err != nil {
				t.Errorf("Got unexpected error: %v", err)
			}
			if test.wantErr && err == nil {
				t.Error("Did not get expected error")
			}
			if chunkCount != test.expChunks {
				t.Error("Did not upload expected number of chunks")
			}
		})
	}
}
