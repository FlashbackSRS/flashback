package model

// Deck represents a single deck.
type Deck struct {
	Name           string
	ID             string
	TotalCards     int
	DueCards       int
	LearningCards  int
	MatureCards    int
	NewCards       int
	SuspendedCards int
}

// DeckList returns a complete list of decks available for study.
func (r *Repo) DeckList() ([]Deck, error) {
	return []Deck{
		{
			Name:           "All",
			ID:             "",
			TotalCards:     1142,
			DueCards:       120,
			LearningCards:  150,
			MatureCards:    400,
			NewCards:       577,
			SuspendedCards: 15,
		},
		{
			Name:           "Foo",
			ID:             "deck-asdf",
			TotalCards:     600,
			DueCards:       20,
			LearningCards:  50,
			MatureCards:    50,
			NewCards:       497,
			SuspendedCards: 3,
		},
		{
			Name:           "Bar",
			ID:             "deck-qwerty",
			TotalCards:     542,
			DueCards:       100,
			LearningCards:  100,
			MatureCards:    350,
			NewCards:       80,
			SuspendedCards: 12,
		},
	}, nil

}
