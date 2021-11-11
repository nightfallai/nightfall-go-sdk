package nightfall

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

const DefaultThreshold = 5 * time.Minute

// WebhookValidator validates incoming requests from Nightfall using the signing secret which can be fetched from the
// Nightfall dashboard
type WebhookValidator struct {
	signingSecret []byte
	threshold     time.Duration
}

// WebhookValidatorOption defines an option for a WebhookValidator
type WebhookValidatorOption func(*WebhookValidator)

// NewWebhookValidator returns a new webhook validator
func NewWebhookValidator(signingSecret []byte, options ...WebhookValidatorOption) *WebhookValidator {
	w := &WebhookValidator{
		signingSecret: signingSecret,
		threshold:     DefaultThreshold,
	}

	for _, opt := range options {
		opt(w)
	}

	return w
}

// OptionThreshold sets the threshold of the webhook validator. If the difference between the time the webhook is
// received and the timestamp value sent in the X-Nightfall-Timestamp header is greater than this threshold, the
// request will be rejected.
func OptionThreshold(threshold time.Duration) func(*WebhookValidator) {
	return func(w *WebhookValidator) {
		w.threshold = threshold
	}
}

// Validates that the provided request payload is an authentic request that originated from Nightfall. If this
// method returns false, request handlers shall not process the provided body any further.
func (w *WebhookValidator) Validate(requestBody, requestSignature, requestTime string) (bool, error) {
	if requestBody == "" || requestSignature == "" || requestTime == "" {
		return false, nil
	}

	i, err := strconv.ParseInt(requestTime, 10, 64)
	if err != nil {
		return false, err
	}

	unixTime := time.Unix(i, 0)
	if time.Now().Sub(unixTime) > w.threshold {
		return false, nil
	}

	h := hmac.New(sha256.New, w.signingSecret)
	hashPayload := fmt.Sprintf("%s:%s", requestTime, requestBody)
	h.Write([]byte(hashPayload))
	hexHash := hex.EncodeToString(h.Sum(nil))

	return hexHash == requestSignature, nil
}
