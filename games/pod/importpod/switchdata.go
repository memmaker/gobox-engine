package importpod

import (
	eclscript "gobox/ecl/script"
	"gobox/rules"
	"gobox/rules1e"
	"os"
	"path/filepath"
)

// Layout1e is the shape of this game's slot tables (rules1e.Layout).
func Layout1e() rules1e.Layout {
	return rules1e.Layout{SlotStride: spellLevels, SlotTotals: true, RangerMageColumn: rangerMageColumn, Races: rules1e.RaceNumbers(rules1e.Layout510)}
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
func JournalEntries() []eclscript.Entry {
	var es []eclscript.Entry
	for n := 1; n <= journalEntryMax; n++ {
		if t := journalTitle(byte(n)); journalBook().Has(t) {
			es = append(es, eclscript.Entry{ID: eclscript.EntryID(n), Name: "JournalEntry" + itoa(n), Book: journalFile, Title: t, Box: journalBox})
		}
	}
	return es
}

// WriteJournal copies the journal book and its pictures into dir.
func WriteJournal(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	files, err := journalFS.ReadDir("journal")
	if err != nil {
		return err
	}
	for _, f := range files {
		b, err := journalFS.ReadFile("journal/" + f.Name())
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, f.Name()), b, 0o644); err != nil {
			return err
		}
	}
	return nil
}
