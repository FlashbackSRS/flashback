package fb

import (
	"testing"

	"github.com/flimzy/diff"
)

func TestNewReview(t *testing.T) {
	tests := []struct {
		name     string
		cardID   string
		expected *Review
		err      string
	}{
		{
			name: "validation fails",
			err:  "card id required",
		},
		{
			name:   "valid",
			cardID: "card-krsxg5baij2w4zdmmu.mViuXQThMLoh1G1Nlc4d_E8kR8o.0",
			expected: &Review{
				CardID:    "card-krsxg5baij2w4zdmmu.mViuXQThMLoh1G1Nlc4d_E8kR8o.0",
				Timestamp: now(),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := NewReview(test.cardID)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestReviewMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		review   *Review
		expected string
		err      string
	}{
		{
			name:   "invalid review",
			review: &Review{},
			err:    "card id required",
		},
		{
			name:     "valid",
			review:   &Review{CardID: "card-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0", Timestamp: now()},
			expected: `{"cardID":"card-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0", "timestamp":"2017-01-01T00:00:00Z"}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.review.MarshalJSON()
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.JSON([]byte(test.expected), result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestReviewUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Review
		err      string
	}{
		{
			name:  "invalid json",
			input: "invalid json",
			err:   "invalid character 'i' looking for beginning of value",
		},
		{
			name:  "invalid review",
			input: `{"cardID":"card-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0"}`,
			err:   "timestamp required",
		},
		{
			name:     "valid",
			input:    `{"cardID":"card-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0", "timestamp":"2017-01-01T00:00:00Z"}`,
			expected: &Review{CardID: "card-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0", Timestamp: now()},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &Review{}
			err := result.UnmarshalJSON([]byte(test.input))
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestReviewValidate(t *testing.T) {
	tests := []validationTest{
		{
			name: "no card id",
			v:    &Review{},
			err:  "card id required",
		},
		{
			name: "invalid card id",
			v:    &Review{CardID: "oink"},
			err:  "invalid ID type",
		},
		{
			name: "wrong card id type",
			v:    &Review{CardID: "note-Zm9v"},
			err:  "invalid ID type",
		},
		{
			name: "no timestamp",
			v:    &Review{CardID: "card-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0"},
			err:  "timestamp required",
		},
		{
			name: "valid",
			v:    &Review{CardID: "card-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0", Timestamp: now()},
		},
	}
	testValidation(t, tests)
}
