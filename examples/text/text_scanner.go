package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nightfallai/nightfall-go-sdk"
)

func scanText() (*nightfall.ScanTextResponse, error) {
	nc, err := nightfall.NewClient()
	if err != nil {
		return nil, fmt.Errorf("Error initializing client: %w", err)
	}

	resp, err := nc.ScanText(context.Background(), &nightfall.ScanTextRequest{
		Payload: []string{"4242 4242 4242 4242 is my ccn"},
		Config: &nightfall.Config{
			// A rule contains a set of detectors to scan with
			DetectionRules: []nightfall.DetectionRule{{
				// Define some detectors to use to scan your data
				Detectors: []nightfall.Detector{{
					MinNumFindings:    1,
					MinConfidence:     nightfall.ConfidencePossible,
					DisplayName:       "cc#",
					DetectorType:      nightfall.DetectorTypeNightfallDetector,
					NightfallDetector: "CREDIT_CARD_NUMBER",
				}},
				LogicalOp: nightfall.LogicalOpAny,
			},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Error scanning text: %w", err)
	}
	return resp, nil
}

func main() {
	resp, err := scanText()
	if err != nil {
		fmt.Printf("Got error: %v", err)
		os.Exit(-1)
	}
	for _, findings := range resp.Findings {
		for _, finding := range findings {
			fmt.Printf("Got finding %v", finding)
		}
	}
}
