package rules1e

// UnportedAffectHandlers is, per game, the handler names its affectVariants
// give that have no handler yet, each with what blocks it (fold T4c). The list
// only shrinks (TestAffectHandlersFull); a game switches to pool only with its
// list empty.
var UnportedAffectHandlers = map[string]map[HandlerName]string{
	"cotab": {},
	"sotsb": {},
	"pod":   {},
}
