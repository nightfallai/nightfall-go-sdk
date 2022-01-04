package nightfall

import (
	"context"
	"net/http"
)

// ScanTextRequest is the request struct to scan inline plaintext with the Nightfall API.
type ScanTextRequest struct {
	Payload []string `json:"payload"`
	Config  *Config  `json:"config"`
}

// ScanTextResponse is the response object returned by a text scan request. Each index i in the field `findings`
// corresponds one-to-one with the input request payload, so all findings stored in a given sub-list
// refer to matches that occurred in the ith index of the request payload.
type ScanTextResponse struct {
	Findings        [][]*Finding `json:"findings"`
	RedactedPayload []string     `json:"redactedPayload"`
}

// Config is the configuration object to use when scanning inline plaintext with the Nightfall API.
type Config struct {
	DetectionRules         []DetectionRule  `json:"detectionRules"`
	DetectionRuleUUIDs     []string         `json:"detectionRuleUUIDs"`
	ContextBytes           int              `json:"contextBytes"`
	DefaultRedactionConfig *RedactionConfig `json:"defaultRedactionConfig"`
}

// Finding represents an occurrence of a configured detector (i.e. finding) in the provided data.
type Finding struct {
	Finding                   string           `json:"finding"`
	RedactedFinding           string           `json:"redactedFinding"`
	BeforeContext             string           `json:"beforeContext,omitempty"`
	AfterContext              string           `json:"afterContext,omitempty"`
	Detector                  DetectorMetadata `json:"detector"`
	Confidence                string           `json:"confidence"`
	Location                  *Location        `json:"location"`
	RedactedLocation          *Location        `json:"redactedLocation"`
	MatchedDetectionRuleUUIDs []string         `json:"matchedDetectionRuleUUIDs"`
	MatchedDetectionRules     []string         `json:"matchedDetectionRules"`
}

// Location represents where a finding was discovered in content.
type Location struct {
	ByteRange      *Range `json:"byteRange"`
	CodepointRange *Range `json:"codepointRange"`
}

// Range contains references to the start and end of the eponymous range.
type Range struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

// ScanText scans the provided plaintext against the provided detectors, and returns all findings. The response
// object will contain a list of lists representing the findings. Each index i in the findings array will
// correspond one-to-one with the input request payload list, so all findings stored in a given sub-list refer to
// matches that occurred in the ith index of the request payload.
func (c *Client) ScanText(ctx context.Context, request *ScanTextRequest) (*ScanTextResponse, error) {
	req, err := c.newRequest(http.MethodPost, c.baseURL+"v3/scan", request)
	if err != nil {
		return nil, err
	}

	scanResponse := &ScanTextResponse{}
	err = c.do(ctx, req, scanResponse)
	if err != nil {
		return nil, err
	}

	return scanResponse, nil
}
