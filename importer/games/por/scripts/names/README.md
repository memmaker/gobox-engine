# ECL variable names

`cmd/import` reads these when it writes `../area<N>.txt`
(`go run ./cmd/import` from the repo root):
the variables are named by the layout (`../layout/`, gobox ecl
Layout: a field its name, element i of an array `Name_i`), the one source of
their names. `area<N>.txt` names one area's constant tables (its code-space
data) and `labels/area<N>.txt` its jump labels. Lines are `XXXX Name  # evidence`:
the script address in hex, a PascalCase name, an optional comment.

The engine reads the same files (without the labels) when it parses the
scripts, and also:

- `areas.txt`: the areas, `ID SET Name` (NewECL's number, the data set
  holding it). `EnterArea Name` and `NewECL Name` use these.
- `values.txt`: named numbers made from the game data by `cmd/import`
  (monsters, item types), `Kind Set Number Name`. Do not edit; it is
  rewritten.
- `pictures.txt`: named numbers by hand (pictures, sprites, combat icons,
  portrait heads and bodies, encounter reactions, maps, wall sets,
  `LoadMap`'s icons), same format; set 0 is every set.
- `variables.txt`: by hand, the values and flag bits of variables, the
  variable's name as the kind (`CastleFlags 0 1 Disguised`): a number
  beside the variable (`CastleFlags |= Disguised`, `if Mode == Disguised`)
  reads as that name. A group has one comment naming its evidence (the
  code that reads the variable). A value whose name is also a variable,
  label, table or command stays a number.
- `quests.txt`: by hand, the quest journal's quests, `Quest 0 ID Name`
  (`rules/quests.go` by id): `Quest(Name)` hands one over.
- `values.txt` also has the shops' stocks, `Stock Set Block Item,Item,...`,
  which `OpenShop(Item, Item*20, ...)` writes out.

A command's operand of a kind (`LoadMonster`'s monster and icon,
`Picture`, `SetupMonster`'s sprite and picture, `Carried`'s item,
`ShowPortrait`'s head and body, `LoadMap`'s map, `LoadWalls`' wall sets,
`NewECL`'s area) reads as the name of its value; one without a name stays a
number. `../README.md` explains the script format and the commands.
