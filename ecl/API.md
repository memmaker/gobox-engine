# gobox/ecl

Shared core of the ECL area-script interpreter (cotab, sotsb). Synchronous,
no goroutines, no game knowledge: the address map and the per-game opcodes
are handed in.

## API

- `AddrLayout{Area, Area2, Aux, Code Range}`, `AddrLayout.KindOf(a)` -> `AddrArea|AddrArea2|AddrAux|AddrCode|AddrSpecial`.
  Both games: Area 19200..20223, Aux 31232..31743, Area2 31744..32767, Code 32768..40447.
- `Host` (the game implements): `Get(ref)`, `Set(ref, v)`, `Text(ref)`, `SetText(ref, s)`, `Random(n)`, `Exit(vm)`.
  A `Ref` is a field of the running script's layout (`vm.Script.Layout`, set by `ResolveRefs`); a variable
  operand without one panics (the script was not resolved). Games whose state is still the original's memory map a Ref
  back with `Layout.Address`.
- `New(host, layout, code) *VM`; `vm.Code` is the script bytes (Code[0] is address Code.First),
  `vm.PC`, `vm.ArgType/ArgN/ArgRef/Text`, `vm.Flags`, `vm.Calls`, `vm.Command/LastCommand`, `vm.Stop`, `vm.Remainder`
  (DIVIDE's field; the zero Ref keeps none).
- `vm.ReadArgs(n)`, `vm.Arg(k, h)`, `vm.Ref(k)`, `vm.Label(k)`, `vm.Call(k)`, `vm.SetFlags`, `vm.SetTextFlags`,
  `vm.CodeByte/SetCodeByte`, `vm.Skip(h)` (steps over the command at PC using the data table `vm.Args[cmd]`).
- `vm.Ops [256]Op` (`func(vm *VM, h Host)`), filled by `Default()` in `New`; the game overrides entries.
- `vm.Step(h)` runs one command; `vm.Run(h, halt)` loops until Stop or `halt()`.
- Shared opcodes: GOTO GOSUB COMPARE ADD SUB DIV MUL RANDOM SAVE RETURN COMPARE-AND IF* ON-GOTO ON-GOSUB GET-TABLE AND OR SAVE-TABLE.
  DIVIDE writes its remainder with `Host.Set(vm.Remainder, ..)` (cotab/sotsb: the field at 31744+831, DivideRemainder). Checked in both games: word 831 has no case in charValue/setCharValue (cvPartyCount is 830; memo, coin and wall ranges end below or sit elsewhere), so SetVar(area2 word 831) equals the original raw putWord; no SetRaw needed.
- `vm.Trace` hook (replaces sotsb's scriptTrace; `ArgsTrace` gives the same output).

## Suspension

Games differ: cotab re-runs from a checkpoint and replays Decisions (a blocked
`ask` is inside the Host/opcode); sotsb passes `then` callbacks (`runEvent(entry, then)`).
This module owns neither. Per-game opcodes that ask the player (menus, input,
combat) stay in the game and override `vm.Ops`; the game's own event loop calls
`vm.Step` (or `Run`) and resumes however it already does. The `then`, checkpoint
and event-stage fields stay in the game's own state next to the VM.

## Stays per game

Everything touching game state: EXIT (closure over pictures/msg position/selection; also `Host.Exit`),
PRINT, LOAD CHARACTER/MONSTER, SETUP MONSTER, APPROACH, PICTURE, INPUT, menus, CLEAR MONSTERS, PARTY STRENGTH,
NEW ECL, COMBAT, TREASURE, ... (and sotsb's GAIN EXP, 65), getVar/setVar (char values, specials, view flags),
initScript/loadScript, event loop/restart/adventure loop. Several are text-identical in cotab and sotsb
(Approach, LoadMonster, SetupMonster, NewECL, PartyStrength, ClearMonsters, PrintReturn, Clock, Protection, Rob ...)
but read Engine fields (party, monsters, sc.*); they could join once those sit behind Host, not before.

## Migrate

Both games (cotab-go, sotsb-go `engine/`):

1. go.mod: `require gobox v0.0.0`, `replace gobox => ../gobox-engine` and `replace gobox/rules => ../gobox-engine/rules` (one module: every package below imports as `gobox/<dir>`).
2. `eclvm.go`: delete scriptArea*/Code* consts, addrKindOf, readArgs, arg, varAddr, letterOf, unpackText, codeByte/setCodeByte,
   compareValues/Texts, scriptCall (about 170 of 322 lines). Keep `script` but replace `code`, `pc`, `argType/Lo/Hi`, `text`, `flags`, `calls`,
   `command`, `lastCommand`, `stop` with an embedded/owned `*ecl.VM` (`e.sc.vm`). `e.readArgs(n)` -> `e.sc.vm.ReadArgs(n)`,
   `e.arg(k)` -> `vm.Arg(k, e)`, `s.pc` -> `vm.PC` (mechanical, about 400 sites each; gopls rename of the fields helps).
   `code` becomes a `[]byte` of 7680 handed to `ecl.New`.
3. Engine implements Host: GetVar/SetVar = existing getVar/setVar, TextVar/SetTextVar = readTextVar/setTextVar (they
   use `vm.CodeByte` for the addrCode case), Random, Exit = cmdExit.
4. `eclcmds.go`: drop the 18 shared cmd funcs and their table lines; `init` becomes `vm.Ops[cmdX] = func(vm, h){ e.cmdX() }`
   for the remaining ones (or change their receiver). The `Args` table: from skipCommand's switch, fill `vm.Args[c]` once
   (sotsb adds cmd 65 to the 2-operand list), then `skipCommand` = `vm.Skip(e)`.
5. cotab commandTable bounds check is covered by the fixed 256 table; sotsb's trace -> `vm.Trace = ecl.ArgsTrace(e, ...)`.
6. Net: about 450 lines removed per game, ~600 touched call sites; run the fixture/replay tests for equality.

## Por

por's ecl.go/eclops*.go use int addresses and `eclByte`, with a different operand decoder. It could join by
(a) widening nothing: add an adapter Host that converts uint16 addresses to por's ints and a Layout with por's ranges, and
(b) moving its operand reading onto `ReadArgs`, then taking `Default()` opcodes that match por's semantics after a diff
(not done here: por was not diffed opcode by opcode).

## Cloning state

When a game deep-copies its state, it must rebind the Host the VM holds to the copy (cotab's clone does this); otherwise the cloned VM calls back into the original. A game whose own names clash with the API (e.g. Host.Exit) wraps them in a small adapter.
