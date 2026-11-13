package importer

// The unexported names the external tests use (package importer_test, which
// may import engine: engine imports importer).
var (
	EmbeddedScripts = embeddedScripts
	PodScriptData   = podScriptData
	PorStart        = porStart
	PorEnd          = porEnd
	OriginalRules   = originalRules
)
