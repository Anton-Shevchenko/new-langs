package spellio

import "time"

type root struct {
	Software               software                `json:"software"`
	Warnings               warnings                `json:"warnings"`
	Language               language                `json:"language"`
	Matches                []match                 `json:"matches"`
	SentenceRanges         [][]int                 `json:"sentenceRanges"`
	ExtendedSentenceRanges []extendedSentenceRange `json:"extendedSentenceRanges"`
}

type software struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	BuildDate   time.Time `json:"buildDate"`
	ApiVersion  int       `json:"apiVersion"`
	Premium     bool      `json:"premium"`
	PremiumHint string    `json:"premiumHint"`
	Status      string    `json:"status"`
}

type warnings struct {
	IncompleteResults bool `json:"incompleteResults"`
}

type language struct {
	Name             string           `json:"name"`
	DetectedLanguage detectedLanguage `json:"detectedLanguage"`
}

type detectedLanguage struct {
	Name       string  `json:"name"`
	Code       string  `json:"code"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
}

type match struct {
	Message                     string        `json:"message"`
	ShortMessage                string        `json:"shortMessage"`
	Replacements                []replacement `json:"replacements"`
	Offset                      int           `json:"offset"`
	Length                      int           `json:"length"`
	Context                     context       `json:"context"`
	Sentence                    string        `json:"sentence"`
	MatchType                   matchType     `json:"type"`
	Rule                        rule          `json:"rule"`
	IgnoreForIncompleteSentence bool          `json:"ignoreForIncompleteSentence"`
	ContextForSureMatch         int           `json:"contextForSureMatch"`
}

type replacement struct {
	Value string `json:"value"`
}

type context struct {
	Text   string `json:"text"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

type matchType struct {
	TypeName string `json:"typeName"`
}

type rule struct {
	ID          string   `json:"id"`
	Description string   `json:"description"`
	IssueType   string   `json:"issueType"`
	Category    category `json:"category"`
	IsPremium   bool     `json:"isPremium"`
	Confidence  float64  `json:"confidence"`
}

type category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type extendedSentenceRange struct {
	From              int                `json:"from"`
	To                int                `json:"to"`
	DetectedLanguages []detectedLanguage `json:"detectedLanguages"`
}
