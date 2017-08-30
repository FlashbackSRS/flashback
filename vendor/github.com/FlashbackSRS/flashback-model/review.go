package fb

import (
	"encoding/json"
	"errors"
	"time"
)

// Review represents a single card-review event.
type Review struct {
	CardID    string    `json:"cardID"`
	Timestamp time.Time `json:"timestamp"`
	// Ease             ReviewEase     `json:"ease"`
	// Interval         *time.Duration `json:"interval"`
	// PreviousInterval *time.Duration `json:"previousInterval"`
	// SRSFactor        float32        `json:"srsFactor"`
	// ReviewTime       *time.Duration `json:"reviewTime"`
	// Type             ReviewType     `json:"reviewType"`
}

// Validate validates that all of the data in the review appears valid and self
// consistent. A nil return value means no errors were detected.
func (r *Review) Validate() error {
	if r.CardID == "" {
		return errors.New("card id required")
	}
	if _, _, _, err := parseCardID(r.CardID); err != nil {
		return err
	}
	if r.Timestamp.IsZero() {
		return errors.New("timestamp required")
	}
	return nil
}

type reviewAlias Review

// MarshalJSON satisfies the json.Marshaler interface.
func (r *Review) MarshalJSON() ([]byte, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(reviewAlias(*r))
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (r *Review) UnmarshalJSON(data []byte) error {
	doc := &reviewAlias{}
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}
	*r = Review(*doc)
	return r.Validate()
}

// type ReviewEase int
//
// const (
// 	ReviewEaseWrong ReviewEase = 1
// 	ReviewEaseHard  ReviewEase = 2
// 	ReviewEaseOK    ReviewEase = 3
// 	ReviewEaseEasy  ReviewEase = 4
// )
//
// type urReviewType int
//
// const (
// 	ReviewTypeLearn ReviewType = iota
// 	ReviewTypeReview
// 	ReviewTypeRelearn
// 	ReviewTypeCram
// )

// NewReview returns a new, empty Review for the provided Card.
func NewReview(cardID string) (*Review, error) {
	r := &Review{
		CardID:    cardID,
		Timestamp: now().UTC(),
	}
	return r, r.Validate()
}
