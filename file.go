package nightfall

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ScanPolicy contains configuration that describes how to scan a file. Since the file is scanned asynchronously,
// the results from the scan are delivered to the provided webhook URL. The scan configuration may contain both
// inline detection rule definitions and UUID's referring to existing detection rules (up to 10 of each).
type ScanPolicy struct {
	WebhookURL         string          `json:"webhookURL"` // Deprecated: use AlertConfig instead
	DetectionRules     []DetectionRule `json:"detectionRules"`
	DetectionRuleUUIDs []string        `json:"detectionRuleUUIDs"`
	AlertConfig        *AlertConfig    `json:"alertConfig"`
}

// ScanFileRequest represents a request to scan a file that was uploaded via the Nightfall API. Exactly one of
// PolicyUUID or Policy should be provided.
type ScanFileRequest struct {
	PolicyUUID       *string       `json:"policyUUID"`
	Policy           *ScanPolicy   `json:"policy"`
	RequestMetadata  string        `json:"requestMetadata"`
	Content          io.Reader     `json:"-"`
	ContentSizeBytes int64         `json:"-"`
	Timeout          time.Duration `json:"-"`
}

// ScanFileResponse is the object returned by the Nightfall API when an (asynchronous) file scan request
// was successfully triggered.
type ScanFileResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type fileUploadResponse struct {
	ID            uuid.UUID `json:"id"`
	FileSizeBytes int64     `json:"fileSizeBytes"`
	ChunkSize     int64     `json:"chunkSize"`
	MIMEType      string    `json:"mimeType"`
}

type fileUploadRequest struct {
	FileSizeBytes int64 `json:"fileSizeBytes"`
}

// ScanFile is a convenience method that abstracts the details of the multi-step file upload and scan process.
// Calling this method for a given file is equivalent to (1) manually initializing a file upload session,
// (2) uploading all chunks of the file, (3) completing the upload, and (4) triggering a scan of the file.
//
// The maximum allowed ContentSizeBytes is dependent on the terms of your current
// Nightfall usage plan agreement; check the Nightfall dashboard for more details.
//
// This method consumes the provided reader, but it does not close it; closing remains
// the caller's responsibility.
func (c *Client) ScanFile(ctx context.Context, request *ScanFileRequest) (*ScanFileResponse, error) {
	var cancel context.CancelFunc
	if request.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, request.Timeout)
	} else {
		ctx, cancel = context.WithCancel(ctx)
	}
	defer cancel()

	fileUpload, err := c.initFileUpload(ctx, &fileUploadRequest{FileSizeBytes: request.ContentSizeBytes})
	if err != nil {
		return nil, err
	}

	err = c.doChunkedUpload(ctx, fileUpload, request.Content)
	if err != nil {
		return nil, err
	}

	err = c.completeFileUpload(ctx, fileUpload.ID)
	if err != nil {
		return nil, err
	}

	return c.scanUploadedFile(ctx, request, fileUpload.ID)
}

func (c *Client) initFileUpload(ctx context.Context, request *fileUploadRequest) (*fileUploadResponse, error) {
	req, err := c.newRequest(http.MethodPost, c.baseURL+"v3/upload", request)
	if err != nil {
		return nil, err
	}

	uploadResponse := &fileUploadResponse{}
	err = c.do(ctx, req, uploadResponse)
	if err != nil {
		return nil, err
	}

	return uploadResponse, nil
}

func (c *Client) doChunkedUpload(ctx context.Context, fileUpload *fileUploadResponse, content io.Reader) error {
	errChan := make(chan error, 1)
	wg := &sync.WaitGroup{}
	concurrencyChan := make(chan struct{}, c.fileUploadConcurrency)

	uploadCtx, cancel := context.WithCancel(ctx)
	defer cancel()

upload:
	for offset := int64(0); offset < fileUpload.FileSizeBytes; offset += fileUpload.ChunkSize {
		// Check if we are at max upload concurrency limit and block if we are
		concurrencyChan <- struct{}{}

		// Check if there were any errors from uploading previous chunks, and break if there were
		select {
		case <-uploadCtx.Done():
			break upload
		default:
		}

		buf := make([]byte, fileUpload.ChunkSize)
		bytesRead, err := content.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if int64(bytesRead) < fileUpload.ChunkSize {
			buf = buf[:bytesRead]
		}

		wg.Add(1)
		go func(o int64, data []byte) {
			defer func() {
				wg.Done()
				<-concurrencyChan
			}()

			req, err := c.newUploadRequest(http.MethodPatch, c.baseURL+"v3/upload/"+fileUpload.ID.String(), bytes.NewBuffer(data))
			if err != nil {
				// If error channel is full already just discard this error, first error is most likely the most useful one anyways
				select {
				case errChan <- err:
				default:
				}
				cancel()
				return
			}
			req.Header.Set("X-Upload-Offset", strconv.FormatInt(o, 10))

			err = c.do(uploadCtx, req, nil)
			if err != nil {
				// If error channel is full already just discard this error, first error is most likely the most useful one anyways
				select {
				case errChan <- err:
				default:
				}
				cancel()
				return
			}
		}(offset, buf)
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return err
	}

	return nil
}

func (c *Client) completeFileUpload(ctx context.Context, fileUUID uuid.UUID) error {
	req, err := c.newRequest(http.MethodPost, c.baseURL+"v3/upload/"+fileUUID.String()+"/finish", nil)
	if err != nil {
		return err
	}

	return c.do(ctx, req, nil)
}

func (c *Client) scanUploadedFile(ctx context.Context, request *ScanFileRequest, fileUUID uuid.UUID) (*ScanFileResponse, error) {
	req, err := c.newRequest(http.MethodPost, c.baseURL+"v3/upload/"+fileUUID.String()+"/scan", request)
	if err != nil {
		return nil, err
	}

	scanResponse := &ScanFileResponse{}
	err = c.do(ctx, req, scanResponse)
	if err != nil {
		return nil, err
	}

	return scanResponse, nil
}
