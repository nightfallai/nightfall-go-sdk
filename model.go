package nightfall

const (
	DetectorTypeNightfallDetector DetectorType = "NIGHTFALL_DETECTOR"
	DetectorTypeRegex             DetectorType = "REGEX"
	DetectorTypeWordList          DetectorType = "WORD_LIST"

	MatchTypeFull    MatchType = "FULL"
	MatchTypePartial MatchType = "PARTIAL"

	ExclusionRuleTypeRegex    ExclusionType = "REGEX"
	ExclusionRuleTypeWordlist ExclusionType = "WORD_LIST"

	// A modifier that is used to decide when a finding should be surfaced in the context of a detection rule.
	// When ALL is specified, all detectors in a detection rule must trigger a match in order for the
	// finding to be reported. This is the equivalent of a logical "AND" operator.
	// When ANY is specified, only one of the detectors in a detection rule must trigger a match in order
	// for the finding to be reported. This is the equivalent of a logical "OR" operator.
	LogicalOpAny LogicalOp = "ANY"
	LogicalOpAll LogicalOp = "ALL"

	// Confidence describes the certainty that a piece of content matches a detector.
	ConfidenceVeryUnlikely Confidence = "VERY_UNLIKELY"
	ConfidenceUnlikely     Confidence = "UNLIKELY"
	ConfidencePossible     Confidence = "POSSIBLE"
	ConfidenceLikely       Confidence = "LIKELY"
	ConfidenceVeryLikely   Confidence = "VERY_LIKELY"
)

type (
	DetectorType  string
	MatchType     string
	ExclusionType string
	LogicalOp     string
	Confidence    string
)

// An object that contains a set of detectors to be used when scanning content. Findings matches are
// triggered according to the provided logicalOp; valid values are ANY(logical
// OR, i.e. a finding is emitted only if any of the provided detectors match), or ALL
// (logical AND, i.e. a finding is emitted only if all provided detectors match).
type DetectionRule struct {
	Name      string     `json:"name"`
	Detectors []Detector `json:"detectors"`
	LogicalOp LogicalOp  `json:"logicalOp"`
}

// A Detector represents a data type or category of information. Detectors are used to scan content
// for findings.
type Detector struct {
	DetectorUUID      string           `json:"detectorUUID,omitempty"`
	MinNumFindings    int              `json:"minNumFindings"`
	MinConfidence     Confidence       `json:"minConfidence"`
	DisplayName       string           `json:"displayName"`
	DetectorType      DetectorType     `json:"detectorType"`
	NightfallDetector string           `json:"nightfallDetector"`
	Regex             *Regex           `json:"regex,omitempty"`
	WordList          *WordList        `json:"wordList,omitempty"`
	ContextRules      []ContextRule    `json:"contextRules"`
	ExclusionRules    []ExclusionRule  `json:"exclusionRules"`
	RedactionConfig   *RedactionConfig `json:"redactionConfig,omitempty"`
}

// An object that describes a regular expression or list of keywords that may be used to disqualify a
// candidate finding from triggering a detector match.
type ExclusionRule struct {
	MatchType     MatchType     `json:"matchType"`
	ExclusionType ExclusionType `json:"exclusionType"`
	Regex         *Regex        `json:"regex"`
	WordList      *WordList     `json:"wordList"`
}

// An object that describes how a regular expression may be used to adjust the confidence of a candidate finding.
// This context rule will be applied within the provided byte proximity, and if the regular expression matches, then
// the confidence associated with the finding will be adjusted to the value prescribed.
type ContextRule struct {
	Regex                Regex                `json:"regex"`
	Proximity            Proximity            `json:"proximity"`
	ConfidenceAdjustment ConfidenceAdjustment `json:"confidenceAdjustment"`
}

// An object representing a regular expression to customize the behavior of a detector while Nightfall performs a scan.
type Regex struct {
	Pattern         string `json:"pattern"`
	IsCaseSensitive bool   `json:"isCaseSensitive"`
}

// A list of words that can be used to customize the behavior of a detector while Nightfall performs a scan.
type WordList struct {
	Values          []string `json:"values"`
	IsCaseSensitive bool     `json:"isCaseSensitive"`
}

// An object representing a range of bytes to consider around a candidate finding.
type Proximity struct {
	WindowBefore int `json:"windowBefore"`
	WindowAfter  int `json:"windowAfter"`
}

// Describes how to adjust confidence on a given finding. Valid values for the adjustment are
// VERY_UNLIKELY, UNLIKELY, POSSIBLE, LIKELY, and VERY_LIKELY.
type ConfidenceAdjustment struct {
	FixedConfidence Confidence `json:"fixedConfidence"`
}

// A container for minimal information representing a detector. A detector may be uniquely identified by its UUID;
// the name field helps provide human-readability.
type DetectorMetadata struct {
	DisplayName  string `json:"name"`
	DetectorUuid string `json:"uuid"`
}
