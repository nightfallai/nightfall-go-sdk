# Nightfall Go SDK #

nightfall-go-sdk is a Go client library for accessing the Nightfall API. 
It allows you to add functionality to your applications to
scan plain text and files in order to detect different categories of information. You can leverage any of
the detectors in Nightfall's pre-built library, or you may programmatically define your own custom detectors. 

Additionally, this library provides convenient features including the steps to chunk and upload files.

To obtain an API Key, login to the [Nightfall dashboard](https://app.nightfall.ai/) and click the section
titled "Manage API Keys".

See our [developer documentation](https://docs.nightfall.ai/docs/entities-and-terms-to-know) for more details about
integrating with the Nightfall API.

## Installation ##

Nightfall Go SDK is compatible with modern Go releases in module mode, with Go installed:

```bash
go get github.com/nightfallai/nightfall-go-sdk
```

will resolve and add the package to the current development module, along with its dependencies.

## Usage

### Scanning Plain Text

Nightfall provides pre-built detector types, covering data types ranging from PII to PHI to credentials. The following
snippet shows an example of how to scan using pre-built detectors.

####  Sample Code
```go
nc, err := nightfall.NewClient()
if err != nil {
    log.Printf("Error initializing client: %v", err)
    return
}

resp, err := nc.ScanText(context.Background(), &nightfall.ScanTextRequest{
    Payload: []string{"4242 4242 4242 4242 is my ccn"},
    Config:  &nightfall.Config{
        // A rule contains a set of detectors to scan with
        DetectionRules:     []nightfall.DetectionRule{{
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
    log.Printf("Error scanning text: %v", err)
    return
}
```

### Scanning Files

Scanning common file types like PDFs or office documents typically requires cumbersome text
extraction methods like OCR.

Rather than implementing this functionality yourself, the Nightfall API allows you to upload the
original files, and then we'll handle the heavy lifting.

The file upload process is implemented as a series of requests to upload the file in chunks. The library
provides a single method that wraps the steps required to upload your file. Please refer to the
[API Reference](https://docs.nightfall.ai/reference) for more details.

The file is uploaded synchronously, but as files can be arbitrarily large, the scan itself is conducted asynchronously.
The results from the scan are delivered by webhook; for more information about setting up a webhook server, refer to
[the docs](https://docs.nightfall.ai/docs/creating-a-webhook-server).

#### Sample Code

```go
nc, err := nightfall.NewClient()
if err != nil {
    log.Printf("Error initializing client: %v", err)
    return
}

f, err := os.Open("./ccn.txt")
if err != nil {
    log.Printf("Error opening file: %v", err)
    return
}
defer f.Close()

fi, err := f.Stat()
if err != nil {
    log.Printf("Error getting file info: %v", err)
    return
}

resp, err := nc.ScanFile(context.Background(), &nightfall.ScanFileRequest{
    Policy: &nightfall.ScanPolicy{
        // File scans are conducted asynchronously, so provide a webhook route to an HTTPS server to send results to.
        WebhookURL: "https://my-service.com/nightfall/listener",
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
    RequestMetadata: "{\"hello\": \"world\", \"goodnight\": \"moon\"}",
    Content:          f,
    ContentSizeBytes: fi.Size(),
    Timeout:          0,
})
if err != nil {
    log.Printf("Error scanning file: %v", err)
    return
}
```
