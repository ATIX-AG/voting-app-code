package vote

import wr "github.com/mroth/weightedrand"

type VoteChooser struct {
	chooser *wr.Chooser
}

type Voter interface {
	Pick() string
}

func NewChooser(choices []string, preference string) (*VoteChooser, error) {
	weightedChoices := make([]wr.Choice, len(choices))
	for i, o := range choices {
		weight := 1
		if preference == o {
			weight = 3
		}
		weightedChoices[i] = wr.Choice{Item: o, Weight: uint(weight)}
	}

	chooser, err := wr.NewChooser(weightedChoices...)

	if err != nil {
		return nil, err
	}

	vc := VoteChooser{
		chooser: chooser,
	}
	return &vc, nil
}

// chooseVote chooses randomly one string from an array of strings
// It returns the chosen string or an empty string if no choices are found
func (c VoteChooser) Pick() string {
	return c.chooser.Pick().(string)
}
