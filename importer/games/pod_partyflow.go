package games

// AddCharacter adds roster file File (SilverBlades: from Silver Blades'
// save directory) to the party, when the party's rules allow.
type AddCharacter struct {
	File         string
	SilverBlades bool
}
