<!-- claude-reflect: learned rules (saved from the band; edit freely) -->
## Learned rules
- After completing each milestone, update the checklist and commit your changes.
- Do not use raw memory blocks or offsets in this codebase; represent every value as a typed variable.
- Never scale images, textures, or other assets without preserving their aspect ratio.
- Do not implement map edge wrapping; game maps never wrap at their edges, so treat wrapping as a bug.
- Every screen fills the whole window or screen: no black bars or borders on any side, in windowed mode, in fullscreen and after any resize. The 40x25 page is the lower bound; larger windows give more cells and every screen lays itself out to them (winRight, winBottom, promptRow). Windowed mode is locked to the 16:10 aspect of the title images. Fullscreen keeps the screen's own aspect: UI pages adapt to it, and original pictures (title images, backdrops) are the one exception, shown centred with bars rather than cropped or stretched.
- The UI can be used with keyboard only, mouse only and gamepad only, and with any mixture of them: every action is reachable from each device on its own, and switching devices mid-screen just works. A keyboard is always assumed present, so these exceptions stand until a port to a platform without one: typing text, the combat camera (turn, pan, recentre, zoom), opening the battle log, the font switch (Ctrl-F) and opening help outside the command menu may require the keyboard (or, for zoom, the mouse).
- The input that opens or changes a screen never acts on the screen it opens: no mouse click, key or pad button falls through. Anything held across a screen change is ignored until it is released and pressed again (cmd/por/latch.go, TestLatch). New screens and new input paths must go through the latch, never read raw held state (IsMouseButtonPressed, KeyPressDuration, the pad's durations) directly.
- One glyph cell size for all UI: a text glyph and one tile of the red ribbon border are both 8·K window pixels, K = max(1, min(windowW/400, windowH/250)) whole, as the character sheet and the inventory use (invW x invH, invdraw.go). Every screen, menu, pop-up, question, tooltip and box draws its text and ribbon at that K; a larger window gives more cells, not bigger ones. Text boxes and menus have no size setting; the Graphics sizes that remain (portraits, minimap, picture box) step in whole multiples of K only (1x-4x). Shrinking a box that would not fit the window is the only other departure (scale, fitScale).
- One source of truth for every piece of state: each piece has exactly one owner (the ruleset, the engine, the adventure's data, or the ui for presentation only) and is stored exactly once, in the owner's types. Everyone else reads it from the owner when it is needed. No intermediate copy in another type, no snapshot, nothing derived stored; a cache is allowed only when it is keyed by the owner's version, written by no one but its own loader, and covered by a test.
- Never draw UI or import UI code from game modules; all UI and frontend comes from gobox transparently, and games must not know a UI is drawn.
- Have all four Pool games select portraits and combat icons from one shared collection that every game contributes to, and deduplicate them.
- Games must never draw or assemble portraits; all portrait visual logic belongs only in gobox.
- Define font colors only in gobox; never override them in any game.
- Before remastering a wall, check whether an upscaled version already exists (the pool series reuses walls) to avoid paying for new API calls.
- Do not give games their own input layer; route all input handling through gobox.
- Make the map ('m') always show the current area's name in the top left, and ensure every game sets its current area name.
- Never run scripted game sessions in fullscreen; keep scripted runs windowed.
- When an agent identifies a bug, have it report how to detect the same error class across all other games and instances, then broadcast that check so every agent scans its codebase.
- Keep one canonical implementation for each concern across the codebase; never duplicate shared infrastructure inside individual games.
- Design games so their default state is waiting for user input, never blocked mid-action; do not use coroutines or callbacks.
- Never add per-game settings to gobox; it must stay game-agnostic and make no game-specific differences.
<!-- end claude-reflect -->
