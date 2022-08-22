package nightfall

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDo(t *testing.T) {
	var callCount int
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		expCalls int
		wantErr  bool
	}{
		{
			name: "happy path",
			handler: func(w http.ResponseWriter, r *http.Request) {
				callCount++
				w.WriteHeader(http.StatusOK)
			},
			expCalls: 1,
			wantErr:  false,
		},
		{
			name: "happy path - retry 2 times",
			handler: func(w http.ResponseWriter, r *http.Request) {
				callCount++
				if callCount == 3 {
					w.WriteHeader(http.StatusOK)
					return
				}
				w.WriteHeader(http.StatusTooManyRequests)
			},
			expCalls: 3,
			wantErr:  false,
		},
		{
			name: "429 error after 5 retries",
			handler: func(w http.ResponseWriter, r *http.Request) {
				callCount++
				w.WriteHeader(http.StatusTooManyRequests)
			},
			expCalls: 6,
			wantErr:  true,
		},
		{
			name: "transient error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				callCount++
				w.WriteHeader(http.StatusInternalServerError)
			},
			expCalls: 1,
			wantErr:  true,
		},
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer s.Close()

	client, err := NewClient(OptionAPIKey("some key"))
	if err != nil {
		t.Fatal("Error initializing client")
	}

	reqParams := requestParams{
		method: http.MethodPost,
		url:    s.URL,
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			callCount = 0
			s.Config.Handler = test.handler
			err = client.do(context.Background(), reqParams, nil)
			if !test.wantErr && err != nil {
				t.Errorf("Got unexpected error: %v", err)
			}
			if test.wantErr && err == nil {
				t.Error("Did not get expected error")
			}
			if callCount != test.expCalls {
				t.Error("Did not call expected number of times")
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name                  string
		apiKey                string
		fileUploadConcurrency int
		wantErr               bool
	}{
		{
			name:                  "happy path",
			apiKey:                "some key",
			fileUploadConcurrency: 5,
			wantErr:               false,
		},
		{
			name:                  "missing api key",
			fileUploadConcurrency: 5,
			wantErr:               true,
		},
		{
			name:                  "file concurrency too high",
			apiKey:                "some key",
			fileUploadConcurrency: 101,
			wantErr:               true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewClient(OptionAPIKey(test.apiKey), OptionFileUploadConcurrency(test.fileUploadConcurrency))
			if !test.wantErr && err != nil {
				t.Errorf("Got unexpected error: %v", err)
			}
			if test.wantErr && err == nil {
				t.Error("Did not get expected error")
			}
		})
	}
}
