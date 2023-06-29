package spinner_utils

import (
	"github.com/briandowns/spinner"
	"time"
)

type config struct {
	characterSet []string
	duration     time.Duration
}

type spinnerType struct {
	sleep   config
	apiCall config
}

var types = spinnerType{
	sleep: config{
		characterSet: spinner.CharSets[40],
		duration:     100 * time.Millisecond,
	},
	apiCall: config{
		characterSet: spinner.CharSets[14],
		duration:     100 * time.Millisecond,
	},
}

func NewSleepSpinner(options ...spinner.Option) *spinner.Spinner {
	s := spinner.New(types.sleep.characterSet, types.sleep.duration, options...)
	return s
}

func NewAPICallSpinner(options ...spinner.Option) *spinner.Spinner {
	s := spinner.New(types.apiCall.characterSet, types.apiCall.duration, options...)
	return s
}
