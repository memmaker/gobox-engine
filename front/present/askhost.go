package present

import (
	"errors"

	"gobox/front"
)

// AskSure asks err's question when it is a Confirm or a ConfirmNo, and
// calls again on the answer that goes on; false: err is neither.
func AskSure(h RuleHost, err error, again func()) bool {
	var yes front.Confirm
	var no front.ConfirmNo
	switch {
	case errors.As(err, &yes):
		AskYes(h, string(yes), again)
	case errors.As(err, &no):
		h.Ask(&front.Question{Ask: front.Ask{Kind: front.AskConfirm, Menu: front.Menu{Prompt: string(no)}}, Answer: func(a front.Action) bool {
			switch a.Kind {
			case front.ActNo:
				again()
			case front.ActYes, front.ActClose:
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
