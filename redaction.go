package nightfall

// RedactionConfig describes how any detected findings should be redacted when returned to the client. When this
// configuration is provided as part of a request, exactly one of the four types of redaction should be non-nil:
//  1. Masking: replacing the characters of a finding with another character, such as '*' or '👀'
//  2. Info Type Substitution: replacing the finding with the name of the detector it matched, such
//                             as CREDIT_CARD_NUMBER
//  3. Substitution: replacing the finding with a custom string, such as "oh no!"
//  4. Encryption: encrypting the finding with an RSA public key
type RedactionConfig struct {
	MaskConfig                 *MaskConfig                 `json:"maskConfig"`
	InfoTypeSubstitutionConfig *InfoTypeSubstitutionConfig `json:"infoTypeSubstitutionConfig"`
	SubstitutionConfig         *SubstitutionConfig         `json:"substitutionConfig"`
	CryptoConfig               *CryptoConfig               `json:"cryptoConfig"`
	RemoveFinding              bool                        `json:"removeFinding"`
}

// MaskConfig specifies how findings should be masked when returned by the API.
type MaskConfig struct {
	MaskingChar             string   `json:"maskingChar"`
	CharsToIgnore           []string `json:"charsToIgnore"`
	NumCharsToLeaveUnmasked int      `json:"numCharsToLeaveUnmasked"`
	MaskLeftToRight         bool     `json:"maskLeftToRight"`
}

// InfoTypeSubstitutionConfig specifies that findings should be masked with the name of the matched info type.
type InfoTypeSubstitutionConfig struct {
}

// SubstitutionConfig specifies that findings should be masked with a configured custom phrase.
type SubstitutionConfig struct {
	SubstitutionPhrase string `json:"substitutionPhrase"`
}

// CryptoConfig specifies that findings should be encrypted with the provided RSA public key.
type CryptoConfig struct {
	PublicKey string `json:"publicKey"`
}

