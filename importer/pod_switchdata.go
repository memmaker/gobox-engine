package importer

import (
	"gobox/ecl"
	"gobox/rules"
	"gobox/rules1e"
)

// Layout1e is the shape of this game's slot tables (rules1e.Layout).
func PodLayout1e() rules1e.Layout {
	return rules1e.Layout{SlotStride: podSpellLevels, SlotTotals: true, RangerMageColumn: rangerMageColumn, Races: rules1e.RaceNumbers(rules1e.Layout510)}
}

// SoundKind is what SOUND's operand n plays: song n&^128 of the music, the
// digitized sound n, or nothing.
func SoundKind(n byte) (song int, music, digitized bool) {
	switch {
	case n >= soundMusicFlag:
		return int(n &^ soundMusicFlag), true, false
	case n >= soundDigitizedMin && n <= soundDigitizedMax:
		return 0, false, true
	}
	return 0, false, false
}

// AreaMonsterTable is the monsters a fight adds by area screen (monster, icon),
// how many, and the status and control they get.
func AreaMonsterTable() (monsters, icons []byte, count, status, control int) {
	return areaMonsters[:], areaMonsterIcons[:], areaMonsterCount, areaMonsterStatus, areaMonsterControl
}

// Dialect is the game's script byte-code dialect.
const Dialect = rules.DialectPOD

// JournalEntries maps each journal entry the book has (no 13) to its text:
// the scripts' Journal n.
func JournalEntries() []ecl.Entry {
	var es []ecl.Entry
	for n := 1; n <= journalEntryMax; n++ {
		if t := journalTitle(byte(n)); journalBook().Has(t) {
			es = append(es, ecl.Entry{ID: ecl.EntryID(n), Name: "JournalEntry" + itoa(n), Book: journalFile, Title: t, Box: journalBox})
		}
	}
	return es
}
