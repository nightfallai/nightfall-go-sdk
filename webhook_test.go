package nightfall

import (
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name         string
		threshold    time.Duration
		reqBody      string
		reqSignature string
		reqTimestamp string
		expValid     bool
	}{
		{
			name:         "happy path",
			threshold:    time.Hour * 24 * 365 * 100,
			reqBody:      "hello world",
			reqSignature: "3ccf9cc16507ca9f55b73c94fc2e872bb7ea312b2ab9d90785c6c55f153df4f3",
			reqTimestamp: "1633368643", // 2021-10-04T17:30:43Z
			expValid:     true,
		},
		{
			name:         "invalid signature",
			threshold:    time.Hour * 24 * 365 * 100,
			reqBody:      "hello world",
			reqSignature: "fe07c9a938ac1da7e1c14774bff295f27a05b3cb4e78275eeb873977322b63d1",
			reqTimestamp: "1633368643", // 2021-10-04T17:30:43Z
			expValid:     false,
		},
		{
			name:         "request time past threshold",
			threshold:    time.Minute,
			reqBody:      "hello world",
			reqSignature: "3ccf9cc16507ca9f55b73c94fc2e872bb7ea312b2ab9d90785c6c55f153df4f3",
			reqTimestamp: "1633368643", // 2021-10-04T17:30:43Z
			expValid:     false,
		},
	}

	for _, test := range tests {
		validator := NewWebhookValidator([]byte("some secret"), OptionThreshold(test.threshold))
		valid, err := validator.Validate(test.reqBody, test.reqSignature, test.reqTimestamp)
		if err != nil {
			t.Errorf("unexpected error validating request: %v", err)
		}
		if valid != test.expValid {
			t.Error("did not get expected validation result")
		}
	}
}
