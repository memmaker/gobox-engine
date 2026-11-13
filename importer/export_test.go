package importer

import (
	"gobox/ecl"
	"gobox/rules1e"
)

// The unexported names the external tests use (package importer_test, which
// may import engine: engine imports importer).
var (
	EmbeddedScripts = embeddedScripts
	PodScriptData   = podScriptData
	PorStart        = porStart
	PorEnd          = porEnd
	OriginalRules   = originalRules
)

// originalRules is the ruleset's part of the original save b laid out as s
// whose script fields are l (a game's test against its real saves).
func originalRules(b []byte, s rules1e.SaveLayout, l *ecl.Layout) ([]byte, error) {
	o, err := rules1e.ReadOriginalSave(b, s)
	if err != nil {
		return nil, err
	}
	return rulesOf(&o, s, l)
}
