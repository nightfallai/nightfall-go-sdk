package nightfall

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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

// ClientOption defines an option for a Client
type ClientOption func(*Client) error

var (
	errMissingAPIKey                = errors.New("missing api key")
	errInvalidFileUploadConcurrency = errors.New("fileUploadConcurrency must be in range [1,100]")
	errRetryable429                 = errors.New("429 retryable error")
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

// OptionHTTPClient sets the number of goroutines that will upload chunks of data when scanning files with the Nightfall client
func OptionFileUploadConcurrency(fileUploadConcurrency int) func(*Client) error {
	return func(c *Client) error {
		if fileUploadConcurrency > 100 || fileUploadConcurrency <= 0 {
			return errInvalidFileUploadConcurrency
		}
		c.fileUploadConcurrency = fileUploadConcurrency
		return nil
	}
}

func (c *Client) newRequest(method, urlStr string, body interface{}) (*http.Request, error) {
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

	req, err := http.NewRequest(method, urlStr, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	return req, nil
}

func (c *Client) newUploadRequest(method, urlStr string, reader io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, urlStr, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	return req, nil
}

func (c *Client) do(ctx context.Context, req *http.Request, retResp interface{}) error {
	req = req.WithContext(ctx)

	for attempt := 1; attempt <= c.retryCount+1; attempt++ {
		err := func() error {
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

// The error model returned by Nightfall API requests that are unsuccessful. This object is generally returned
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
