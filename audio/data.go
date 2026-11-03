package audio

import (
	"encoding/json"
	"fmt"
)

// The JSON cmd/sounddata writes. Ops name their kind and field as strings.
type jsonData struct {
	Program        []jsonOp                   `json:"program"`
	Speaker        [][sndChannels]int         `json:"speaker"`
	Tandy          [][sndChannels]int         `json:"tandy"`
	Vibratos       [][]int8                   `json:"vibratos"`
	Envelopes      [][envelopeSteps][2]uint16 `json:"envelopes"`
	SpeakerPeriods *[notesPerOctave]uint16    `json:"speakerPeriods"`
	TandyPeriods   *[notesPerOctave]uint16    `json:"tandyPeriods"`
}

type jsonOp struct {
	Kind   string `json:"k,omitempty"`
	Field  string `json:"f,omitempty"`
	Chan   int    `json:"c,omitempty"` // a note's channel (sndOp.field)
	Arg    int    `json:"a,omitempty"`
	Length int    `json:"l,omitempty"`
}

// data is a game's sound tables, decoded.
type data struct {
	program                      []sndOp
	speaker, tandy               [][sndChannels]int
	vibratos                     [][]int8
	envelopes                    [][envelopeSteps][2]uint16
	speakerPeriods, tandyPeriods *[notesPerOctave]uint16
}

var kindNames = map[string]sndKind{"set": sndSet, "call": sndCall, "return": sndReturn, "switch": sndSwitch, "loop": sndLoop, "note": sndNote}

var fieldNames = map[string]sndField{
	"": sndTimer, "sndTimer": sndTimer, "sndPeriod": sndPeriod, "sndSlide": sndSlide, "sndLevel": sndLevel,
	"sndLevelSlide": sndLevelSlide, "sndVibrato": sndVibrato, "sndVibratoSpeed": sndVibratoSpeed,
	"sndVibratoDepth": sndVibratoDepth, "sndVibratoLength": sndVibratoLength, "sndCounter": sndCounter,
	"sndCounter2": sndCounter2, "sndTempo": sndTempo, "sndTranspose": sndTranspose, "sndEnvelope": sndEnvelope,
}

func parseData(b []byte) (*data, error) {
	var j jsonData
	if err := json.Unmarshal(b, &j); err != nil {
		return nil, fmt.Errorf("audio: sound data: %w", err)
	}
	d := &data{speaker: j.Speaker, tandy: j.Tandy, vibratos: j.Vibratos, envelopes: j.Envelopes,
		speakerPeriods: j.SpeakerPeriods, tandyPeriods: j.TandyPeriods}
	for _, o := range j.Program {
		if o.Kind == "" { // op 0 ends a script
			d.program = append(d.program, sndOp{})
			continue
		}
		k, ok1 := kindNames[o.Kind]
		f, ok2 := fieldNames[o.Field]
		if k == sndNote {
			f, ok2 = sndField(o.Chan), true
		}
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("audio: sound data: unknown op %+v", o)
		}
		d.program = append(d.program, sndOp{kind: k, field: f, arg: o.Arg, length: o.Length})
	}
	if len(d.speaker) != len(d.tandy) || len(d.speaker) == 0 {
		return nil, fmt.Errorf("audio: sound data: song tables of %d and %d", len(d.speaker), len(d.tandy))
	}
	return d, nil
}
