package ui

import (
	"errors"
)

// AskSure asks err's question when it is a Confirm or a ConfirmNo, and
// calls again on the answer that goes on; false: err is neither.
func AskSure(h RuleHost, err error, again func()) bool {
	var yes Confirm
	var no ConfirmNo
	switch {
	case errors.As(err, &yes):
		AskYes(h, string(yes), again)
	case errors.As(err, &no):
		h.Ask(&Question{Ask: FrontAsk{Kind: AskConfirm, Menu: Menu{Prompt: string(no)}}, Answer: func(a Action) bool {
			switch a.Kind {
			case ActNo:
				again()
			case ActYes, ActClose:
			default:
				return false
			}
			return true
		}})
	default:
		return false
	}
	return true
}
