package nightfall

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"time"
)

const (
	APIURL = "https://api.nightfall.ai/"

	DefaultFileUploadConcurrency = 1
	DefaultRetryCount            = 5
)

// Client manages communication with the Nightfall API
type Client struct {
	baseURL               string
	apiKey                string
	httpClient            *http.Client
	fileUploadConcurrency int
	retryCount            int
}

// Set the base URL to a different value. Needed to use the client with
// Nightfall's dev or staging environments.
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// ClientOption defines an option for a Client
type ClientOption func(*Client) error

var (
	errMissingAPIKey                = errors.New("missing api key")
	errInvalidFileUploadConcurrency = errors.New("fileUploadConcurrency must be in range [1,100]")
	errRetryable429                 = errors.New("429 retryable error")

	userAgent = loadUserAgent()
)

// NewClient configures, validates, then creates an instance of a Nightfall Client.
func NewClient(options ...ClientOption) (*Client, error) {
	c := &Client{
		baseURL:               APIURL,
		apiKey:                os.Getenv("NIGHTFALL_API_KEY"),
		httpClient:            &http.Client{},
		fileUploadConcurrency: DefaultFileUploadConcurrency,
		retryCount:            DefaultRetryCount,
	}

	for _, opt := range options {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	if c.apiKey == "" {
		return nil, errMissingAPIKey
	}

	return c, nil
}

// OptionAPIKey sets the api key used in the Nightfall client
func OptionAPIKey(apiKey string) func(*Client) error {
	return func(c *Client) error {
		c.apiKey = apiKey
		return nil
	}
}

// OptionHTTPClient sets the http client used in the Nightfall client
func OptionHTTPClient(client *http.Client) func(*Client) error {
	return func(c *Client) error {
		c.httpClient = client
		return nil
	}
}

// OptionFileUploadConcurrency sets the number of goroutines that will upload chunks of data when scanning files with the Nightfall client
func OptionFileUploadConcurrency(fileUploadConcurrency int) func(*Client) error {
	return func(c *Client) error {
		if fileUploadConcurrency > 100 || fileUploadConcurrency <= 0 {
			return errInvalidFileUploadConcurrency
		}
		c.fileUploadConcurrency = fileUploadConcurrency
		return nil
	}
}

func loadUserAgent() string {
	prefix := "nightfall-go-sdk"

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return prefix
	}
	for _, dep := range buildInfo.Deps {
		if dep.Path == "github.com/nightfallai/nightfall-go-sdk" {
			return fmt.Sprintf("%s/%s", prefix, dep.Version)
		}
	}

	return prefix
}

type requestParams struct {
	method  string
	url     string
	body    []byte
	headers map[string]string
}

func (c *Client) defaultHeaders() map[string]string {
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + c.apiKey,
		"User-Agent":    userAgent,
	}
	return headers
}

func (c *Client) chunkedUploadHeaders(o int64) map[string]string {
	headers := map[string]string{
		"X-Upload-Offset": strconv.FormatInt(o, 10),
		"Content-Type":    "application/octet-stream",
		"Authorization":   "Bearer " + c.apiKey,
		"User-Agent":      userAgent,
	}
	return headers
}

func encodeBodyAsJSON(body interface{}) ([]byte, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		// Marshal() does not encode some special characters like "&" properly so we need to do this
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}
	b, err := io.ReadAll(buf)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (c *Client) do(ctx context.Context, reqParams requestParams, retResp interface{}) error {
	for attempt := 1; attempt <= c.retryCount+1; attempt++ {
		req, err := http.NewRequestWithContext(ctx, reqParams.method, reqParams.url, bytes.NewReader(reqParams.body))
		if err != nil {
			return err
		}
		for k, v := range reqParams.headers {
			req.Header.Set(k, v)
		}
		err = func() error {
			resp, err := c.httpClient.Do(req)
			if err != nil {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					return err
				}
			}
			defer resp.Body.Close()

			err = checkResponse(resp)
			if err != nil {
				if resp.StatusCode == http.StatusTooManyRequests {
					if attempt >= c.retryCount+1 {
						// We've hit the retry count limit, so just return the error
						return err
					}
					return errRetryable429
				}
				return err
			}

			// Request was successful so read response if any then return
			if retResp != nil {
				err = json.NewDecoder(resp.Body).Decode(retResp)
				if errors.Is(err, io.EOF) {
					err = nil
				}
			}

			return err
		}()
		if err == nil {
			break
		} else if errors.Is(err, errRetryable429) {
			// Sleep for 1s then retry on 429's
			time.Sleep(time.Second)
			continue
		} else {
			return err
		}
	}

	return nil
}

// Error is the struct returned by Nightfall API requests that are unsuccessful. This struct is generally returned
// when the HTTP status code is outside the range 200-299.
type Error struct {
	Code           int               `json:"code"`
	Message        string            `json:"message"`
	Description    string            `json:"description"`
	AdditionalData map[string]string `json:"additionalData"`
}

func (e *Error) Error() string {
	return e.Message
}

func checkResponse(r *http.Response) error {
	if 200 <= r.StatusCode && r.StatusCode <= 299 {
		return nil
	}

	e := &Error{}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil || len(b) == 0 {
		e.Code = r.StatusCode
		return e
	}

	err = json.Unmarshal(b, e)
	if err != nil {
		e.Code = r.StatusCode
		return e
	}

	return e
}
