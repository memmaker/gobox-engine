package core

// WordName names one of the word-sized values a script and the engine read
// and write in the ruleset's state: the area and game record words
// (ScriptWordNames) and the few places the engine itself keeps there.
type WordName string

// The words outside a script's own table.
const (
	WordPartyCount WordName = "PartyCount" // the game record's party count
	WordSymbolSet  WordName = "SymbolSet"
	WordBeginFlag  WordName = "BeginFlag"
	WordDebug      WordName = "Debug"
	WordTraining   WordName = "Training" // read only: the training hall is open
	WordPartyExp   WordName = "PartyExp"
	WordWho        WordName = "Who" // the member the rules act for (the actor of a call)
	WordAreaGeo    WordName = "Geo"
	WordEffect     WordName = "EffectTable" // read only, Index: the party's affect table
)

// ScriptWordNames are the plain record words a script reads and writes by
// name; the Sky words are PortraitWordNames, a write there redraws the portrait.
var ScriptWordNames = []WordName{
	"MapX", "MapY", "Geo", "TenMinutes", "Hours", "NoMagic", "In3D", "WallSplitA", "WallSplitB",
	"PrevX", "PrevY", "AreaKind", "NoCoords", "MsgDelay", "ViewFlags",
	"MarkFlags4", "Mark3", "Mark4Off", "MarkFlags57", "RiverOpen", "Boss", "Mark4",
	"PartyExp", "MaxDist", "MonsterDist", "Morale", "CombatEnd", "Beaten", "HoldMove", "Search", "Surprise",
	"RestEvery", "RestChance", "LeftMap", "Portrait", "Temple", "NoLoot", "Duel", "AttackedAlly", "Shop",
	"MonsterHit", "PartyHit", "PartyMove",
}

var PortraitWordNames = []WordName{"Sky", "SkyInside"}

type (
	// WordIs asks the value of a word (Answer.Value; Yes: it exists, and for
	// WordTraining and WordDebug the flag).
	WordIs struct {
		Name  WordName
		Index int
	}
	// SetWord stores a word.
	SetWord struct {
		Name  WordName
		Value int32
	}
)

func (WordIs) check()   {}
func (SetWord) action() {}
