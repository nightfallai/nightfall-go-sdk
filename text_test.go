package nightfall

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestScanText(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		wantErr bool
	}{
		{
			name: "happy path",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantErr: false,
		},
		{
			name: "transient error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
	}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer s.Close()

	client, err := NewClient(OptionAPIKey("some key"))
	if err != nil {
		t.Fatal("Error initializing client")
	}
	client.baseURL = s.URL + "/"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s.Config.Handler = test.handler
			_, err = client.ScanText(context.Background(), &ScanTextRequest{
				Payload: []string{"4242 4242 4242 4242"},
				Policy: &Config{
					DetectionRules: []DetectionRule{{
						Detectors: []Detector{{
							MinNumFindings:    1,
							MinConfidence:     ConfidencePossible,
							DisplayName:       "cc#",
							DetectorType:      DetectorTypeNightfallDetector,
							NightfallDetector: "CREDIT_CARD_NUMBER",
						}},
						LogicalOp: LogicalOpAny,
					}},
					DefaultRedactionConfig: &RedactionConfig{
						MaskConfig: &MaskConfig{
							MaskingChar:   "🤫",
							CharsToIgnore: []string{"-"},
						},
						RemoveFinding: true,
					},
				},
			})
			if !test.wantErr && err != nil {
				t.Errorf("Got unexpected error: %v", err)
			}
			if test.wantErr && err == nil {
				t.Error("Did not get expected error")
			}
		})
	}
}
