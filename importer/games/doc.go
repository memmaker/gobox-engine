// Package games is the importers of the four pool games (Pool of Radiance,
// Curse of the Azure Bonds, Secret of the Silver Blades, Pools of Darkness):
// each registers its importer.Game, so `gobox import <game>` (importer.Run)
// selects the game by name. Files are named <game>_<topic>.go for what one
// game does alone (names there carry the game's prefix where two games'
// names would clash) and shared_<topic>.go or the first game's file for what
// the games have in common. A game's embedded data lives in its folder
// (por/, cotab/, pod/, sotsb/): journals, area script names and layouts,
// tables.json. The importers may know their game; nothing at run time may.
//
// It imports engine, so engine cannot import it: it sits under importer.
package games
