package rules1e

// AffectType is an affect number: the Type of an affect record and the
// index into the ovr13 handler table (setupAffectHandlers). Spell affects
// are named after the spell that gives them (spell record byte 10).
type AffectType uint16

const (
	affAny                   AffectType = 0x00 // blockAffect: whatever affect is being added
	AffBless                 AffectType = 0x01
	AffCurse                 AffectType = 0x02
	affBonusVsRace4          AffectType = 0x03
	affStudyingManual        AffectType = 0x04 // Manual of Bodily Health
	affDetectMagic           AffectType = 0x05
	AffBonusByTargetRace     AffectType = 0x06
	affTrainingManual        AffectType = 0x07
	AffProtEvil              AffectType = 0x08
	affProtGood              AffectType = 0x09
	affResistCold            AffectType = 0x0A
	AffCharmed               AffectType = 0x0B // Charm Person, charm gazes
	affEnlarged              AffectType = 0x0C
	affReduced               AffectType = 0x0D
	affFriends               AffectType = 0x0E
	affPoisonDamage          AffectType = 0x0F
	AffReadMagic             AffectType = 0x10
	affShield                AffectType = 0x11
	affHitBonusVsNames       AffectType = 0x12
	affFindTraps             AffectType = 0x13
	affResistFire            AffectType = 0x14
	affSilenced              AffectType = 0x15
	affSlowPoison            AffectType = 0x16
	affSpiritualHammer       AffectType = 0x17
	affDetectInvisible       AffectType = 0x18
	affInvisible             AffectType = 0x19
	affHitBonusVsTargetNames AffectType = 0x1A
	AffFeatherFall           AffectType = 0x1B
	affMirrorImage           AffectType = 0x1C
	affEnfeebled             AffectType = 0x1D // Ray of Enfeeblement
	affCoughing              AffectType = 0x1E // in a stinking cloud
	affHelpless              AffectType = 0x1F
	affAnimated              AffectType = 0x20 // Animate Dead
	AffBlind                 AffectType = 0x21
	affCauseDisease          AffectType = 0x22
	affPrayerBonus           AffectType = 0x23
	AffBestowCurse           AffectType = 0x24
	affBlink                 AffectType = 0x25
	affStrength              AffectType = 0x26
	AffHaste                 AffectType = 0x27
	affStinkingCloud         AffectType = 0x28 // the caster's cloud
	affProtMissiles          AffectType = 0x29
	AffSlow                  AffectType = 0x2A
	affWeakened              AffectType = 0x2B
	affSick                  AffectType = 0x2C
	affProtEvilRadius        AffectType = 0x2D
	affProtGoodRadius        AffectType = 0x2E
	affMinus4VsNames         AffectType = 0x2F
	affMinus4VsBugbearGnoll  AffectType = 0x30
	affPrayer                AffectType = 0x31
	affMummyDisease          AffectType = 0x32 // no healing
	affSnakeCharm            AffectType = 0x33
	AffHeld                  AffectType = 0x34 // Hold Person
	affAsleep                AffectType = 0x35
	affRepulsed              AffectType = 0x36
	AffPoisoned              AffectType = 0x37
	affTurnInvisible         AffectType = 0x38
	affRotting               AffectType = 0x39
	affHeldFast              AffectType = 0x3A // no moves
	AffRegenerating          AffectType = 0x3B
	AffFireResistance        AffectType = 0x3D
	affRegenerate            AffectType = 0x3E
	affResistNormalWeapons   AffectType = 0x3C // CotAB name
	affMinorGlobe            AffectType = 0x3F // CotAB minor_globe_of_invulnerability
	affPoisonBite            AffectType = 0x40
	affPoisonBite4           AffectType = 0x41
	affPoisonBite2           AffectType = 0x42
	affParalyse2d8           AffectType = 0x43
	affParalyseNotElf        AffectType = 0x44
	affParalyse1d9           AffectType = 0x45
	affPoisonBiteMinus2      AffectType = 0x46
	AffMonsterInvisible      AffectType = 0x47
	affCamouflaged           AffectType = 0x48
	affRake                  AffectType = 0x49
	affStrikeTied            AffectType = 0x4A
	affBloodDrain            AffectType = 0x4B
	affHoldTarget            AffectType = 0x4C
	affHoldWeakTarget        AffectType = 0x4D
	affGetUp                 AffectType = 0x4E
	affMagicDamage2d10       AffectType = 0x4F
	affAcidDamage            AffectType = 0x50
	affFear                  AffectType = 0x51
	affFearLifted            AffectType = 0x52
	affStoneGaze             AffectType = 0x53
	affCharmGaze             AffectType = 0x54
	affWeakeningTouch        AffectType = 0x55 // was drain one level (POR 0x55, PoD 101): weakens the target
	affWeakeningTouch2       AffectType = 0x56 // was drain two levels (POR 0x56, PoD 102): weakens the target
	AffBreathWeapon          AffectType = 0x58 // run by ovr23 breathWeapon
	affDisease               AffectType = 0x57
	affDisplaced             AffectType = 0x59
	affSaveBonusVsType0      AffectType = 0x5A
	affResistFireElectric    AffectType = 0x5B
	affUnknown5C             AffectType = 0x5C // nothing adds it; unnamed in CotAB too
	affHalveFire             AffectType = 0x5D
	affHalveVsWeaponFlags    AffectType = 0x5E
	affFallsDead             AffectType = 0x5F
	affOnlyMagicWeapons      AffectType = 0x60
	affSaveBonusVsType42     AffectType = 0x61
	affHeal3                 AffectType = 0x62
	affStandUpBleeding       AffectType = 0x63
	affRegenUnlessFireAcid   AffectType = 0x64
	affStartRegenerating     AffectType = 0x65
	affReviveFull            AffectType = 0x66
	affGaseousEscape         AffectType = 0x67
	affAvoidDistant          AffectType = 0x68
	affResist50              AffectType = 0x69
	affResist15              AffectType = 0x6A
	affBlock90               AffectType = 0x6B // elves: sleep and charm
	affBlockSleepCharm       AffectType = 0x6C
	affBlockHold             AffectType = 0x6D
	affBlockCold             AffectType = 0x6E
	affBlockPoisonHold       AffectType = 0x6F
	affBlockFire             AffectType = 0x70
	affResistFireLessPerDie  AffectType = 0x71
	affHalveElectric         AffectType = 0x72
	affHalveUnlessPlain      AffectType = 0x73
	affHalveVsMagicWeapon    AffectType = 0x74
	affHolyWaterDamage       AffectType = 0x75
	affHalveCold             AffectType = 0x76
	affOnlyMagicOrStrong     AffectType = 0x77
	affDodgeBoulders         AffectType = 0x78
	affSpitAcid              AffectType = 0x79
	affOilDamage             AffectType = 0x7A
	affHalveOrCancelUnmagic  AffectType = 0x7B
	affBlock30               AffectType = 0x7C // half-elves: sleep and charm
	affBlockMind             AffectType = 0x7D
	affGazeReflected         AffectType = 0x7E
	affReflectableGaze       AffectType = 0x7F // a mirror turns the stone gaze back
	// Markers of item powers named in the item's record (CotAB item_affect_N).
	AffItem0               AffectType = 0x80
	affItem1               AffectType = 0x81
	affItem2               AffectType = 0x82
	affItem5               AffectType = 0x85
	affItem6               AffectType = 0x86
	affItem8               AffectType = 0x88
	affItemA               AffectType = 0x8A
	affItemB               AffectType = 0x8B
	affGiantStrengthGirdle AffectType = 0x83
	affCursedAlignment     AffectType = 0x84
	affNeedsGiantStrength  AffectType = 0x87
	affEndsSpiritualHammer AffectType = 0x89
	AffLaidOnHands         AffectType = 0x8C // CotAB paladinDailyHealCast
	AffCuredDisease        AffectType = 0x8D // CotAB paladinDailyCureRefresh
)

