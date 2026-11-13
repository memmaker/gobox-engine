# Original routines without a `name = SEG:OFF` doc comment in rules/

Built from cotab-re/ghidra/functions.txt against every SEG:OFF mention in the Go tree (a mention anywhere in a routine's range counts as loose). Segment-end labels (None/emptyNN) and RTL are left out.

Finding: no rules routine the game reaches is unported. Every rules entry below is already in Go, either inlined into its caller (nested Pascal procedures, one-line helpers) or with a marker the checklist generator does not pick up (by name, trailing `// SEG:OFF`, or an offset a few bytes off the ghidra entry). The rest is drawing, file/DAX, keyboard, sound, debug and DOS memory code.

| routine | name | what it does | decision |
|---|---|---|---|
| O036:0042 | mem_check | memory check at start, exits if short | skip: DOS memory |
| O039:0000 | pO039_0000 | delay_or_key | skip: timing/input |
| O039:01DD | credits | credits screen | skip: drawing |
| O039:0632 | title_screen | title screen | skip: drawing |
| O03C:13F6 | pO03C_13F6 | nested store helper of an ECL command | ported inline (eclcmds2.go store) |
| O03C:33A2 | commandTable | ECL command dispatch table init | ported as the Go command table (eclrun/eclcmds) |
| O03F:0118 | copy_protection | code wheel copy protection | skip: deliberately left out (engine.go) |
| O042:0000 | pO042_0000 | unit init calling overlay inits | skip: overlay init |
| O04C:0442 | pO04C_0442 | load_mob: GetMem a record, O0A1:329C, return its item/affect lists | ported inline: loadMonsterRecord (eclcmds2.go) returns the record |
| O04C:077E | pO04C_077E | party index of a character | ported inline (records.go partyIndex) |
| O04C:07E4 | get_player_values | get_player_values | ported: charValue (eclvars.go), marker by name |
| O04C:18AE | pO04C_18AE | ECL menu wrapper: hotkeys then list menu | ported: scriptMenu/menuHotkeys (eclcmds.go) |
| O04C:193B | pO04C_193B | compare_strings: six comparison flags | ported: compareValues (eclvm.go, "O04C:1A11 / 193B") |
| O05A:02AF | pO05A_02AF | nested: area catches a target | ported inline (combatai.go) |
| O05D:0046 | pO05D_0046 | set_background_tile | ported inline in combatMapSetup (combatsetup.go) |
| O05D:03FC | pO05D_03FC | build_background_tiles_1 | ported inline (combatsetup.go) |
| O05D:04A1 | pO05D_04A1 | build_background_tiles_2 | ported inline (combatsetup.go) |
| O05D:051E | pO05D_051E | build_background_tiles_3 | ported inline (combatsetup.go) |
| O05D:06F6 | pO05D_06F6 | build_background_tiles_4 | ported inline (combatsetup.go) |
| O05D:09AC | pO05D_09AC | nested 1-in-20 roll | ported inline (combatsetup.go) |
| O05D:0CA2 | pO05D_0CA2 | scatter boulders/bushes | ported inline (combatsetup.go scatterOdds) |
| O05D:1C3B | battle_begins | battle_begins | ported: combatSetup (combatsetup.go), marker "O05D:battle_begins" has no offset |
| O064:1274 | pO064_1274 | empty affect handler | ported: no-op (nil handler) |
| O064:1520 | pO064_1520 | poison sting handler (bonus 0) | ported: poisonStingHandler, trailing comment marker |
| O064:1534 | pO064_1534 | poison sting handler (+4) | ported: poisonSting4Handler, trailing comment marker |
| O064:1548 | pO064_1548 | poison sting handler (+2) | ported: poisonSting2Handler, trailing comment marker |
| O064:1671 | pO064_1671 | poison sting handler (-2) | ported: poisonStingMinus2Handler, trailing comment marker |
| O064:170E | pO064_170E | empty affect handler | ported: no-op (nil handler) |
| O064:20DA | pO064_20DA | avoid distant attack 60% | ported (affects.go), trailing comment marker |
| O064:214F | pO064_214F | magic resistance 50% | ported: magicResist50Handler, trailing comment marker |
| O064:215D | pO064_215D | magic resistance 15% | ported: magicResist15Handler, trailing comment marker |
| O064:26D3 | pO064_26D3 | do_items_affect: readied item adds/removes its affect | ported: itemHandler (affects.go) |
| O08D:276D | spell_menu3 | spell_menu3: combat spell menu | presentation; menu ported in spellmenu.go |
| O08D:3027 | Aim_menu | Aim_menu | combat presentation (out of scope) |
| O08D:388F | pO08D_388F | copy_sorted_players | ported inline (combatturn2.go rebuild) |
| O08D:3932 | pO08D_3932 | step_combat_list | ported inline (combatturn2.go step) |
| O08D:4159 | pO08D_4159 | draw missile from attacker to target | skip: drawing |
| O08D:41C1 | pO08D_41C1 | aim line-of-fire check | ported inline (affects.go) |
| O08D:48D9 | god_intervene | god_intervene debug cheat | ported: godIntervene (combatturn2.go), marker by name |
| O099:003C | pO099_003C | empty procedure | skip: no-op |
| O09D:0000 | pO09D_0000 | unit init | skip: overlay init |
| O09D:0E89 | pO09D_0E89 | list menu node allocation | ported: Go slices (listmenu.go) |
| O09D:1F02 | pO09D_1F02 | player_is_okey: health status 0 | ported inline at callers (status checks) |
| O0A1:008B | pO0A1_008B | nested roster file reader | ported: rosterFiles (charimport.go) |
| O0A1:09FD | icon_xx | icon_xx | skip: icon drawing |
| O0A1:0A90 | pO0A1_0A90 | LoadPlayerCombatIcon | skip: icon drawing (cmd/cotab) |
| O0A1:37B5 | loadGame | loadGame | ported: LoadGame (saveload.go, O0A1:37A9) |
| O0A1:3DAB | pO0A1_3DAB | party memory size for the heap check | skip: DOS memory |
| O0A7:3C5B | pO0A7_3C5B | duplicateCombatIcon | skip: icon drawing |
| O0A7:3EC6 | icon_builder | icon_builder | skip: icon editor UI |
| O0AC:0000 | int_0x0c | int_0x0c | skip: drawing/screen/file/driver |
| O0AC:0030 | int_0x0d | int_0x0d | skip: drawing/screen/file/driver |
| O0AC:0068 | pO0AC_0068 | pO0AC_0068 | skip: drawing/screen/file/driver |
| O0AC:00B8 | pO0AC_00B8 | pO0AC_00B8 | skip: drawing/screen/file/driver |
| O0AC:04F7 | pO0AC_04F7 | pO0AC_04F7 | skip: drawing/screen/file/driver |
| O0AC:079B | pO0AC_079B | pO0AC_079B | skip: drawing/screen/file/driver |
| O0AC:085E | pO0AC_085E | pO0AC_085E | skip: drawing/screen/file/driver |
| O0AC:09F4 | pO0AC_09F4 | pO0AC_09F4 | skip: drawing/screen/file/driver |
| O0AF:0081 | playerDisplayFull | playerDisplayFull | skip: drawing/screen/file/driver |
| O0AF:0596 | pO0AF_0596 | pO0AF_0596 | skip: drawing/screen/file/driver |
| O0AF:0697 | display_player_stats01 | display_player_stats01 | skip: drawing/screen/file/driver |
| O0AF:0B70 | viewPlayer | viewPlayer | ported: viewPlayer (viewplayer.go, O0AF:0B64) |
| O0AF:10A6 | pO0AF_10A6 | ItemDisplayStats | skip: drawing |
| O0BA:0039 | pO0BA_0039 | remove_weight | ported inline (Weight -=, coins.go/items) |
| O0BA:0053 | pO0BA_0053 | add_weight | ported inline (Weight +=, coins.go/charcalc.go) |
| O0C2:1676 | pO0C2_1676 | nested: start the path line | ported inline in the bolt path (O11C newLine) |
| O0C2:1702 | pO0C2_1702 | nested find_players_on_path | ported inline in the bolt path |
| O0C2:27EB | create_noxious_cloud | create_noxious_cloud | ported: createCloud (clouds.go), marker by name |
| O0C2:4F95 | spell_poisonous_cloud | spell_poisonous_cloud | ported: createCloud (clouds.go), marker by name |
| O0C2:56CC | pO0C2_56CC | plain spell effect | ported: fxPlain (spells2.go, "569E, 56CC, 56FA") |
| O0C2:56FA | pO0C2_56FA | plain spell effect | ported: fxPlain (spells2.go) |
| O0C2:57F1 | cast_heal2 | cast_heal2: byte-for-byte cast_heal (2d4+2, "is Healed") | ported: spellFx[99] = fxHeal2d4 (spells1.go, marked O0C2:40C2 only) |
| O0E3:1771 | pO0E3_1771 | nested max_strength | ported inline (stats.go takeBetter) |
| O0E3:17BE | pO0E3_17BE | nested add levels | ported inline (stats.go addLevels) |
| O0ED:0798 | Player_Summary | Player_Summary | skip: drawing/screen/file/driver |
| O0ED:0933 | minus | minus | skip: drawing/screen/file/driver |
| O0ED:09BA | display_hp | display_hp | skip: drawing/screen/file/driver |
| O0ED:0A2D | pO0ED_0A2D | pO0ED_0A2D | skip: drawing/screen/file/driver |
| O0ED:1108 | pO0ED_1108 | ConcatWord | skip: string helper |
| O0ED:1142 | pO0ED_1142 | integer to string | skip: RTL Str |
| O0ED:186F | pO0ED_186F | pO0ED_186F | skip: drawing/screen/file/driver |
| O0ED:18A2 | pO0ED_18A2 | pO0ED_18A2 | skip: drawing/screen/file/driver |
| O0ED:1924 | pO0ED_1924 | pO0ED_1924 | skip: drawing/screen/file/driver |
| O0ED:1A59 | pO0ED_1A59 | pO0ED_1A59 | skip: drawing/screen/file/driver |
| O0ED:1AA4 | pO0ED_1AA4 | pO0ED_1AA4 | skip: drawing/screen/file/driver |
| O0ED:2B66 | camping_search | camping_search status line | presentation; referenced in adventure.go |
| O0ED:2DC0 | pO0ED_2DC0 | pO0ED_2DC0 | skip: drawing/screen/file/driver |
| O0ED:327F | pO0ED_327F | new item helper | ported inline (charimport.go) |
| O0FE:13B2 | pO0FE_13B2 | is_human | ported inline (classchange.go) |
| O104:0076 | pO104_0076 | list length | ported: len(list) (listmenu.go) |
| O104:01E9 | pO104_01E9 | pO104_01E9 | skip: drawing/screen/file/driver |
| O104:07EE | redraw_screen | redraw_screen | skip: drawing/screen/file/driver |
| O104:0804 | pO104_0804 | pO104_0804 | skip: drawing/screen/file/driver |
| O104:0897 | pO104_0897 | pO104_0897 | skip: drawing/screen/file/driver |
| O104:09D3 | pO104_09D3 | pO104_09D3 | skip: drawing/screen/file/driver |
| O104:0A1F | pO104_0A1F | pO104_0A1F | skip: drawing/screen/file/driver |
| O104:0A65 | pO104_0A65 | pO104_0A65 | skip: drawing/screen/file/driver |
| O104:0B07 | pO104_0B07 | pO104_0B07 | skip: drawing/screen/file/driver |
| O104:0C08 | pO104_0C08 | pO104_0C08 | skip: drawing/screen/file/driver |
| O104:0D38 | pO104_0D38 | pO104_0D38 | skip: drawing/screen/file/driver |
| O104:0DCA | pO104_0DCA | pO104_0DCA | skip: drawing/screen/file/driver |
| O104:1306 | pO104_1306 | pO104_1306 | skip: drawing/screen/file/driver |
| O109:0005 | pO109_0005 | pO109_0005 | skip: drawing/screen/file/driver |
| O109:002E | pO109_002E | pO109_002E | skip: drawing/screen/file/driver |
| O109:005D | pO109_005D | pO109_005D | skip: drawing/screen/file/driver |
| O10D:0014 | pO10D_0014 | pO10D_0014 | skip: drawing/screen/file/driver |
| O10D:0048 | pO10D_0048 | pO10D_0048 | skip: drawing/screen/file/driver |
| O10D:007C | pO10D_007C | pO10D_007C | skip: drawing/screen/file/driver |
| O10D:00BA | pO10D_00BA | pO10D_00BA | skip: drawing/screen/file/driver |
| O110:000A | pO110_000A | pO110_000A | skip: drawing/screen/file/driver |
| O110:008D | pO110_008D | pO110_008D | skip: drawing/screen/file/driver |
| O110:00AD | pO110_00AD | pO110_00AD | skip: drawing/screen/file/driver |
| O110:06DC | pO110_06DC | pO110_06DC | skip: drawing/screen/file/driver |
| O110:074D | Show3DSprite | Show3DSprite | skip: drawing/screen/file/driver |
| O110:087A | pO110_087A | pO110_087A | skip: drawing/screen/file/driver |
| O116:000F | pO116_000F | pO116_000F | skip: drawing/screen/file/driver |
| O116:0184 | pO116_0184 | pO116_0184 | skip: drawing/screen/file/driver |
| O116:0434 | pO116_0434 | pO116_0434 | skip: drawing/screen/file/driver |
| O116:0820 | pO116_0820 | pO116_0820 | skip: drawing/screen/file/driver |
| O121:00F2 | pO121_00F2 | pO121_00F2 | skip: drawing/screen/file/driver |
| O121:0130 | pO121_0130 | pO121_0130 | skip: drawing/screen/file/driver |
| O121:016E | pO121_016E | redraw a combat square | skip: drawing |
| O121:031C | pO121_031C | redraw a combat square and its occupant | skip: drawing |
| O121:03E7 | pO121_03E7 | setup_mapToPlayerIndex: who stands where | ported: Go keeps positions in cpos and looks them up (mapAt, cmap.go) |
| O121:0572 | pO121_0572 | RedrawPlayerBackground | skip: drawing |
| O121:0C32 | pO121_0C32 | combat X of a character | ported: combatX (cmap.go), trailing comment marker |
| O121:0C5A | pO121_0C5A | combat Y of a character | ported: combatY (cmap.go), trailing comment marker |
| O121:1356 | pO121_1356 | draw a combatant's icon state | skip: drawing |
| O12A:002C | Load24x24Set | Load24x24Set | skip: drawing/screen/file/driver |
| O12A:00F7 | pO12A_00F7 | pO12A_00F7 | skip: drawing/screen/file/driver |
| O12A:0156 | pO12A_0156 | pO12A_0156 | skip: drawing/screen/file/driver |
| O12A:01D0 | chead_cbody_comspr_icon | chead_cbody_comspr_icon | skip: drawing/screen/file/driver |
| O12A:0504 | pO12A_0504 | pO12A_0504 | skip: drawing/screen/file/driver |
| O130:023E | print_command | print_command | skip: drawing/screen/file/driver |
| O19C:0000 | pO19C_0000 | pO19C_0000 | skip: drawing/screen/file/driver |
| O19C:0045 | pO19C_0045 | pO19C_0045 | skip: drawing/screen/file/driver |
| O19C:0171 | pO19C_0171 | pO19C_0171 | skip: drawing/screen/file/driver |
| R133:0000 | draw8x8_01 | draw8x8_01 | skip: drawing/screen/file/driver |
| R133:00D3 | draw8x8_02 | draw8x8_02 | skip: drawing/screen/file/driver |
| R133:0136 | draw8x8_03 | draw8x8_03 | skip: drawing/screen/file/driver |
| R133:0252 | draw8x8_04 | draw8x8_04 | skip: drawing/screen/file/driver |
| R133:0292 | draw8x8_05 | draw8x8_05 | skip: drawing/screen/file/driver |
| R133:0397 | draw8x8_06 | draw8x8_06 | skip: drawing/screen/file/driver |
| R133:048D | draw8x8_07 | draw8x8_07 | skip: drawing/screen/file/driver |
| R133:04CD | draw8x8_clear_area | draw8x8_clear_area | skip: drawing/screen/file/driver |
| R1A0:0020 | pR1A0_0020 | read a digit key | skip: keyboard input |
| R226:0000 | pR226_0000 | pR226_0000 | skip: drawing/screen/file/driver |
| R226:007B | pR226_007B | pR226_007B | skip: drawing/screen/file/driver |
| R226:00CF | pR226_00CF | pR226_00CF | skip: drawing/screen/file/driver |
| R226:00FF | pR226_00FF | pR226_00FF | skip: drawing/screen/file/driver |
| R226:0111 | load_dax | load_dax | skip: drawing/screen/file/driver |
| R226:02A5 | turn_dax_to_videolayout | turn_dax_to_videolayout | skip: drawing/screen/file/driver |
| R226:08D7 | pR226_08D7 | pR226_08D7 | skip: drawing/screen/file/driver |
| R226:09C3 | pR226_09C3 | pR226_09C3 | skip: drawing/screen/file/driver |
| R226:0A51 | pR226_0A51 | pR226_0A51 | skip: drawing/screen/file/driver |
| R226:0AD6 | pR226_0AD6 | pR226_0AD6 | skip: drawing/screen/file/driver |
| R226:0B63 | pR226_0B63 | pR226_0B63 | skip: drawing/screen/file/driver |
| R226:0B94 | merge_icon | merge_icon | skip: drawing/screen/file/driver |
| R226:0EDB | ega_01 | ega_01 | skip: drawing/screen/file/driver |
| R226:10E7 | draw_picture | draw_picture | skip: drawing/screen/file/driver |
| R226:12E8 | ega_03 | ega_03 | skip: drawing/screen/file/driver |
| R226:15F3 | pR226_15F3 | pR226_15F3 | skip: drawing/screen/file/driver |
| R226:1C25 | init_dax_block | init_dax_block | skip: drawing/screen/file/driver |
| R226:1CEC | pR226_1CEC | pR226_1CEC | skip: drawing/screen/file/driver |
| R226:1D67 | pR226_1D67 | pR226_1D67 | skip: drawing/screen/file/driver |
| R226:2142 | pR226_2142 | pR226_2142 | skip: drawing/screen/file/driver |
| R226:217F | pR226_217F | pR226_217F | skip: drawing/screen/file/driver |
| R4CE:0000 | ega_04 | ega_04 | skip: drawing/screen/file/driver |
| R4CE:0071 | pR4CE_0071 | pR4CE_0071 | skip: drawing/screen/file/driver |
| R4CE:00A0 | pR4CE_00A0 | pR4CE_00A0 | skip: drawing/screen/file/driver |
| R4CE:00B7 | ega01 | ega01 | skip: drawing/screen/file/driver |
| R4CE:017C | load_8x8d1_201 | load_8x8d1_201 | skip: drawing/screen/file/driver |
| R4CE:01DF | display_char01 | display_char01 | skip: drawing/screen/file/driver |
| R4CE:030A | pR4CE_030A | pR4CE_030A | skip: drawing/screen/file/driver |
| R4CE:0349 | pR4CE_0349 | pR4CE_0349 | skip: drawing/screen/file/driver |
| R4CE:03CE | pR4CE_03CE | pR4CE_03CE | skip: drawing/screen/file/driver |
| R4CE:0444 | pR4CE_0444 | pR4CE_0444 | skip: drawing/screen/file/driver |
| R4CE:0A02 | pR4CE_0A02 | pR4CE_0A02 | skip: drawing/screen/file/driver |
| R4CE:0B2E | pR4CE_0B2E | pR4CE_0B2E | skip: drawing/screen/file/driver |
| R5BD:0000 | free_dax_raw | free_dax_raw | skip: drawing/screen/file/driver |
| R5BD:0020 | pR5BD_0020 | pR5BD_0020 | skip: drawing/screen/file/driver |
| R5BD:0040 | pR5BD_0040 | pR5BD_0040 | skip: drawing/screen/file/driver |
| R5BD:00AC | pR5BD_00AC | pR5BD_00AC | skip: drawing/screen/file/driver |
| R5BD:011C | pR5BD_011C | pR5BD_011C | skip: drawing/screen/file/driver |
| R5BD:013A | pR5BD_013A | pR5BD_013A | skip: drawing/screen/file/driver |
| R5BD:0194 | delete_file | delete_file | skip: drawing/screen/file/driver |
| R5BD:0208 | check_overlay_file | check_overlay_file | skip: drawing/screen/file/driver |
| R5BD:0259 | pR5BD_0259 | pR5BD_0259 | skip: drawing/screen/file/driver |
| R5BD:02E7 | find_and_open_file | find_and_open_file | skip: drawing/screen/file/driver |
| R5BD:0501 | file_find | file_find | skip: drawing/screen/file/driver |
| R5BD:0704 | decode_dax_raw | decode_dax_raw | skip: drawing/screen/file/driver |
| R5BD:07C0 | pR5BD_07C0 | pR5BD_07C0 | skip: drawing/screen/file/driver |
| R5BD:08B3 | pR5BD_08B3 | pR5BD_08B3 | skip: drawing/screen/file/driver |
| R5BD:08D8 | load_decode_dax | load_decode_dax | skip: drawing/screen/file/driver |
| R5BD:0B0C | pR5BD_0B0C | set_game_area | ported inline (eclvars.go) |
| R5BD:0B1F | restore_game_area | restore_game_area | skip: drawing/screen/file/driver |
| R670:0063 | pR670_0063 | debug dump of affects | skip: debug output |
| R670:0112 | debug_txt | debug_txt | skip: debug output |
| R670:0252 | clear_keyboard | clear_keyboard | skip: drawing/screen/file/driver |
| R670:026E | pR670_026E | pR670_026E | skip: drawing/screen/file/driver |
| R699:00D4 | pR699_00D4 | sound state | skip: sound |
