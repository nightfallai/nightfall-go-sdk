package main

import (
	"context"
	"fmt"
	"os"

	"github.com/nightfallai/nightfall-go-sdk"
)

func scanFile(webhookURL, filePath string) (*nightfall.ScanFileResponse, error) {
	nc, err := nightfall.NewClient()
	if err != nil {
		return nil, fmt.Errorf("Error initializing client: %w", err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening file: %w", err)
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("Error getting file info: %w", err)
	}

	resp, err := nc.ScanFile(context.Background(), &nightfall.ScanFileRequest{
		Policy: &nightfall.ScanPolicy{
			// File scans are conducted asynchronously, so provide a webhook route to an HTTPS server to send results to.
			WebhookURL: webhookURL,
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
		RequestMetadata:  "{\"hello\": \"world\", \"goodnight\": \"moon\"}",
		Content:          f,
		ContentSizeBytes: fi.Size(),
		Timeout:          0,
	})
	if err != nil {
		return nil, fmt.Errorf("Error scanning file: %w", err)
	}
	return resp, nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Usage: file_scanner <webhookURL> <filename>")
		os.Exit(-1)
	}
	webhookURL := os.Args[1]
	filename := os.Args[2]

	resp, err := scanFile(webhookURL, filename)

	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(-1)
	}

	fmt.Printf("Got response with ID %s, message %s\n", resp.ID, resp.Message)
}