// Affects only the later games have (fold T4b): one id per behaviour, from the
// union of CotAB, SotSB and PoD. An id named after its game is that game's
// handler where it differs from POR's and the other games'. The games'
// affectIds (affectids.go) map their numbers here.
const (
	affSticksToSnakes          AffectType = 0x8E  // cotab, sotsb, pod: Sticks to Snakes victim: each round a.Data -= attacks(Attacks[0]+Attacks[1]); when Data...
	affArmourBonusCotab        AffectType = 0x90  // cotab: Faerie Fire (spell 79): every AC value (60-AC encoding) +2, capped at 60
	affSuffocatingCotab        AffectType = 0x91  // cotab: engulfed victim: each call Data--; at Data 0 killed with "Suffocates" (status dead). Ad...
	affStunnedCotab            AffectType = 0x96  // cotab: Fumble (spell 86): lose the turn (endTurn = POR clearCombatBlock behaviour). Behaviour ...
	affConfusedCotab           AffectType = 0x98  // cotab: Confusion: d100 <=10 remove, turned, auto/charmed control, add runningAway(142, 10 roun...
	affEngulfsCotab            AffectType = 0x9E  // cotab: when both attacks hit (hits[1]==2) a living target without engulfed/suffocating: "engul...
	affUnaffectedByType32Cotab AffectType = 0xA3  // cotab: damage type bit 32 (breath): block affect and damage, "is unaffected"
	affMiscastingCotab         AffectType = 0xA4  // cotab: empty handler; castframes.go: caster with it fails half its spells (Dice(1,2)==1); comb...
	affVsKind3Cotab            AffectType = 0xA5  // cotab: target kind 3: damage = 3*1d12+4+strength damage bonus, hit +2
	affVsKind8Cotab            AffectType = 0xA6  // cotab: target kind 8: hit +3, damage +3
	affBerserkCotab            AffectType = 0xA7  // cotab: on add: auto, control 179 (if <=127 or charmed) else 178, target nearest combatant of e...
	affHalfDamageCotab         AffectType = 0xA8  // cotab: all damage halved (unconditional)
	affSaveVsElementsCotab     AffectType = 0xA9  // cotab: fire or cold damage: save vs spell succeeds and spell table field 8 (save allowed) !=0:...
	affElectricityHealsCotab   AffectType = 0xAA  // cotab: electricity: block affect+damage, HP += 8 (rolls 1d8 unused; no max cap)
	affArrowsScratchCotab      AffectType = 0xAB  // cotab: attacker weapon base itemFlags == 1 (arrows): damage = 1
	affSpitAcidCotab           AffectType = 0xAC  // cotab: O0C2 spit: target via spellTargetHook 65, range <7, 30%: "Spits Acid" damage = spitter ...
	affDeathRays               AffectType = 0xAD  // cotab, pod: eye tyrant: up to 4 tries, each ray once: range<=2 disintegrate (save breath), <=3 ston...
	affBreathAcidCotab         AffectType = 0xAE  // cotab: 3 breaths per combat (Data reset round 0), cone width 1 length 6 via spellTargetHook 0x...
	affCloudkill               AffectType = 0xAF  // cotab+sotsb: caster's cloudkill cloud; ending it clears the cloud (POR has stinking-cloud clearing o...
	affHugsCotab               AffectType = 0xB0  // cotab: attack HitRoll>=18: "hugs X", target gets engulfed(58) run now, attacker gets hugging(1...
	affBurnsCotab              AffectType = 0xB1  // cotab: target without resist fire(20), ice body(54), fire reduced(61): damage +1d6
	affSlingStonesCotab        AffectType = 0xB4  // cotab: attacker Slots[0] type 85 (cotab sling): damage = 1d6+1. POR: same rule for ItemHolyWat...
	affAvoidsBowsCotab         AffectType = 0xB5  // cotab: attacker Slots[0] long/short bow (87/88): 50% avoided (avoidDistantAttack). POR: boulde...
	affParalysingGazeCotab     AffectType = 0xB6  // cotab: target save vs paralysis or paralysed(52) permanent (time 0, data 255), "is paralyzed"
	affFireTouchCotab          AffectType = 0xB7  // cotab: target takes 2d8 cold|magic (named fire, deals cold)
	affGazeParalyseCotab       AffectType = 0xB8  // cotab: O0C2 gaze: target via spellTargetHook 36, "gazes...", save vs petrify or paralysed(52) ...
	affBreathFireCotab         AffectType = 0xB9  // cotab: 3 breaths/combat, cone width 3 length 9 (spellTargetHook 61), each target MaxHP fire|32...
	affMagicImmuneCotab        AffectType = 0xBA  // cotab: any affect or magic damage blocked
	affSlainByArrowCotab       AffectType = 0xBB  // cotab: attacker ammo is arrow type 28 with Words[2]==135 (slaying): disintegrated (status 8), ...
	affBreathFireCloseCotab    AffectType = 0xBC  // cotab: 50%, target range<2: ends turn, "Breathes Fire" 7 fire damage, half on breath save
	affThrowLightningCotab     AffectType = 0xBD  // cotab: rounds 0-3 only: bolt at target square 16d6 then along line reach 10 (each 16d6, save v...
	affNoMindAffectsCotab      AffectType = 0xBE  // cotab: block runningAway(142), quarterLess(29), feebleminded(68); dmgType bit 64: block all
	AffRangerVsGiants          AffectType = 0xBF  // ranger: + ranger level damage against giants (POR: CotAB's handler; sotsb/pod variants)
	affElectricityImmuneCotab  AffectType = 0xC0  // cotab: electricity blocked (POR resistFireElectric 0x5B also blocks Magic Missile with message...
	affRootedCotab             AffectType = 0xC1  // cotab: Entangle: Moves=0 (no attack change, unlike POR heldNoMoves)
	affBerserkFoeCotab         AffectType = 0xC2  // cotab: on add: goBerserk (as 77), side = target side ^1; on removal control 179->0 and side = ...
	affAlwaysInvisibleCotab    AffectType = 0xC3  // cotab: adds invisible(25) time 255 data 255
	affEngulfingCotab          AffectType = 0xC4  // cotab: holds target (data=index): each round sets 2 attacks 2d8, attacks it, ends turn; on end...
	affHuggingCotab            AffectType = 0xC8  // cotab: as engulfing(139) with 1 attack 2d8; frees target from engulfed
	affGiantStrengthCotab      AffectType = 0xCA  // cotab: Stronger spell: giant strength 21 in data, 1d4*10+40; applied in stats.go. POR has no s...
	affDispellingEvil          AffectType = 0xCC  // sotsb+pod: defender: attacker with CreatureFlags&enchanted gets hit-7
	affElectricityImmuneSotsb  AffectType = 0xCD  // sotsb: defender: dmgType&electricity -> blockAffect(any) (no damage). Like affBlockFire/affBlo...
	affBarkskin                AffectType = 0xCF  // sotsb+pod: defender: save+1 unless dmgType&magic; attacker hit-1 (AC bonus)
	affBefuddledSotsb          AffectType = 0xD3  // sotsb: Fumble: turn ends (endTurn), same as POR clearCombatBlock; POR has no Fumble so affHelp...
	affDispelEvil              AffectType = 0xD5  // sotsb: on hit: target with CreatureFlags&enchanted that fails save vs spell (bonus 0) is kille...
	affConfused                AffectType = 0xD6  // sotsb: each round d100: <=10 remove, flee (Frightened 10 rounds); <=60 lose turn "is confused"...
	affFireShieldHot           AffectType = 0xDA  // sotsb+pod: defender: cold -> save+2; fire not saved -> damage*2
	affFireShieldCold          AffectType = 0xDB  // sotsb+pod: defender: fire -> save+2; cold and !e.saved -> damage*2
	affGlobe                   AffectType = 0xDC  // sotsb+pod: blocks spells of level <=4 (spell table byte 1)
	affNeedsPlus3WeaponSotsb   AffectType = 0xDE  // sotsb: damage 0 unless attacker readied weapon Plus>=3
	affStoningTouch            AffectType = 0xE0  // sotsb+pod: on hit: target save vs petrify (0) or killed statusStoned "is turned to stone"
	affSwallowsWholeSotsb      AffectType = 0xE1  // sotsb: HitRoll>=75: target dies "swallowed whole", body under swallower, 50% belch msg
	affBreathFireCloseSotsb    AffectType = 0xE2  // sotsb: when combatRound!=0 and BaseAttack[0]>=round, target within 1: fire|breath 8 damage, ha...
	affFeebleminded            AffectType = 0xE3  // sotsb: Int and Wis = 3; in combat no cast, loses the memorised spell being cast
	affInvisibleToAnimals      AffectType = 0xE4  // sotsb: unseen, hit-4 when own CreatureFlags&69 and no detect invisibility (original tests the ...
	affConfusingGaze           AffectType = 0xE5  // sotsb: aims with spell 68, target within 1: save vs spell or Confused 1d10+2 rounds
	affHighlighted             AffectType = 0xE6  // sotsb+pod: Faerie fire: AC-2 both
	affMirror                  AffectType = 0xE7  // sotsb: marker: carries a mirror that reflects affect 59 gazes (POR checks item words instead)
	affDragonBreathImmune      AffectType = 0xE8  // sotsb+pod: dmgType&breath -> block
	affMagicalBootsSotsb       AffectType = 0xE9  // sotsb: AC+2 both
	affBaneWeapon              AffectType = 0xEA  // sotsb: target MonsterFlags&1: damage = 6d6+strength bonus, hit+1
	affDiseaseImmuneSotsb      AffectType = 0xEB  // sotsb: marker, empty
	affBerserk                 AffectType = 0xEC  // sotsb: computer control, attacks nearest of either side, no casting; removal: control back, Si...
	affMagicalStone            AffectType = 0xED  // sotsb+pod: save+1
	affFireHeals               AffectType = 0xEE  // sotsb: magic fire heals by damage; electricity adds Slow(3); all damage/affect blocked
	affBreathFireSotsb         AffectType = 0xEF  // sotsb: aim spell 68: fire|breath 7 damage, half on save vs breath
	affSotsb84                 AffectType = 0xF3  // sotsb: unnamed, empty handler
	affSotsb85                 AffectType = 0xF4  // sotsb: unnamed, empty handler
	affSotsb87                 AffectType = 0xF5  // sotsb: unnamed, empty handler
	affBoltsHarmlessSotsb      AffectType = 0xF7  // sotsb: damage 0 when attacker weapon base flags&128 (uses bolts)
	affFearless                AffectType = 0xF8  // sotsb+pod: blocks Frightened (111)
	affHardyAll                AffectType = 0xF9  // sotsb+pod: con save bonus (4-6:1..18+:5) vs paralysis, rod and spell = affSaveBonusVsType0 + affSa...
	affDelayedBlastSotsb       AffectType = 0xFA  // sotsb: on removal in combat: blast at stored square, radius 3, (mage level)d6 magic fire, half...
	affSotsb101                AffectType = 0xFB  // sotsb: unnamed, empty handler
	affSotsb102                AffectType = 0xFC  // sotsb: unnamed, empty handler
	affBreathWeapon2Sotsb      AffectType = 0xFE  // sotsb: as 83 with fire cone length 9
	affStunnedSotsb            AffectType = 0x100 // sotsb: loses turn
	affBerserkFoe              AffectType = 0x101 // sotsb: like Berserk; removal restores Side=data
	affFrightened              AffectType = 0x104 // cotab (RunningAway, merged: equal handlers)+sotsb+pod: on removal: control back, Turned=0 (stops fleeing)
	affZapsAttackers           AffectType = 0x105 // sotsb: attacker within 1 takes 2x dealt damage as magic
	affGiantStrengthSpell      AffectType = 0x106 // sotsb: sotsb: spell-given (spells1.go:499), message only; strength via statAffects. POR girdle...
	affElectricityImmunePod    AffectType = 0x108 // pod: blocks any damage/affect whose dmgType has electricity (4) (blockAffect(0)). Close to P...
	affBefuddledPod            AffectType = 0x110 // pod: loses its turn (endTurn). Close to POR affHelpless (0x1F clearCombatBlock); POR 0x1B is...
	affRakshasaProtectionPod   AffectType = 0x116 // pod: blocks (blockAffect(0)) any current spell whose spell-table level < 8 (POR 0x28 is stin...
	affNeedsPlus3WeaponPod     AffectType = 0x11C // pod: damage 0 unless attacker weapon Plus>=3; weaponless monster hurts only with MaxLevel>=9...
	affHalfFromPiercingPod     AffectType = 0x11F // pod: attacker weapon base byte 7 <=1 (piercing/cutting): damage /2
	affHalfFromCrushingPod     AffectType = 0x120 // pod: attacker weapon base byte 7 >=128 (crushing): damage /2
	affMagicalBootsPod         AffectType = 0x125 // pod: handler does nothing; used in charcalc.go:223 (speed group)
	affDiseaseImmunePod        AffectType = 0x127 // pod: blocks Diseased (34)
	affLuckStonePod            AffectType = 0x12A // pod: attacker weapon Plus<3: damage /2
	affHalfFromWeakPod         AffectType = 0x12B // pod: attacker weapon Plus<4: damage /2
	affSlainByBoltsPod         AffectType = 0x12C // pod: attacker readied ammo/weapon type itemTypeBolts2 -> c killed 'is slain'
	affBreathFireCone24d10     AffectType = 0x12D // pod: hookFireBreath: fire cone 3x9, damage 24d10+12, half save vs breath, 3 charges/fight, o...
	affWhiteBreathPod          AffectType = 0x12F // pod: cold cone 3x7, MaxHP, half on save, 3 charges
	affPod87                   AffectType = 0x130 // pod: empty handler
	affGreenBreathPod          AffectType = 0x132 // pod: poison breath cone, MaxHP damage half on save, 3 charges, ends turn
	affBlueBreathPod           AffectType = 0x133 // pod: lightning breath bolt, MaxHP, 3 charges
	affBlackBreathPod          AffectType = 0x134 // pod: acid cone 1x6, MaxHP, half on save, 3 charges
	affBulletRakePod           AffectType = 0x135 // pod: no flag test: d4==1 -> second-hand attacks 8, else 4
	affDragonFearAttackPod     AffectType = 0x136 // pod: each round (area): every foe Okay, fails save vs petrify -4, lacks 120 -> DragonFear(12...
	affFireDamage4d6Pod        AffectType = 0x13A // pod: no flag test, combat only: target takes 4d6 fire+magic, half on save vs spell
	affStunnedPod              AffectType = 0x13C // pod: loses its turn (endTurn)
	affFireDamage2d6Pod        AffectType = 0x13E // pod: as 104 with 2d6
	affParalysingGazePod       AffectType = 0x142 // pod: aim spell 10, dmgType gaze, reflectable; save vs petrify +3 or 126 -> else Held(52) 10 ...
	affVorpalWeaponPod         AffectType = 0x143 // pod: combat, critical hit, target MonsterFlags&32==0 -> killed 'is beheaded'
	affWeaponSequencePod       AffectType = 0x144 // pod: blow n (1..6) of second hand: damage = str bonus + dice of 2H sword 1d10, bastard 2d4, ...
	affDancingPod              AffectType = 0x146 // pod: no flag test: AC0/AC1 -4
	affFireResistDruidPod      AffectType = 0x147 // pod: fire soaked from pool data*12; data shrinks, affect removed at 0
	affDragonFearPod           AffectType = 0x148 // pod: no flag test: damage -2 (floor 0), to-hit -2
	affFireTouchPod            AffectType = 0x149 // pod: same as 108: 2d6 fire to target, half save vs spell
	affIronSkinPod             AffectType = 0x14A // pod: no flag test: AC0/AC1 +4
	affNeedsPlus2WeaponPod     AffectType = 0x14D // pod: damage 0 unless weapon Plus>=2; weaponless monster needs MaxLevel>=7
	affGazeImmunePod           AffectType = 0x14E // pod: dmgType&16 (gaze): Save = 20
)
