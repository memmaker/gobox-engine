package ecl

// v1ExitOverrides correct the heuristic by area and decoded instruction number:
// the row, or NoExitRow where every row is a real choice.
var v1ExitOverrides = map[ExitPlace]int{
	// the rows are answers to a question, NO no way out
	{25, 167}: NoExitRow, // ARE YOU THE DIPLOMATIC ENVOYS?
	{13, 340}: NoExitRow, // ARE YOU VAMPIRES?
	{28, 330}: NoExitRow, // IS BISHOP BRACCIO THE POWER? YES, NO, DON'T KNOW
	{28, 359}: NoExitRow, // IS THE COUNCIL CORRUPT? YES, NO, DON'T KNOW
	// a shared YesNoMenu: callers differ, and a NO can start a fight
	{4, 49}: NoExitRow, {7, 0}: NoExitRow, {8, 559}: NoExitRow, {9, 157}: NoExitRow, {10, 922}: NoExitRow,
	{11, 268}: NoExitRow, {17, 196}: NoExitRow, {21, 453}: NoExitRow, {22, 9}: NoExitRow, {23, 10}: NoExitRow,
	// L is LISTEN's hotkey first: LEAVE has no key of its own
	{18, 324}: NoExitRow,
	// LEAVE is taking the way out; staying declines it
	{25, 461}: 1, // LEAVE CAVE, STAY IN
	{27, 453}: 0, // STAY, LEAVE: a passage going out
	// no label says so
	{26, 211}: 1, // TAKE BOAT, STAY
	{27, 150}: 1, // TAKE BOAT, STAY
	{0, 81}:   0, // GO, STAY: told to move along
}
