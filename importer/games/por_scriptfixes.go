package games

import (
	"gobox/ecl/script"
)

// scriptFixes are POR's special cases in its decompiler: each turns a site of
// the decompiled text into the port's (script.Fix, the site exactly once or
// the import fails). Each says what the original did, why it changes and its
// row in gobox-engine/CHANGES-FROM-ORIGINAL.md. They were patches/*.patch,
// byte for byte the same texts.
var porScriptFixes = []script.Fix{
	// 0: area 0 InsideTown. Original: nothing here.
	// Why: fix: the arms shop's south door (wall 3, not 9) was never locked.
	// CHANGES: POR script, area 0 InsideTown.
	{Area: 0, Old: `    if Scratch10 == 0 and WallAhead == 9 { goto LockedDoor }
    if Scratch10 == 4 and WallAhead == 7 { goto LockedDoor }
    if Scratch10 == 26 and WallAhead == 11 { goto LockedDoor }
    exit

    LockedDoor:
    Text("THE DOOR IS LOCKED.  DO YOU WANT TO BREAK IN?")
    YesNoMenu()
`, New: `    if Scratch10 == 0 and WallAhead == 9 { goto LockedDoor }
    if Scratch10 == 4 and WallAhead == 7 { goto LockedDoor }
    if Scratch10 == 26 and WallAhead == 11 { goto LockedDoor }
    if PartyX == 9 and PartyY == 14 { goto ArmsShopSouth } # fix: the arms shop's south door (wall 3, not 9) was never locked
    exit

    ArmsShopSouth:
    if Facing == North { goto LockedDoor }
    exit

    LockedDoor:
    Text("THE DOOR IS LOCKED.  DO YOU WANT TO BREAK IN?")
    YesNoMenu()
`},
	// 1: area 0 on step. Original: nothing here.
	// Why: (extra) Rolf tells the history of Phlan.
	// CHANGES: POR script, area 0 Rolf's history of Phlan.
	{Area: 0, Old: `    Portrait = None
    Scratch10 = ~Inside & Cell
    if TourDone < 1 { exit }
    PartyMemberFound = 0
    Scratch1 = Random(19)
    if Scratch1 == 19 {
`, New: `    Portrait = None
    Scratch10 = ~Inside & Cell
    if TourDone < 1 { exit }
    if PartyX == 2 and PartyY == 2 { goto RolfEvent } # (extra) Rolf tells the history of Phlan
    PartyMemberFound = 0
    Scratch1 = Random(19)
    if Scratch1 == 19 {
`},
	// 2: area 0 TempleOfTyr. Original: `goto SeekHealingAsk`; `goto SeekHealingAsk`.
	// Why: the temple asks nothing first (8679e68): the jump to SeekHealingAsk and its 'DO YOU SEEK HEALING?' print go, the healing menu (TempleHealing) opens directly.
	// CHANGES: POR script, area 0 shop and temple greetings.
	{Area: 0, Old: `    $AuxText = "MAD MAN"
    FindMemberByName()
    if PartyMemberFound != 0 { PriestWarnsMadMan() }
    goto SeekHealingAsk

TempleOfSune:
    ShowPortrait(BlondWoman, GreyApron2)
    if FoughtWatch > 0 { goto RefuseService }
    Text("YOU ARE WELCOMED BY PRIESTESS JOY OF SUNE.")
    goto SeekHealingAsk

RiverView:
    Text("YOU WALK NEAR THE ")
`, New: `    $AuxText = "MAD MAN"
    FindMemberByName()
    if PartyMemberFound != 0 { PriestWarnsMadMan() }
    goto TempleHealing

TempleOfSune:
    ShowPortrait(BlondWoman, GreyApron2)
    if FoughtWatch > 0 { goto RefuseService }
    Text("YOU ARE WELCOMED BY PRIESTESS JOY OF SUNE.")
    goto TempleHealing

RiverView:
    Text("YOU WALK NEAR THE ")
`},
	// 3: area 0 GeneralShop. Original: `goto GeneralShopBrowse`; ``; `GeneralShopBrowse:`.
	// Why: the shops open their menu directly (46b05b8): with the vendor gate folded the jump to the next line and its label are left; gone, no behaviour change.
	// CHANGES: POR script, area 0 shop and temple greetings.
	{Area: 0, Old: `    if FoughtWatch > 0 { goto RefuseService }
    ShowPortrait(BlackBeard, ShopkeeperApron)
    Text("THE SHOPKEEPER GREETS YOU AND OFFERS YOU ITEMS FOR SALE. 'CAN I SHOW YOU SOMETHING?'")
    goto GeneralShopBrowse

    GeneralShopBrowse:
    ClearMonsters()
    OpenShop(HolyWaterVial, SilverHolySymbolOfSune, IronHolySymbolOfTempus, WoodenHolySymbolOfTyr, FlaskOfOil, SilverMirror, BrassMirror)
    goto ResetVisits
`, New: `    if FoughtWatch > 0 { goto RefuseService }
    ShowPortrait(BlackBeard, ShopkeeperApron)
    Text("THE SHOPKEEPER GREETS YOU AND OFFERS YOU ITEMS FOR SALE. 'CAN I SHOW YOU SOMETHING?'")
    ClearMonsters()
    OpenShop(HolyWaterVial, SilverHolySymbolOfSune, IronHolySymbolOfTempus, WoodenHolySymbolOfTyr, FlaskOfOil, SilverMirror, BrassMirror)
    goto ResetVisits
`},
	// 4: area 0 TavernTale. Original: `Text("YOU OVERHEAR TAVERN TALE ")`; `Scratch6 += 1`; `Print(Scratch6)`; `WaitReturnPressed()`.
	// Why: (extra) the tale itself in the pager; the original printed "YOU OVERHEAR TAVERN TALE n" for the player to look up.
	// CHANGES: POR script, area 0 TavernTale.
	{Area: 0, Old: `    goto TavernTale

TavernTale:
    Text("YOU OVERHEAR TAVERN TALE ")
    Scratch6 = Random(22)
    Scratch6 += 1
    Print(Scratch6)
    WaitReturnPressed()
    AuxRoll = Random(3)
    if AuxRoll == 1 { goto DrunkenBrawl }
    exit
`, New: `    goto TavernTale

TavernTale:
    # (extra) the tale itself in the pager; the original printed "YOU OVERHEAR TAVERN TALE n" for the player to look up
    Scratch6 = Random(22)
    Text("") # the text window starts over under the pager, as the tale print did (Journal leaves it)
    goto [Tale1, Tale2, Tale3, Tale4, Tale5, Tale6, Tale7, Tale8, Tale9, Tale10, Tale11, Tale12, Tale13, Tale14, Tale15, Tale16, Tale17, Tale18, Tale19, Tale20, Tale21, Tale22, Tale23][Scratch6]
    exit

    Tale1:
    Journal(TavernTale1)
    goto TaleTold

    Tale2:
    Journal(TavernTale2)
    goto TaleTold

    Tale3:
    Journal(TavernTale3)
    goto TaleTold

    Tale4:
    Journal(TavernTale4)
    goto TaleTold

    Tale5:
    Journal(TavernTale5)
    goto TaleTold

    Tale6:
    Journal(TavernTale6)
    goto TaleTold

    Tale7:
    Journal(TavernTale7)
    goto TaleTold

    Tale8:
    Journal(TavernTale8)
    goto TaleTold

    Tale9:
    Journal(TavernTale9)
    goto TaleTold

    Tale10:
    Journal(TavernTale10)
    goto TaleTold

    Tale11:
    Journal(TavernTale11)
    goto TaleTold

    Tale12:
    Journal(TavernTale12)
    goto TaleTold

    Tale13:
    Journal(TavernTale13)
    goto TaleTold

    Tale14:
    Journal(TavernTale14)
    goto TaleTold

    Tale15:
    Journal(TavernTale15)
    goto TaleTold

    Tale16:
    Journal(TavernTale16)
    goto TaleTold

    Tale17:
    Journal(TavernTale17)
    goto TaleTold

    Tale18:
    Journal(TavernTale18)
    goto TaleTold

    Tale19:
    Journal(TavernTale19)
    goto TaleTold

    Tale20:
    Journal(TavernTale20)
    goto TaleTold

    Tale21:
    Journal(TavernTale21)
    goto TaleTold

    Tale22:
    Journal(TavernTale22)
    goto TaleTold

    Tale23:
    Journal(TavernTale23)
    goto TaleTold

    TaleTold:
    AuxRoll = Random(3)
    if AuxRoll == 1 { goto DrunkenBrawl }
    exit
`},
	// 5: area 0 ArmorShop. Original: `goto ArmorShopBrowse`; ``; `ArmorShopBrowse:`.
	// Why: the shops open their menu directly (46b05b8): with the vendor gate folded the jump to the next line and its label are left; gone, no behaviour change.
	// CHANGES: POR script, area 0 shop and temple greetings.
	{Area: 0, Old: `    ShowPortrait(CurlyBlondYouth, GreyApron3)
    if FoughtWatch > 0 { goto RefuseService }
    Text("THE SHOPKEEPER OFFERS JEWELRY FOR SALE. 'CAN I SHOW YOU SOMETHING?'")
    goto ArmorShopBrowse

    ArmorShopBrowse:
    ClearMonsters()
    OpenShop(BrassDragonFigurine, GoldChain, PearlNecklace, GoldLightningFigurine, JewelledSilverWings, SilverDragonStatuette, GoldScarab, EmeraldBroach, PlatinumSphere, FineOpalPendant, DiamondNecklace)
    goto ResetVisits
`, New: `    ShowPortrait(CurlyBlondYouth, GreyApron3)
    if FoughtWatch > 0 { goto RefuseService }
    Text("THE SHOPKEEPER OFFERS JEWELRY FOR SALE. 'CAN I SHOW YOU SOMETHING?'")
    ClearMonsters()
    OpenShop(BrassDragonFigurine, GoldChain, PearlNecklace, GoldLightningFigurine, JewelledSilverWings, SilverDragonStatuette, GoldScarab, EmeraldBroach, PlatinumSphere, FineOpalPendant, DiamondNecklace)
    goto ResetVisits
`},
	// 6: area 0 WeaponShop. Original: `goto WeaponShopBrowse`; ``; `WeaponShopBrowse:`.
	// Why: the shops open their menu directly (46b05b8): with the vendor gate folded the jump to the next line and its label are left; gone, no behaviour change.
	// CHANGES: POR script, area 0 shop and temple greetings.
	{Area: 0, Old: `    ShowPortrait(BlackBeard, ShopkeeperApron)
    if FoughtWatch > 0 { goto RefuseService }
    Text("THE SHOP SPECIALIZES IN ARMS AND ARMOR. 'CAN I SHOW YOU OUR WARES?'")
    goto WeaponShopBrowse

    WeaponShopBrowse:
    ClearMonsters()
    OpenShop(Shield, PlateMail, BandedMail, SplintMail, ChainMail, ScaleMail, RingMail, StuddedLeatherArmor, PaddedArmor, LeatherArmor, Arrow*10, LightCrossbow, Sling, ShortBow, LongBow, CompositeShortBow, CompositeLongBow, Voulge, Trident, TwoHandedSword, ShortSword, LongSword, BroadSword, BastardSword, QuarterStaff, Spetum, Spear, Scimitar, Ranseur, Quarrel*20, AwlPike, MilitaryPick, Partisan, MorningStar, Mace, JoStick, Javelin*2, Hammer, LucernHammer, Halberd, GuisarmeVoulge, Guisarme, GlaiveGuisarme, Glaive, MilitaryFork, Flail, FauchardFork, Fauchard, Dart*4, Dagger, Club, BoStick, BillGuisarme, BecDeCorbin, Bardiche, HandAxe, BattleAxe)
    goto ResetVisits
`, New: `    ShowPortrait(BlackBeard, ShopkeeperApron)
    if FoughtWatch > 0 { goto RefuseService }
    Text("THE SHOP SPECIALIZES IN ARMS AND ARMOR. 'CAN I SHOW YOU OUR WARES?'")
    ClearMonsters()
    OpenShop(Shield, PlateMail, BandedMail, SplintMail, ChainMail, ScaleMail, RingMail, StuddedLeatherArmor, PaddedArmor, LeatherArmor, Arrow*10, LightCrossbow, Sling, ShortBow, LongBow, CompositeShortBow, CompositeLongBow, Voulge, Trident, TwoHandedSword, ShortSword, LongSword, BroadSword, BastardSword, QuarterStaff, Spetum, Spear, Scimitar, Ranseur, Quarrel*20, AwlPike, MilitaryPick, Partisan, MorningStar, Mace, JoStick, Javelin*2, Hammer, LucernHammer, Halberd, GuisarmeVoulge, Guisarme, GlaiveGuisarme, Glaive, MilitaryFork, Flail, FauchardFork, Fauchard, Dart*4, Dagger, Club, BoStick, BillGuisarme, BecDeCorbin, Bardiche, HandAxe, BattleAxe)
    goto ResetVisits
`},
	// 7: area 0 SilverShop. Original: `goto SilverShopBrowse`; ``; `SilverShopBrowse:`.
	// Why: the shops open their menu directly (46b05b8): with the vendor gate folded the jump to the next line and its label are left; gone, no behaviour change.
	// CHANGES: POR script, area 0 shop and temple greetings.
	{Area: 0, Old: `    ShowPortrait(CurlyBlondYouth, GreyApron3)
    if FoughtWatch > 0 { goto RefuseService }
    Text("THE SHOP SELLS SILVER ITEMS. 'WOULD YOU LIKE TO SEE OUR SELECTION?'")
    goto SilverShopBrowse

    SilverShopBrowse:
    ClearMonsters()
    OpenShop(FineCompositeLongBow, SilverRing, SilverPlateMail, SilverChainMail, SilverQuarrel, SilverArrow*20, SilverDagger, SilverTwoHandedSword, SilverShortSword, SilverLongSword, SilverBroadSword, SilverBastardSword, SilverMace)
    goto ResetVisits
`, New: `    ShowPortrait(CurlyBlondYouth, GreyApron3)
    if FoughtWatch > 0 { goto RefuseService }
    Text("THE SHOP SELLS SILVER ITEMS. 'WOULD YOU LIKE TO SEE OUR SELECTION?'")
    ClearMonsters()
    OpenShop(FineCompositeLongBow, SilverRing, SilverPlateMail, SilverChainMail, SilverQuarrel, SilverArrow*20, SilverDagger, SilverTwoHandedSword, SilverShortSword, SilverLongSword, SilverBroadSword, SilverBastardSword, SilverMace)
    goto ResetVisits
`},
	// 8: area 0 HealerTemple. Original: `goto SeekHealingAsk`; ``; `SeekHealingAsk:`; `Print("' DO YOU SEEK HEALING?'")`.
	// Why: as TempleOfTyr: the SeekHealingAsk print goes, TempleHealing opens directly.
	// CHANGES: POR script, area 0 shop and temple greetings.
	{Area: 0, Old: `    ShowPortrait(NoseguardHelmet, PlateArmor)
    if FoughtWatch > 0 { goto WatchFightsBack }
    Text("THE PRIESTS OF TEMPUS BID YOU WELCOME.")
    goto SeekHealingAsk

SeekHealingAsk:
    Print("' DO YOU SEEK HEALING?'")
    goto TempleHealing

MustLeave:
`, New: `    ShowPortrait(NoseguardHelmet, PlateArmor)
    if FoughtWatch > 0 { goto WatchFightsBack }
    Text("THE PRIESTS OF TEMPUS BID YOU WELCOME.")
    goto TempleHealing

MustLeave:
`},
	// 9: area 0 RefuseService. Original: nothing here.
	// Why: (extra) Rolf, the tour guide, after the tour: the Adventurer's Journal's; history of Phlan, a chapter at a time (rules/history.go)..
	// CHANGES: POR script, area 0 Rolf's history of Phlan.
	{Area: 0, Old: `    Scratch1 = Menu("YES", exit="NO")
    return
}

ResetVisits:
    TrainerVisited = 0
    BuildingEntered = 0
`, New: `    Scratch1 = Menu("YES", exit="NO")
    return
}

# (extra) Rolf, the tour guide, after the tour: the Adventurer's Journal's
# history of Phlan, a chapter at a time (rules/history.go).
RolfEvent:
    ShowPortrait(GrinningBlond, ShopkeeperApron)
    Text("ROLF, THE COUNCIL'S GUIDE, SMILES AS YOU APPROACH. 'WELL MET AGAIN, FRIENDS. WHAT CAN I DO FOR YOU?'")
    Scratch1 = Menu("ASK ABOUT PHLAN", "LEAVE")
    goto [RolfHistory, RolfLeave][Scratch1]
    goto RolfLeave

RolfHistory:
    Scratch1 = VMenu("'PHLAN HAS RISEN AND FALLEN MORE THAN ONCE. WHAT WOULD YOU HEAR ABOUT?'", "THE MOONSEA REACHES", "THE ANCIENT CITY OF PHLAN",
        "GREATHAMMER AND THE FIRST REBIRTH", "THE FALL OF PHLAN", "THE REEMERGENCE OF PHLAN",
        "PHLAN TODAY", exit="LEAVE")
    goto [PhlanHistory1, PhlanHistory2, PhlanHistory3, PhlanHistory4, PhlanHistory5,
        PhlanHistory6, RolfLeave][Scratch1]
    goto RolfLeave

PhlanHistory1:
    Journal(History1)
    goto RolfHistory

PhlanHistory2:
    Journal(History2)
    goto RolfHistory

PhlanHistory3:
    Journal(History3)
    goto RolfHistory

PhlanHistory4:
    Journal(History4)
    goto RolfHistory

PhlanHistory5:
    Journal(History5)
    goto RolfHistory

PhlanHistory6:
    Journal(History6)
    goto RolfHistory

RolfLeave:
    Text("'FAREWELL, AND MAY TYMORA SMILE ON YOU.'")
    Portrait = None
    goto ResetVisits

ResetVisits:
    TrainerVisited = 0
    BuildingEntered = 0
`},
	// 10: area 1 InsideSlavePen. Original: nothing here.
	// Why: fixed: the boy was offered again after he was bought, rescued or lost.
	// CHANGES: POR script, area 1 InsideSlavePen.
	{Area: 1, Old: `
InsideSlavePen:
    if SlavePenBlocked != 0 { goto Done }
    ShowPortrait(None, 42)
    Text("YOU ARE IN THE SLAVE PEN.  STANDING NEAR YOU IS THE YOUNG BOY.  WHAT WILL YOU DO?")
    Scratch1 = Menu("GET HIM", exit="LEAVE")
`, New: `
InsideSlavePen:
    if SlavePenBlocked != 0 { goto Done }
    if QuestBivant != 0 { goto Done } # fixed: the boy was offered again after he was bought, rescued or lost
    ShowPortrait(None, 42)
    Text("YOU ARE IN THE SLAVE PEN.  STANDING NEAR YOU IS THE YOUNG BOY.  WHAT WILL YOU DO?")
    Scratch1 = Menu("GET HIM", exit="LEAVE")
`},
	// 11: area 2 on camp. Original: nothing here.
	// Why: fix: no encounter is left (RandomEncounter exits), the rest was still broken off.
	// CHANGES: POR script, area 2 on camp.
	{Area: 2, Old: `    RestEvery = 12
    RestChance = 6
    if TextileHouseCleared > 253 { RestEvery = 0 }
    exit
}

`, New: `    RestEvery = 12
    RestChance = 6
    if TextileHouseCleared > 253 { RestEvery = 0 }
    if RandomEncounters2 == 9 and IronBoxFound == 0 { RestEvery = 0 } # fix: no encounter is left (RandomEncounter exits), the rest was still broken off
    if Scratch10 == 22 and IronBoxFound == 0 { RestEvery = 12 }
    exit
}

`},
	// 12: area 2 PickEncounter. Original: `if CombatEnd > PartyDied { goto EjectToStreet }`.
	// Why: running away goes to RandomFled (ChasedOut's fix).
	// CHANGES: POR script, area 2 fled encounters.
	{Area: 2, Old: `    ClearMonsters()
    LoadMonster(PoisonousFrog, count=8, icon=PoisonousFrog)
    Combat()
    if CombatEnd > PartyDied { goto EjectToStreet }
    exit

    HobgoblinsCharge:
`, New: `    ClearMonsters()
    LoadMonster(PoisonousFrog, count=8, icon=PoisonousFrog)
    Combat()
    if CombatEnd > PartyDied { goto RandomFled }
    exit

    HobgoblinsCharge:
`},
	// 13: area 2 HobgoblinsCharge. Original: `if CombatEnd > PartyDied { goto EjectToStreet }`.
	// Why: running away goes to RandomFled (ChasedOut's fix).
	// CHANGES: POR script, area 2 fled encounters.
	{Area: 2, Old: `    LoadMonster(Hobgoblin, count=8, icon=Hobgoblin)
    LoadMonster(HobgoblinLdr, count=1, icon=HobgoblinLdr)
    Combat()
    if CombatEnd > PartyDied { goto EjectToStreet }
    exit

    SkeletalForce:
`, New: `    LoadMonster(Hobgoblin, count=8, icon=Hobgoblin)
    LoadMonster(HobgoblinLdr, count=1, icon=HobgoblinLdr)
    Combat()
    if CombatEnd > PartyDied { goto RandomFled }
    exit

    SkeletalForce:
`},
	// 14: area 2 SkeletalForce. Original: `if CombatEnd > PartyDied { goto EjectToStreet }`.
	// Why: running away goes to RandomFled (ChasedOut's fix).
	// CHANGES: POR script, area 2 fled encounters.
	{Area: 2, Old: `    LoadMonster(Skeleton, count=10, icon=Skeleton)
    LoadMonster(Zombie, count=8, icon=Zombie)
    Combat()
    if CombatEnd > PartyDied { goto EjectToStreet }
    exit

    GhoulsAttack:
`, New: `    LoadMonster(Skeleton, count=10, icon=Skeleton)
    LoadMonster(Zombie, count=8, icon=Zombie)
    Combat()
    if CombatEnd > PartyDied { goto RandomFled }
    exit

    GhoulsAttack:
`},
	// 15: area 2 GhoulsAttack. Original: `if CombatEnd > PartyDied { goto EjectToStreet }`.
	// Why: running away goes to RandomFled (ChasedOut's fix).
	// CHANGES: POR script, area 2 fled encounters.
	{Area: 2, Old: `    LoadMonster(Wight, count=1, icon=Wight)
    LoadMonster(Ghoul, count=5, icon=Ghoul)
    Combat()
    if CombatEnd > PartyDied { goto EjectToStreet }
    exit

    ScorpionsApproach:
`, New: `    LoadMonster(Wight, count=1, icon=Wight)
    LoadMonster(Ghoul, count=5, icon=Ghoul)
    Combat()
    if CombatEnd > PartyDied { goto RandomFled }
    exit

    ScorpionsApproach:
`},
	// 16: area 2 ScorpionsApproach. Original: `if CombatEnd > PartyDied { goto EjectToStreet }`.
	// Why: running away goes to RandomFled (ChasedOut's fix).
	// CHANGES: POR script, area 2 fled encounters.
	{Area: 2, Old: `    LoadMonster(LargeScorpion, count=3, icon=Scorpion)
    LoadMonster(HugeScorpion, count=1, icon=Scorpion)
    Combat()
    if CombatEnd > PartyDied { goto EjectToStreet }
    exit

CourtyardAmbush:
`, New: `    LoadMonster(LargeScorpion, count=3, icon=Scorpion)
    LoadMonster(HugeScorpion, count=1, icon=Scorpion)
    Combat()
    if CombatEnd > PartyDied { goto RandomFled }
    exit

CourtyardAmbush:
`},
	// 17: area 2 ChasedOut. Original: nothing here.
	// Why: fix: running away still counted the encounter (ten clear the house).
	// CHANGES: POR script, area 2 fled encounters.
	{Area: 2, Old: `
    ChasedOut:
    Text("YOU ARE CHASED OUT INTO THE STREETS.")
    goto EjectToStreet

EjectToStreet:
`, New: `
    ChasedOut:
    Text("YOU ARE CHASED OUT INTO THE STREETS.")
    goto EjectToStreet

RandomFled: # fix: running away still counted the encounter (ten clear the house)
    if Scratch10 != 22 or IronBoxFound != 0 { RandomEncounters2 -= 1 }
    goto EjectToStreet

EjectToStreet:
`},
	// 18: area 2 MeetRestal. Original: nothing here.
	// Why: fix: once, not every visit.
	// CHANGES: POR script, area 2 Restal's seal once.
	{Area: 2, Old: `    if IronBoxState == 0 { goto TextileHouseOffer }
    if QuestCadornaPayment == Rewarded { goto RestalFarewell }
    if RestalDealDone == 1 { goto RestalFarewell }
    Text("'WE CAN DUPLICATE THE SEAL ON THAT CHEST -- FOR HALF THE TREASURE -- NO ONE WILL EVER KNOW. INTERESTED?'")
    Scratch1 = Menu("YES", exit="NO")
    goto [DuplicateSeal, RestalFarewell][Scratch1]
`, New: `    if IronBoxState == 0 { goto TextileHouseOffer }
    if QuestCadornaPayment == Rewarded { goto RestalFarewell }
    if RestalDealDone == 1 { goto RestalFarewell }
    if RestalSealDuplicated == 1 { goto RestalFarewell } # fix: once, not every visit
    Text("'WE CAN DUPLICATE THE SEAL ON THAT CHEST -- FOR HALF THE TREASURE -- NO ONE WILL EVER KNOW. INTERESTED?'")
    Scratch1 = Menu("YES", exit="NO")
    goto [DuplicateSeal, RestalFarewell][Scratch1]
`},
	// 19: area 2 DuplicateSeal. Original: nothing here.
	// Why: marks the duplication done (MeetRestal's fix).
	// CHANGES: POR script, area 2 Restal's seal once.
	{Area: 2, Old: `    Treasure(platinum=200, gems=4, jewelry=2, items=[GauntletsOfOgrePower, GoldBandedWand, JewelStuddedBowl, JewelledDragonStatuette, Ring])
    Combat()
    RestalDealDone = 1
    IronBoxState = CarriedUnopened
    goto RestalFarewell

`, New: `    Treasure(platinum=200, gems=4, jewelry=2, items=[GauntletsOfOgrePower, GoldBandedWand, JewelStuddedBowl, JewelledDragonStatuette, Ring])
    Combat()
    RestalDealDone = 1
    RestalSealDuplicated = 1
    IronBoxState = CarriedUnopened
    goto RestalFarewell

`},
	// 20: area 5 ThroneRoom. Original: `if Scratch10 >= 85 { goto ArrogantKnave }`.
	// Why: fixed: was >= 85, so a weak abusive party got the confession and a strong one the fight.
	// CHANGES: POR script, area 5 ThroneRoom.
	{Area: 5, Old: `    goto [ArrogantKnave, HeavyCrown, BegoneOrWrath][Scratch10]
    Scratch10 = PartyStrength()
    Scratch10 += SelReaction
    if Scratch10 >= 85 { goto ArrogantKnave }
    Text("HE SEEMS TO COLLAPSE. 'ALL RIGHT. I'M NOT TYRANTHRAXUS. HE JUST HIRED ME TO IMPERSONATE HIM. IT'S NOT A BAD LIFE, BUT, HEY! I CAN TELL WHICH WAY THE WIND'S BLOWING.'")
    WaitReturnPressed()
    Text("'JUST LET ME GO, OKAY?")
`, New: `    goto [ArrogantKnave, HeavyCrown, BegoneOrWrath][Scratch10]
    Scratch10 = PartyStrength()
    Scratch10 += SelReaction
    if Scratch10 < 85 { goto ArrogantKnave } # fixed: was >= 85, so a weak abusive party got the confession and a strong one the fight
    Text("HE SEEMS TO COLLAPSE. 'ALL RIGHT. I'M NOT TYRANTHRAXUS. HE JUST HIRED ME TO IMPERSONATE HIM. IT'S NOT A BAD LIFE, BUT, HEY! I CAN TELL WHICH WAY THE WIND'S BLOWING.'")
    WaitReturnPressed()
    Text("'JUST LET ME GO, OKAY?")
`},
	// 21: area 7 VoteNext. Original: `TowerEndFlags = GenheerisFlags | DragonDone`.
	// Why: fixed: the original stored GenheerisFlags|DragonDone in TowerEndFlags (4A72, never read), so the dragon fight and the ending could be replayed.
	// CHANGES: POR script, area 7 VoteNext.
	{Area: 7, Old: `    Combat()
    if CombatEnd == PartyFled { goto RetreatFromHall }
    if QuestTyranthraxus != Rewarded { QuestTyranthraxus = RewardDue }
    TowerEndFlags = GenheerisFlags | DragonDone
    TyranthraxusDefeated = 1
    Program(8)
    goto LeaveCastle
`, New: `    Combat()
    if CombatEnd == PartyFled { goto RetreatFromHall }
    if QuestTyranthraxus != Rewarded { QuestTyranthraxus = RewardDue }
    GenheerisFlags |= DragonDone # fixed: the original stored GenheerisFlags|DragonDone in TowerEndFlags (4A72, never read), so the dragon fight and the ending could be replayed
    TyranthraxusDefeated = 1
    Program(8)
    goto LeaveCastle
`},
	// 22a: area 9 CheckMemberMoney. Original: the member's five coins summed in platinum (copper/10 + silver, /10 + electrum, /2 + gold, /5 + platinum).
	// Why: gold and copper only (Felix 2026-11-12, script.GoldCoins): the same sum, in platinum, of the two coins there are.
	// CHANGES: POR script, area 9 coins in gold.
	{Area: 9, Old: `    &Scratch6 = &SelCoins / 10
    &Scratch6 += &SelCoins # coins: rewritten to gold/copper; operands not converted by value
    &Scratch6 = &Scratch6 / 10
    &Scratch6 += &SelCoins # coins: rewritten to gold/copper; operands not converted by value
    &Scratch6 = &Scratch6 / 2
    &Scratch6 += &SelCoins_6
    &Scratch6 = &Scratch6 / 5
    &Scratch6 += &SelCoins_6 # coins: rewritten to gold/copper; operands not converted by value
`, New: `    &Scratch6 = &SelCoins / 200
    &Scratch6 += &SelCoins_6
    &Scratch6 = &Scratch6 / 5 # gold and copper only (Felix 2026-11-12): the member's money in platinum, as the original's sum of its five coins
`},
	// 22: area 9 NextPayer. Original: `TakeCoins:`; `Scratch10 = 195`; `Scratch10 = 193`; `Scratch10 = 191`; `Scratch10 = 189`; `Scratch10 = 187`.
	// Why: typed: the coins are SelCoins[Scratch10] (gold 6, copper 0); was SelName[187 + Scratch10] with its high byte in V_6B01, Scratch7. Gold and copper only (Felix 2026-11-12): the platinum cost scaled to gold (x5), then copper (x200).
	// CHANGES: POR script, area 9 coins typed.
	{Area: 9, Old: `    WaitReturnPressed()
    return 0 != 1

    TakeCoins:
    Scratch10 = 195
    if TakeCoinKind() { return }
    Scratch2 = 4
    ScaleCost()
    Scratch10 = 193
    if TakeCoinKind() { return }
    Scratch2 = 1
    ScaleCost()
    Scratch10 = 191
    if TakeCoinKind() { return }
    Scratch2 = 9
    ScaleCost()
    Scratch10 = 189
    if TakeCoinKind() { return }
    Scratch2 = 9
    ScaleCost()
    Scratch10 = 187
    TakeCoinKind()
    return 0 != 0

`, New: `    WaitReturnPressed()
    return 0 != 1

    TakeCoins: # typed: the coins are SelCoins[Scratch10] (gold 6, copper 0); was SelName[187 + Scratch10] with its high byte in V_6B01, Scratch7; gold and copper only (Felix 2026-11-12): the cost, in platinum, scaled to gold, then copper
    Scratch2 = 4
    ScaleCost()
    Scratch10 = 6
    if TakeCoinKind() { return }
    Scratch2 = 199
    ScaleCost()
    Scratch10 = 0
    TakeCoinKind()
    return 0 != 0

`},
	// 23: area 9 TakeFromMember. Original: `Scratch6 = SelName[Scratch10]`; `Scratch7 = V_6B01[Scratch10]`.
	// Why: typed, as NextPayer.
	// CHANGES: POR script, area 9 coins typed.
	{Area: 9, Old: `TakeFromMember:
    SelectMember(Scratch1)
    if CheckMemberOk() { goto NextCoinMember }
    Scratch6 = SelName[Scratch10]
    Scratch7 = V_6B01[Scratch10]
    &Scratch8 = &Scratch6
    if &Scratch8 < &Scratch4 { goto DeductCoins }
    &Scratch8 = &Scratch4
`, New: `TakeFromMember:
    SelectMember(Scratch1)
    if CheckMemberOk() { goto NextCoinMember }
    Scratch6 = SelCoins[Scratch10]
    &Scratch8 = &Scratch6
    if &Scratch8 < &Scratch4 { goto DeductCoins }
    &Scratch8 = &Scratch4
`},
	// 24: area 9 DeductCoins. Original: `SelName[Scratch10] = Scratch6`; `V_6B01[Scratch10] = Scratch7`.
	// Why: typed, as NextPayer.
	// CHANGES: POR script, area 9 coins typed.
	{Area: 9, Old: `
    DeductCoins:
    &Scratch6 -= &Scratch8
    SelName[Scratch10] = Scratch6
    V_6B01[Scratch10] = Scratch7
    &Scratch4 -= &Scratch8
    Scratch2 = Scratch1 | 128
    SelectMember(Scratch2)
`, New: `
    DeductCoins:
    &Scratch6 -= &Scratch8
    SelCoins[Scratch10] = Scratch6
    &Scratch4 -= &Scratch8
    Scratch2 = Scratch1 | 128
    SelectMember(Scratch2)
`},
	// 25: area 9 AlarmActive. Original: `AlarmEndHour = 1 + AuxCounter`.
	// Why: fixed: was 1 + AuxCounter (a loop temp); the castle areas use AlarmHour.
	// CHANGES: POR script, area 9 AlarmActive.
	{Area: 9, Old: `    if AlarmEndHour == 0 { Text("AN ALARM STARTS RINGING!") }
    AlarmHour = Hours
    AlarmMinute = TenMinutes
    AlarmEndHour = 1 + AuxCounter
    Delay()
    return
}
`, New: `    if AlarmEndHour == 0 { Text("AN ALARM STARTS RINGING!") }
    AlarmHour = Hours
    AlarmMinute = TenMinutes
    AlarmEndHour = 1 + AlarmHour # fixed: was 1 + AuxCounter (a loop temp); the castle areas use AlarmHour
    Delay()
    return
}
`},
	// 26: area 9 GateTowerSquare. Original: `if Scratch9 = GateFlags & WestGarrisonSlain; Scratch9 == 0 { Scratch2 |= 2 }`; `if GateFlags lacks EastGarrisonSlain { Scratch2 |= 1 }`; `goto TowersCharge`; ``; `EastTowerCheck:`.
	// Why: fixed: the two sides tested the wrong tower as the near one (cell 11 is below the east tower, TowerExitCellTable), so a slain garrison charged again; cell 10, below the west tower.
	// CHANGES: POR script, area 9 GateTowerSquare.
	{Area: 9, Old: `    Scratch2 = 0
    Scratch1 = Cell & 1
    goto [EastTowerCheck][Scratch1]
    if Scratch9 = GateFlags & WestGarrisonSlain; Scratch9 == 0 { Scratch2 |= 2 }
    if GateFlags lacks EastGarrisonSlain { Scratch2 |= 1 }
    goto TowersCharge

EastTowerCheck:
    if Scratch9 = GateFlags & EastGarrisonSlain; Scratch9 == 0 { Scratch2 |= 2 }
    if GateFlags lacks WestGarrisonSlain { Scratch2 |= 1 }
    goto TowersCharge

TowersCharge:
`, New: `    Scratch2 = 0
    Scratch1 = Cell & 1
    goto [EastTowerCheck][Scratch1]
    # fixed: the two sides tested the wrong tower as the near one (cell 11 is below the east tower, TowerExitCellTable), so a slain garrison charged again
    if Scratch9 = GateFlags & EastGarrisonSlain; Scratch9 == 0 { Scratch2 |= 2 }
    if GateFlags lacks WestGarrisonSlain { Scratch2 |= 1 }
    goto TowersCharge

EastTowerCheck: # cell 10, below the west tower
    if Scratch9 = GateFlags & WestGarrisonSlain; Scratch9 == 0 { Scratch2 |= 2 }
    if GateFlags lacks EastGarrisonSlain { Scratch2 |= 1 }
    goto TowersCharge

TowersCharge:
`},
	// 27: area 9 FleeToArea6. Original: `NewECL(ValjevoCastle6)`.
	// Why: fix: the original left SymbolSet at 2, loading area 6 from the wrong data set.
	// CHANGES: POR script, area 9 FleeToArea6.
	{Area: 9, Old: `    PartyX = 4
    PartyY = 15
    Facing = North
    NewECL(ValjevoCastle6)

FightTowerMonsters:
    PartySurprise(Scratch9, Scratch10)
`, New: `    PartyX = 4
    PartyY = 15
    Facing = North
    EnterArea(ValjevoCastle6) # fix: the original left SymbolSet at 2, loading area 6 from the wrong data set

FightTowerMonsters:
    PartySurprise(Scratch9, Scratch10)
`},
	// 28: area 9 BugbearsSendAway. Original: `BugbearWarnings = Suspicious`.
	// Why: fixed: BugbearWarnings = Suspicious here made every later visit an impostor accusation; it belongs to ImpostorsAccused.
	// CHANGES: POR script, area 9 bugbear bluff.
	{Area: 9, Old: `
BugbearsSendAway:
    Text("'ALL RIGHT, GET OUT OF HERE THEN. COME BACK WHEN YOU'VE LEARNED THE RULES.'")
    BugbearWarnings = Suspicious
    goto Done

PayToll:
`, New: `
BugbearsSendAway:
    Text("'ALL RIGHT, GET OUT OF HERE THEN. COME BACK WHEN YOU'VE LEARNED THE RULES.'")
    # fixed: BugbearWarnings = Suspicious here made every later visit an impostor accusation; it belongs to ImpostorsAccused
    goto Done

PayToll:
`},
	// 29: area 9 ImpostorsAccused. Original: nothing here.
	// Why: fixed: a party that fled or surrendered could bluff again.
	// CHANGES: POR script, area 9 bugbear bluff.
	{Area: 9, Old: `
ImpostorsAccused:
    Text("'HEY, WAIT A MINUTE. THEY ARE IMPOSTERS! ALL RIGHT, THROW DOWN YOUR WEAPONS!'")
    goto BugbearsMenu

WallSquare:
`, New: `
ImpostorsAccused:
    Text("'HEY, WAIT A MINUTE. THEY ARE IMPOSTERS! ALL RIGHT, THROW DOWN YOUR WEAPONS!'")
    BugbearWarnings = Suspicious # fixed: a party that fled or surrendered could bluff again
    goto BugbearsMenu

WallSquare:
`},
	// 30: area 14 OpenLockMenu. Original: `CastKnock:`; `Spell(31, Scratch5, Scratch6)`; `if Scratch5 != 255 { goto KnockOpens }`; `Text("NO ONE IN THE PARTY HAS THE KNOCK SPELL!  HOW WILL YOU OPEN THE LOCK?")`; `goto OpenLockMenu`; `SelMemorized[Scratch5] = 0`.
	// Why: engine (castKnock): the first okay member with Knock memorized casts it, the slot spent, and goes on at KnockOpens with Scratch6 its place in the party; with none at NoKnock. Was Spell(31, Scratch5, Scratch6) and the script clearing the slot.
	// CHANGES: Cabinets' CAST KNOCK (area 14).
	{Area: 14, Old: `    goto [PickLock, BreakLock, CastKnock, LeaveLock][Scratch1]
    goto CastKnock

CastKnock:
    Spell(31, Scratch5, Scratch6)
    if Scratch5 != 255 { goto KnockOpens }
    Text("NO ONE IN THE PARTY HAS THE KNOCK SPELL!  HOW WILL YOU OPEN THE LOCK?")
    goto OpenLockMenu

    KnockOpens:
    SelectMember(Scratch6)
    SelMemorized[Scratch5] = 0
    Text($SelName)
    Print(" CASTS A KNOCK SPELL AND OPENS THE LOCK.")
    Scratch6 += 128
`, New: `    goto [PickLock, BreakLock, CastKnock, LeaveLock][Scratch1]
    goto CastKnock

CastKnock: # engine (castKnock): the first okay member with Knock memorized casts it, the slot spent, and goes on at KnockOpens with Scratch6 its place in the party; with none at NoKnock. Was Spell(31, Scratch5, Scratch6) and the script clearing the slot
    goto NoKnock

    KnockOpens:
    SelectMember(Scratch6)
    Text($SelName)
    Print(" CASTS A KNOCK SPELL AND OPENS THE LOCK.")
    Scratch6 += 128
`},
	// 31: area 14 KnockOpens. Original: nothing here.
	// Why: CastKnock's NoKnock section (OpenLockMenu's).
	// CHANGES: Cabinets' CAST KNOCK (area 14).
	{Area: 14, Old: `    LockTable[Scratch2] = 0
    return

PickLock:
    Scratch3 = LockTable[Scratch2]
    if Scratch3 == 2 { goto LockTooComplex }
`, New: `    LockTable[Scratch2] = 0
    return

NoKnock:
    Text("NO ONE IN THE PARTY HAS THE KNOCK SPELL!  HOW WILL YOU OPEN THE LOCK?")
    goto OpenLockMenu

PickLock:
    Scratch3 = LockTable[Scratch2]
    if Scratch3 == 2 { goto LockTooComplex }
`},
	// 32: area 14 WeaponPileAmbush. Original: nothing here.
	// Why: fix: 13-15 killed made 5 - 6..8 thieves, a byte gone below 0 (253-255).
	// CHANGES: POR script, area 14 WeaponPileAmbush.
	{Area: 14, Old: `    Scratch1 = 5 + ThievesKilled
    if Scratch1 >= 21 { goto AmbushManyMasters }
    if Scratch1 <= 12 { goto AmbushFewThieves }
    Scratch2 = Scratch1 - 12
    Scratch3 = 5 - Scratch2
    goto AmbushFight
`, New: `    Scratch1 = 5 + ThievesKilled
    if Scratch1 >= 21 { goto AmbushManyMasters }
    if Scratch1 <= 12 { goto AmbushFewThieves }
    if Scratch1 > 17 { # fix: 13-15 killed made 5 - 6..8 thieves, a byte gone below 0 (253-255)
        Scratch1 = 21
        goto AmbushManyMasters
    }
    Scratch2 = Scratch1 - 12
    Scratch3 = 5 - Scratch2
    goto AmbushFight
`},
	// 33: area 14 AmbushManyMasters. Original: nothing here.
	// Why: fix: the original left the thief count from earlier.
	// CHANGES: POR script, area 14 AmbushManyMasters.
	{Area: 14, Old: `    AmbushManyMasters:
    Scratch2 = Scratch1 - 21
    Scratch2 = 5 - Scratch2
    goto AmbushFight

    AmbushFewThieves:
`, New: `    AmbushManyMasters:
    Scratch2 = Scratch1 - 21
    Scratch2 = 5 - Scratch2
    Scratch3 = 0 # fix: the original left the thief count from earlier
    goto AmbushFight

    AmbushFewThieves:
`},
	// 34: area 16 FindPoolCellLoop. Original: `if GiantLizardsKilled < 10 { Scratch8 = 10 - PoolTreasureTaken }`.
	// Why: fixed: the original subtracted PoolTreasureTaken.
	// CHANGES: POR script, area 16 FindPoolCellLoop.
	{Area: 16, Old: `    Scratch5 = Table9A09[Scratch7]
    if Scratch5 > Scratch8 { Scratch5 = Scratch8 }
    Scratch8 = 0
    if GiantLizardsKilled < 10 { Scratch8 = 10 - PoolTreasureTaken }
    Scratch4 = Table9A0E[Scratch7]
    if Scratch4 > Scratch8 { Scratch4 = Scratch8 }
    return
`, New: `    Scratch5 = Table9A09[Scratch7]
    if Scratch5 > Scratch8 { Scratch5 = Scratch8 }
    Scratch8 = 0
    if GiantLizardsKilled < 10 { Scratch8 = 10 - GiantLizardsKilled } # fixed: the original subtracted PoolTreasureTaken
    Scratch4 = Table9A0E[Scratch7]
    if Scratch4 > Scratch8 { Scratch4 = Scratch8 }
    return
`},
	// 35: area 16 YesNoMenu. Original: `Scratch9 = Menu("PRESS <RETURN> OR BUTTON TO CONTINUE")`.
	// Why: fixed: was Scratch9, which holds the building's fight (BuildingEncounter) across a wrong friend-word.
	// CHANGES: POR script, area 16 YesNoMenu.
	{Area: 16, Old: `    return Scratch10

def PressToContinue() {
    Scratch9 = Menu("PRESS <RETURN> OR BUTTON TO CONTINUE")
    Text(" ")
    return
}
`, New: `    return Scratch10

def PressToContinue() {
    Scratch10 = Menu("PRESS <RETURN> OR BUTTON TO CONTINUE") # fixed: was Scratch9, which holds the building's fight (BuildingEncounter) across a wrong friend-word
    Text(" ")
    return
}
`},
	// 36: area 18 on camp. Original: `if PatrolsBeaten18 >= 64 { RestSafe() }`.
	// Why: fix: the original ran on and set RestEvery 2, RestChance 16 over these.
	// CHANGES: POR script, area 18 on camp.
	{Area: 18, Old: `version 1 2
on camp {
    if PatrolsBeaten18 >= 64 { RestSafe() }
    RestEvery = 2
    RestChance = 16
    if Cell == 132 { RestEvery = 0 }
`, New: `version 1 2
on camp {
    if PatrolsBeaten18 >= 64 { # fix: the original ran on and set RestEvery 2, RestChance 16 over these
        RestSafe()
        exit
    }
    RestEvery = 2
    RestChance = 16
    if Cell == 132 { RestEvery = 0 }
`},
	// 37: area 18 PatrolFight. Original: `CountPatrolBeaten()`.
	// Why: fix: running away counted as beating the patrol.
	// CHANGES: POR script, area 18 PatrolFight.
	{Area: 18, Old: `    LoadPatrol()
    Morale = EncMorale
    Combat()
    CountPatrolBeaten()
    if Cell < 127 { goto CoverBlown }
    if CombatEnd > PartyDied { goto PartyFlees }
    goto Done
`, New: `    LoadPatrol()
    Morale = EncMorale
    Combat()
    if CombatEnd != PartyFled { CountPatrolBeaten() } # fix: running away counted as beating the patrol
    if Cell < 127 { goto CoverBlown }
    if CombatEnd > PartyDied { goto PartyFlees }
    goto Done
`},
	// 38: area 18 FreeForAllMember. Original: nothing here.
	// Why: fix: with a one-member party 8+Random(44) can pass the last of the 50 patrons and leave the member selected.
	// CHANGES: POR script, area 18 FreeForAllMember.
	{Area: 18, Old: `    Scratch1 = Random(44)
    Scratch1 += 8
    Scratch2 += 1
    SelectMember(Scratch1)
    SelSide = PartyControlled
    SelControl = NPCMember
    Scratch1 = 128 | Scratch1
    SelectMember(Scratch1)
    if Scratch2 < 21 { goto FreeForAllMember }
    Combat()
    if CombatEnd > PartyDied { goto PartyFlees }
`, New: `    Scratch1 = Random(44)
    Scratch1 += 8
    Scratch2 += 1
    Scratch3 = PartyCount # fix: with a one-member party 8+Random(44) can pass the last of the 50 patrons and leave the member selected
    Scratch3 += 50
    if Scratch1 >= Scratch3 { goto FreeForAllNext }
    SelectMember(Scratch1)
    SelSide = PartyControlled
    SelControl = NPCMember
    Scratch1 = 128 | Scratch1
    SelectMember(Scratch1)
    FreeForAllNext:
    if Scratch2 < 21 { goto FreeForAllMember }
    Combat()
    if CombatEnd > PartyDied { goto PartyFlees }
`},
	// 39: area 19 on camp. Original: `SelMemorized_18 = 1`.
	// Why: remapped: the original writes 6B32 (memorized slot 19 of the selected member), the decompiler reads it as RestChance (6DD3), the camp's 1% rest-interrupt chance (d4).
	// CHANGES: Diogenes' camp (area 19).
	{Area: 19, Old: `
on camp {
    RestEvery = 0
    SelMemorized_18 = 1
    exit
}

`, New: `
on camp {
    RestEvery = 0
    RestChance = 1 # remapped: the original writes 6B32 (memorized slot 19 of the selected member), the decompiler reads it as RestChance (6DD3), the camp's 1% rest-interrupt chance (d4)
    exit
}

`},
	// 40: area 20 OrcsArguePapers. Original: nothing here.
	// Why: fix: running away counted as beating it.
	// CHANGES: POR script, area 20 fled fights.
	{Area: 20, Old: `    LoadMonster(Orc13, count=1, icon=Orc)
    LoadMonster(Orc4, count=3, icon=Orc)
    Combat()
    OrcPapersFight = Yes
    CountSlumsCleared()
    exit
`, New: `    LoadMonster(Orc13, count=1, icon=Orc)
    LoadMonster(Orc4, count=3, icon=Orc)
    Combat()
    if CombatEnd == PartyFled { exit } # fix: running away counted as beating it
    OrcPapersFight = Yes
    CountSlumsCleared()
    exit
`},
	// 41: area 20 GoblinTrainer. Original: nothing here.
	// Why: fix: running away counted as beating it.
	// CHANGES: POR script, area 20 fled fights.
	{Area: 20, Old: `    LoadMonster(GoblinGuard, count=8, icon=GoblinGuard)
    LoadMonster(GoblinLeader12, count=1, icon=GoblinLeader)
    Combat()
    CountSlumsCleared()
    GoblinTrainerFight = Yes
    exit
`, New: `    LoadMonster(GoblinGuard, count=8, icon=GoblinGuard)
    LoadMonster(GoblinLeader12, count=1, icon=GoblinLeader)
    Combat()
    if CombatEnd == PartyFled { exit } # fix: running away counted as beating it
    CountSlumsCleared()
    GoblinTrainerFight = Yes
    exit
`},
	// 42: area 20 OhloAttacks. Original: nothing here.
	// Why: fix: running away counted as beating it.
	// CHANGES: POR script, area 20 fled fights.
	{Area: 20, Old: `    LoadMonster(Level3Mu, count=1, icon=Level3Mu)
    LoadMonster(Orc4, count=8, icon=Orc)
    Combat()
    ClearPortrait()
    OhloPotionState = OhloFought
    CountSlumsCleared()
`, New: `    LoadMonster(Level3Mu, count=1, icon=Level3Mu)
    LoadMonster(Orc4, count=8, icon=Orc)
    Combat()
    if CombatEnd == PartyFled { exit } # fix: running away counted as beating it
    ClearPortrait()
    OhloPotionState = OhloFought
    CountSlumsCleared()
`},
	// 43: area 20 GivePotion. Original: nothing here.
	// Why: fix: running away counted as beating it.
	// CHANGES: POR script, area 20 fled fights.
	{Area: 20, Old: `    ClearMonsters()
    Treasure(platinum=150, jewelry=1, items=Random(1))
    Combat()
    OhloQuestState = Finished
    OhloPotionState = OhloFought
    ClearPortrait()
`, New: `    ClearMonsters()
    Treasure(platinum=150, jewelry=1, items=Random(1))
    Combat()
    if CombatEnd == PartyFled { goto LeaveOhlo } # fix: running away counted as beating it
    OhloQuestState = Finished
    OhloPotionState = OhloFought
    ClearPortrait()
`},
	// 44: area 20 FeedingKobolds. Original: nothing here.
	// Why: fix: running away counted as beating it.
	// CHANGES: POR script, area 20 fled fights.
	{Area: 20, Old: `    LoadMonster(KoboldLeader11, count=1, icon=KoboldLeader)
    LoadMonster(Kobold0, count=15, icon=Kobold)
    Combat()
    CountSlumsCleared()
    KoboldsFed = Yes
    exit
`, New: `    LoadMonster(KoboldLeader11, count=1, icon=KoboldLeader)
    LoadMonster(Kobold0, count=15, icon=Kobold)
    Combat()
    if CombatEnd == PartyFled { exit } # fix: running away counted as beating it
    CountSlumsCleared()
    KoboldsFed = Yes
    exit
`},
	// 45: area 20 OrcHome. Original: nothing here.
	// Why: fix: running away counted as beating it.
	// CHANGES: POR script, area 20 fled fights.
	{Area: 20, Old: `    LoadMonster(OrcLeader14, count=3, icon=OrcLeader)
    LoadMonster(Orc4, count=20, icon=Orc)
    Combat()
    CountSlumsCleared()
    OrcHomeFight = Yes
    exit
`, New: `    LoadMonster(OrcLeader14, count=3, icon=OrcLeader)
    LoadMonster(Orc4, count=20, icon=Orc)
    Combat()
    if CombatEnd == PartyFled { exit } # fix: running away counted as beating it
    CountSlumsCleared()
    OrcHomeFight = Yes
    exit
`},
	// 46: area 20 HobgoblinGold. Original: nothing here.
	// Why: fix: running away counted as beating it.
	// CHANGES: POR script, area 20 fled fights.
	{Area: 20, Old: `    ClearMonsters()
    LoadMonster(Hobgoblin, count=5, icon=Hobgoblin)
    Combat()
    HobgoblinGoldState = HobgoblinsBeaten
    CountSlumsCleared()
    exit
`, New: `    ClearMonsters()
    LoadMonster(Hobgoblin, count=5, icon=Hobgoblin)
    Combat()
    if CombatEnd == PartyFled { exit } # fix: running away counted as beating it
    HobgoblinGoldState = HobgoblinsBeaten
    CountSlumsCleared()
    exit
`},
	// 47: area 20 GuardsIntercept. Original: nothing here.
	// Why: fix: running away counted as beating it.
	// CHANGES: POR script, area 20 fled fights.
	{Area: 20, Old: `    LoadMonster(Orc4, count=30, icon=Orc)
    LoadMonster(OrcLeader5, count=4, icon=OrcLeader)
    Combat()
    CountSlumsCleared()
    GuardsInterceptFight = Yes
    exit
`, New: `    LoadMonster(Orc4, count=30, icon=Orc)
    LoadMonster(OrcLeader5, count=4, icon=OrcLeader)
    Combat()
    if CombatEnd == PartyFled { exit } # fix: running away counted as beating it
    CountSlumsCleared()
    GuardsInterceptFight = Yes
    exit
`},
	// 48: area 20 LeadersMeeting. Original: nothing here.
	// Why: fix: running away counted as beating it.
	// CHANGES: POR script, area 20 fled fights.
	{Area: 20, Old: `    LoadMonster(Gnoll, count=2, icon=Gnoll)
    LoadMonster(HobgoblinLdr, count=2, icon=HobgoblinLdr)
    Combat()
    LeadersMeetingFight = Yes
    CountSlumsCleared()
    exit
`, New: `    LoadMonster(Gnoll, count=2, icon=Gnoll)
    LoadMonster(HobgoblinLdr, count=2, icon=HobgoblinLdr)
    Combat()
    if CombatEnd == PartyFled { exit } # fix: running away counted as beating it
    LeadersMeetingFight = Yes
    CountSlumsCleared()
    exit
`},
	// 49: area 20 AlertedGuards. Original: nothing here.
	// Why: fix: running away counted as beating it.
	// CHANGES: POR script, area 20 fled fights.
	{Area: 20, Old: `    LoadMonster(GoblinGuard, count=12, icon=GoblinGuard)
    LoadMonster(GoblinLeader3, count=21, icon=GoblinLeader)
    Combat()
    CountSlumsCleared()
    AlertedGuardsFight = Yes
    exit
`, New: `    LoadMonster(GoblinGuard, count=12, icon=GoblinGuard)
    LoadMonster(GoblinLeader3, count=21, icon=GoblinLeader)
    Combat()
    if CombatEnd == PartyFled { exit } # fix: running away counted as beating it
    CountSlumsCleared()
    AlertedGuardsFight = Yes
    exit
`},
	// 50: area 20 AttackBoothMan. Original: nothing here.
	// Why: fix: running away counted as beating it.
	// CHANGES: POR script, area 20 fled fights.
	{Area: 20, Old: `    LoadMonster(Hobgoblin, count=15, icon=Hobgoblin)
    LoadMonster(HobgoblinLdr, count=4, icon=HobgoblinLdr)
    Combat()
    BoothFightDone = 1
    if OhloPotionState != OhloFought { CountSlumsCleared() }
    exit
`, New: `    LoadMonster(Hobgoblin, count=15, icon=Hobgoblin)
    LoadMonster(HobgoblinLdr, count=4, icon=HobgoblinLdr)
    Combat()
    if CombatEnd == PartyFled { exit } # fix: running away counted as beating it
    BoothFightDone = 1
    if OhloPotionState != OhloFought { CountSlumsCleared() }
    exit
`},
	// 51: area 20 GrainSack. Original: nothing here.
	// Why: fix: running away counted as beating it; fix: once the 15 patrols are beaten, rest fights no longer count towards clearing the Slums.
	// CHANGES: POR script, area 20 fled fights.
	{Area: 20, Old: `    LoadMonster(Troll, count=4, icon=Troll)
    Treasure(gold=30, platinum=50, jewelry=1, items=Random(3))
    Combat()
    CountSlumsCleared()
    GrainSackFight = Yes
    exit

AfterPatrolFight:
    if CombatEnd > PartyDied { goto RunAway }
    if CombatEnd < 1 {
        PatrolsBeaten20 += 1
        CountSlumsCleared()
`, New: `    LoadMonster(Troll, count=4, icon=Troll)
    Treasure(gold=30, platinum=50, jewelry=1, items=Random(3))
    Combat()
    if CombatEnd == PartyFled { exit } # fix: running away counted as beating it
    CountSlumsCleared()
    GrainSackFight = Yes
    exit

AfterPatrolFight:
    if CombatEnd > PartyDied { goto RunAway }
    if PatrolsBeaten20 >= 15 { goto Done } # fix: once the 15 patrols are beaten, rest fights no longer count towards clearing the Slums
    if CombatEnd < 1 {
        PatrolsBeaten20 += 1
        CountSlumsCleared()
`},
	// 52: area 24 on move. Original: `Scratch1 = Table99A8[Scratch10]`.
	// Why: fix: no area that way (the south openings): the party bumps instead of loading map 255.
	// CHANGES: POR script, area 24 on move.
	{Area: 24, Old: `    if LeftMap == 0 { exit }
    Scratch10 = Facing
    if TempleOutside == Geo { Scratch10 += 4 }
    CallStepForward()
    HoldMove = SkipMove
    Scratch1 = Table99A8[Scratch10]
    if Scratch1 == 26 { SymbolSet = 7 }
    if Scratch1 == 14 { SymbolSet = 3 }
    LoadMap(Scratch1)
`, New: `    if LeftMap == 0 { exit }
    Scratch10 = Facing
    if TempleOutside == Geo { Scratch10 += 4 }
    Scratch1 = Table99A8[Scratch10]
    if Scratch1 == 255 { exit } # fix: no area that way (the south openings): the party bumps instead of loading map 255
    CallStepForward()
    HoldMove = SkipMove
    if Scratch1 == 26 { SymbolSet = 7 }
    if Scratch1 == 14 { SymbolSet = 3 }
    LoadMap(Scratch1)
`},
	// 53: area 25 KeepRiders. Original: `goto [FightKeepRiders, ExitEvent][Scratch1]`.
	// Why: fix: LEAVE exited with the castle picture still up.
	// CHANGES: POR script, area 25 KeepRiders.
	{Area: 25, Old: `    Picture(Castle)
    Text("AS YOU APPROACH THE KEEP, A GROUP OF ARMED MEN ON HORESBACK RIDE OUT TO MEET YOU.  THEIR LEADER SAYS 'YOU ARE TRESSPASSING ON PRIVATE LAND; LEAVE NOW OR FACE THE CONSQUENCES!'")
    Scratch1 = Menu("STAY", exit="LEAVE")
    goto [FightKeepRiders, ExitEvent][Scratch1]
    goto FightKeepRiders

FightKeepRiders:
`, New: `    Picture(Castle)
    Text("AS YOU APPROACH THE KEEP, A GROUP OF ARMED MEN ON HORESBACK RIDE OUT TO MEET YOU.  THEIR LEADER SAYS 'YOU ARE TRESSPASSING ON PRIVATE LAND; LEAVE NOW OR FACE THE CONSQUENCES!'")
    Scratch1 = Menu("STAY", exit="LEAVE")
    goto [FightKeepRiders, Done][Scratch1] # fix: LEAVE exited with the castle picture still up
    goto FightKeepRiders

FightKeepRiders:
`},
	// 54: area 25 PassageOutside. Original: `Scratch1 = Menu(exit="LEAVE CAVE", "STAY IN")`.
	// Why: the menu's Esc row (exit=, 5c9b793) was the decompiler's guess LEAVE CAVE; Esc stays in, as the cave's other menus.
	// CHANGES: POR script, area 25 PassageOutside.
	{Area: 25, Old: `
PassageOutside:
    Text("YOU HAVE FOUND A PASSAGE WHICH LEADS OUTSIDE.  WILL YOU TAKE IT?")
    Scratch1 = Menu(exit="LEAVE CAVE", "STAY IN")
    goto [LeaveLair, Done][Scratch1]
    goto LeaveLair

`, New: `
PassageOutside:
    Text("YOU HAVE FOUND A PASSAGE WHICH LEADS OUTSIDE.  WILL YOU TAKE IT?")
    Scratch1 = Menu("LEAVE CAVE", exit="STAY IN")
    goto [LeaveLair, Done][Scratch1]
    goto LeaveLair

`},
	// 55: area 28 CellDispatch. Original: nothing here.
	// Why: fixed: the open west side of the guard post (4,0) skipped the commandant's attack at its south door (4,1).
	// CHANGES: POR script, area 28 CellDispatch.
	{Area: 28, Old: `}

CellDispatch:
    Scratch7 = Cell & ~Inside
    goto [Done, EscortIn, RestrictedGuards, QuartersEnter, ClimbableWall, DwarfAttackCheck][Scratch7]
    goto AlarmCheck
`, New: `}

CellDispatch:
    if PartyX == 3 and PartyY == 0 { goto DwarfAttackCheck } # fixed: the open west side of the guard post (4,0) skipped the commandant's attack at its south door (4,1)
    Scratch7 = Cell & ~Inside
    goto [Done, EscortIn, RestrictedGuards, QuartersEnter, ClimbableWall, DwarfAttackCheck][Scratch7]
    goto AlarmCheck
`},
	// 56: area 28 DwarfAndCommandant. Original: `LoadMonster(AlHyamDazid, count=1, icon=AlHyamDazid)`.
	// Why: fixed: was AlHyamDazid, met alive later (area5); block 82 (a mage) is not in MON6CHA and aborted the game; the disk has no other mage.
	// CHANGES: POR script, area 28 DwarfAndCommandant.
	{Area: 28, Old: `    SetupMonster(Fighter16, distance=2, closeup=EyepatchMan)
    ApproachTwice()
    ClearMonsters()
    LoadMonster(AlHyamDazid, count=1, icon=AlHyamDazid)
    LoadMonster(Commandant, count=1, icon=Captain)
    LoadMonster(Corporal, count=24, icon=Soldier)
    LoadMonster(Aides, count=20, icon=Soldier)
`, New: `    SetupMonster(Fighter16, distance=2, closeup=EyepatchMan)
    ApproachTwice()
    ClearMonsters()
    LoadMonster(ThLvlThief51, count=1, icon=AlHyamDazid) # fixed: was AlHyamDazid, met alive later (area5); block 82 (a mage) is not in MON6CHA and aborted the game; the disk has no other mage
    LoadMonster(Commandant, count=1, icon=Captain)
    LoadMonster(Corporal, count=24, icon=Soldier)
    LoadMonster(Aides, count=20, icon=Soldier)
`},
	// 57: area 28 RelaxOrWalk. Original: nothing here.
	// Why: fixed: no dinner invitation after fighting the guards.
	// CHANGES: POR script, area 28 RelaxOrWalk.
	{Area: 28, Old: `    goto Done

RelaxOrWalk:
    Text("YOU ARE IN YOUR QUARTERS.  WILL YOU WALK AROUND OR WILL YOU JUST RELAX UNTIL DINNER?")
    Scratch1 = Menu("WALK AROUND", "RELAX")
    goto [Done, EscortToDinner][Scratch1]
`, New: `    goto Done

RelaxOrWalk:
    if AttackedGuards != 0 or OutpostFought != 0 { goto Done } # fixed: no dinner invitation after fighting the guards
    Text("YOU ARE IN YOUR QUARTERS.  WILL YOU WALK AROUND OR WILL YOU JUST RELAX UNTIL DINNER?")
    Scratch1 = Menu("WALK AROUND", "RELAX")
    goto [Done, EscortToDinner][Scratch1]
`},
	// area 2: the original printed "A MUSCULAR MAN LOCKED IN CHAINS LIES HERE. NEAR HIM IS A PAGE THAT YOU PLACE IN YOUR JOURNAL UNDER ENTRY 47." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry47 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 2, Old: `    Print("A MUSCULAR MAN LOCKED IN CHAINS LIES HERE. NEAR HIM IS A PAGE THAT YOU PLACE IN YOUR JOURNAL UNDER ENTRY 47.")
`, New: `    Print("A MUSCULAR MAN LOCKED IN CHAINS LIES HERE. NEAR HIM IS A PAGE THAT YOU PLACE IN YOUR JOURNAL.")
    Journal(JournalEntry47)
`},
	// area 2: the original printed "YOU FIND A PAGE FROM A DIARY AND PLACE IT IN YOUR JOURNAL UNDER ENTRY 11." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry11 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 2, Old: `    Text("YOU FIND A PAGE FROM A DIARY AND PLACE IT IN YOUR JOURNAL UNDER ENTRY 11.")
`, New: `    Text("YOU FIND A PAGE FROM A DIARY AND PLACE IT IN YOUR JOURNAL.")
    Journal(JournalEntry11)
`},
	// area 5: the original printed "HE QUICKLY LEAVES. YOU ARE LEFT WITH THE NOTES OF AL-HYAM DHAZID, WHICH TELL QUITE A LOT ABOUT THE POOL OF RADIANCE. YOU COPY THEM INTO YOUR JOURNAL AS ENTRY 30." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry30 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 5, Old: `    Text("HE QUICKLY LEAVES. YOU ARE LEFT WITH THE NOTES OF AL-HYAM DHAZID, WHICH TELL QUITE A LOT ABOUT THE POOL OF RADIANCE. YOU COPY THEM INTO YOUR JOURNAL AS ENTRY 30.")
`, New: `    Text("HE QUICKLY LEAVES. YOU ARE LEFT WITH THE NOTES OF AL-HYAM DHAZID, WHICH TELL QUITE A LOT ABOUT THE POOL OF RADIANCE. YOU COPY THEM INTO YOUR JOURNAL.")
    Journal(JournalEntry30)
`},
	// area 10: the original printed "THE SCROLL IS PLACED IN YOUR JOURNAL AS ENTRY 43." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry43 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 10, Old: `    Text("THE SCROLL IS PLACED IN YOUR JOURNAL AS ENTRY 43.")
`, New: `    Text("THE SCROLL IS PLACED IN YOUR JOURNAL.")
    Journal(JournalEntry43)
`},
	// area 14: the original printed "INSIDE THE CABINET YOU FIND SOME FILES.  YOU STUFF THE PAPERS INTO YOUR JOURNAL UNDER ENTRIES 23 AND 14." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry23, JournalEntry14 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 14, Old: `    Text("INSIDE THE CABINET YOU FIND SOME FILES.  YOU STUFF THE PAPERS INTO YOUR JOURNAL UNDER ENTRIES 23 AND 14.")
`, New: `    Text("INSIDE THE CABINET YOU FIND SOME FILES.  YOU STUFF THE PAPERS INTO YOUR JOURNAL.")
    Journal(JournalEntry23)
    Journal(JournalEntry14)
`},
	// area 14: the original printed "ON THE FLOOR HERE YOU FIND SOME IMPORTANT LOOKING FILES.  YOU STUFF THE FILES INTO YOUR JOURNAL UNDER ENTRIES 38 AND 51." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry38, JournalEntry51 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 14, Old: `    Text("ON THE FLOOR HERE YOU FIND SOME IMPORTANT LOOKING FILES.  YOU STUFF THE FILES INTO YOUR JOURNAL UNDER ENTRIES 38 AND 51.")
`, New: `    Text("ON THE FLOOR HERE YOU FIND SOME IMPORTANT LOOKING FILES.  YOU STUFF THE FILES INTO YOUR JOURNAL.")
    Journal(JournalEntry38)
    Journal(JournalEntry51)
`},
	// area 14: the original printed "YOU COPY THE PAPERS INTO YOUR JOURNAL UNDER JOURNAL ENTRY 48." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry48 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 14, Old: `    Text("YOU COPY THE PAPERS INTO YOUR JOURNAL UNDER JOURNAL ENTRY 48.")
`, New: `    Text("YOU COPY THE PAPERS INTO YOUR JOURNAL.")
    Journal(JournalEntry48)
`},
	// area 14: the original printed "ON A NEARBY WALL IS A SLATE.  DRAWN ON THE SLATE IS A COMPLEX MAZE WITH THE WORDS 'NW CASTLE' UNDER IT.  YOU COPY THE CASTLE INTO YOUR JOURNAL UNDER JOURNAL ENTRY 41." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry41 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 14, Old: `    Text("ON A NEARBY WALL IS A SLATE.  DRAWN ON THE SLATE IS A COMPLEX MAZE WITH THE WORDS 'NW CASTLE' UNDER IT.  YOU COPY THE CASTLE INTO YOUR JOURNAL UNDER JOURNAL ENTRY 41.")
`, New: `    Text("ON A NEARBY WALL IS A SLATE.  DRAWN ON THE SLATE IS A COMPLEX MAZE WITH THE WORDS 'NW CASTLE' UNDER IT.  YOU COPY THE CASTLE INTO YOUR JOURNAL.")
    Journal(JournalEntry41)
`},
	// area 14: the original printed "ON THE FLOOR IS A CRUDE MAP.  IT IS A MAP OF A COMPLEX MAZE WITH THE WORDS 'NE CASTLE' UNDER IT.  YOU TAKE THE MAP AND PUT IT IN YOUR JOURNAL UNDER JOURNAL ENTRY 29." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry29 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 14, Old: `    Text("ON THE FLOOR IS A CRUDE MAP.  IT IS A MAP OF A COMPLEX MAZE WITH THE WORDS 'NE CASTLE' UNDER IT.  YOU TAKE THE MAP AND PUT IT IN YOUR JOURNAL UNDER JOURNAL ENTRY 29.")
`, New: `    Text("ON THE FLOOR IS A CRUDE MAP.  IT IS A MAP OF A COMPLEX MAZE WITH THE WORDS 'NE CASTLE' UNDER IT.  YOU TAKE THE MAP AND PUT IT IN YOUR JOURNAL.")
    Journal(JournalEntry29)
`},
	// area 15: the original printed "YOU FIND FYERDETHA'S DISCOURSES ON POWER. AMONG ALL THE DRY TEXT YOU FIND AN INTERESTING PASSAGE WHICH YOU COPY INTO YOUR JOURNAL AS ENTRY 7." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry7 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 15, Old: `    Text("YOU FIND FYERDETHA'S DISCOURSES ON POWER. AMONG ALL THE DRY TEXT YOU FIND AN INTERESTING PASSAGE WHICH YOU COPY INTO YOUR JOURNAL AS ENTRY 7.")
`, New: `    Text("YOU FIND FYERDETHA'S DISCOURSES ON POWER. AMONG ALL THE DRY TEXT YOU FIND AN INTERESTING PASSAGE WHICH YOU COPY INTO YOUR JOURNAL.")
    Journal(JournalEntry7)
`},
	// area 15: the original printed "YOU FIND URGUND'S DESCRIPTION OF DARKNESS. THIS IS AN ACCOUNT OF HIS IMPRISONMENT IN THE LOWER REALMS. THERE IS A PASSAGE OF INTEREST WHICH YOU COPY AS ENTRY 19 IN YOUR JOURNAL." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry19 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 15, Old: `    Text("YOU FIND URGUND'S DESCRIPTION OF DARKNESS. THIS IS AN ACCOUNT OF HIS IMPRISONMENT IN THE LOWER REALMS. THERE IS A PASSAGE OF INTEREST WHICH YOU COPY AS ENTRY 19 IN YOUR JOURNAL.")
`, New: `    Text("YOU FIND URGUND'S DESCRIPTION OF DARKNESS. THIS IS AN ACCOUNT OF HIS IMPRISONMENT IN THE LOWER REALMS. THERE IS A PASSAGE OF INTEREST WHICH YOU COPY AS PART OF YOUR JOURNAL.")
    Journal(JournalEntry19)
`},
	// area 15: the original printed "YOU FIND LEX GEOGRAPHICA, AN ATLAS DRAWN BY TOMARUS. IT HAS A MAP OF PHLAN WHICH, THOUGH OLD, STILL COULD BE USEFUL. IT BECOMES ENTRY 37 IN YOUR JOURNAL." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry37 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 15, Old: `    Text("YOU FIND LEX GEOGRAPHICA, AN ATLAS DRAWN BY TOMARUS. IT HAS A MAP OF PHLAN WHICH, THOUGH OLD, STILL COULD BE USEFUL. IT BECOMES ENTRY 37 IN YOUR JOURNAL.")
`, New: `    Text("YOU FIND LEX GEOGRAPHICA, AN ATLAS DRAWN BY TOMARUS. IT HAS A MAP OF PHLAN WHICH, THOUGH OLD, STILL COULD BE USEFUL. IT BECOMES PART OF YOUR JOURNAL.")
    Journal(JournalEntry37)
`},
	// area 15: the original printed "YOU FIND THE HISTORY OF THE NORTH. IT IS MOSTLY WRITTEN TO PLEASE A ROYAL LINE. HOWEVER, THERE IS AN INTERESTING PASSAGE WHICH YOU COPY INTO YOUR JOURNAL AS ENTRY 8." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry8 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 15, Old: `    Text("YOU FIND THE HISTORY OF THE NORTH. IT IS MOSTLY WRITTEN TO PLEASE A ROYAL LINE. HOWEVER, THERE IS AN INTERESTING PASSAGE WHICH YOU COPY INTO YOUR JOURNAL AS ENTRY 8.")
`, New: `    Text("YOU FIND THE HISTORY OF THE NORTH. IT IS MOSTLY WRITTEN TO PLEASE A ROYAL LINE. HOWEVER, THERE IS AN INTERESTING PASSAGE WHICH YOU COPY INTO YOUR JOURNAL.")
    Journal(JournalEntry8)
`},
	// area 15: the original printed "YOU FIND THE GRAND HISTORIAN'S RECORDS OF THE ARTS OF WAR. IN IT YOU FIND A USEFUL PASSAGE WHICH YOU COPY INTO YOUR JOURNAL AS ENTRY 21." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry21 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 15, Old: `    Text("YOU FIND THE GRAND HISTORIAN'S RECORDS OF THE ARTS OF WAR. IN IT YOU FIND A USEFUL PASSAGE WHICH YOU COPY INTO YOUR JOURNAL AS ENTRY 21.")
`, New: `    Text("YOU FIND THE GRAND HISTORIAN'S RECORDS OF THE ARTS OF WAR. IN IT YOU FIND A USEFUL PASSAGE WHICH YOU COPY INTO YOUR JOURNAL.")
    Journal(JournalEntry21)
`},
	// area 15: the original printed "THEY DESCRIBE THE SURROUNDING AREA TO YOU. YOU LISTEN CAREFULLY, QUESTION THEM AND FROM THIS, SKETCH A MAP INTO YOUR JOURNAL CALLING IT ENTRY 10." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry10 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 15, Old: `    Text("THEY DESCRIBE THE SURROUNDING AREA TO YOU. YOU LISTEN CAREFULLY, QUESTION THEM AND FROM THIS, SKETCH A MAP INTO YOUR JOURNAL CALLING IT ENTRY 10.")
`, New: `    Text("THEY DESCRIBE THE SURROUNDING AREA TO YOU. YOU LISTEN CAREFULLY, QUESTION THEM AND FROM THIS, SKETCH A MAP INTO YOUR JOURNAL.")
    Journal(JournalEntry10)
`},
	// area 16: the original printed "ONE OF YOU BUSILY SCRIBBLE IT INTO THE JOURNAL AS ENTRY 31." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry31 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 16, Old: `    Text("ONE OF YOU BUSILY SCRIBBLE IT INTO THE JOURNAL AS ENTRY 31.")
`, New: `    Text("ONE OF YOU BUSILY SCRIBBLE IT INTO THE JOURNAL.")
    Journal(JournalEntry31)
`},
	// area 17: the original printed "DURING THE FESTIVITIES THE CHIEF (HIS TONGUE LOOSENED BY DRINK) TELLS YOU OF THE DANGERS OF THE REGION. YOU MAKE HASSAD AHWAL'S TALK ENTRY 55 IN YOUR JOURNAL." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry55 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 17, Old: `    Text("DURING THE FESTIVITIES THE CHIEF (HIS TONGUE LOOSENED BY DRINK) TELLS YOU OF THE DANGERS OF THE REGION. YOU MAKE HASSAD AHWAL'S TALK ENTRY 55 IN YOUR JOURNAL.")
`, New: `    Text("DURING THE FESTIVITIES THE CHIEF (HIS TONGUE LOOSENED BY DRINK) TELLS YOU OF THE DANGERS OF THE REGION. YOU MAKE HASSAD AHWAL'S TALK PART OF YOUR JOURNAL.")
    Journal(JournalEntry55)
`},
	// area 21: the original printed "BENEATH A FLOORBOARD YOU FIND SOME GEMS AND A FEW PAGES OF A DIARY.  YOU COPY THESE PAGES INTO YOUR JOURNAL AS ENTRY 3." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry3 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 21, Old: `    Text("BENEATH A FLOORBOARD YOU FIND SOME GEMS AND A FEW PAGES OF A DIARY.  YOU COPY THESE PAGES INTO YOUR JOURNAL AS ENTRY 3.")
`, New: `    Text("BENEATH A FLOORBOARD YOU FIND SOME GEMS AND A FEW PAGES OF A DIARY.  YOU COPY THESE PAGES INTO YOUR JOURNAL.")
    Journal(JournalEntry3)
`},
	// area 21: the original printed "ON A HOBGOBLIN BODY, YOU FIND A NOTE THAT YOU COPY INTO YOUR JOURNAL AS ENTRY 57." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry57 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 21, Old: `    Text("ON A HOBGOBLIN BODY, YOU FIND A NOTE THAT YOU COPY INTO YOUR JOURNAL AS ENTRY 57.")
`, New: `    Text("ON A HOBGOBLIN BODY, YOU FIND A NOTE THAT YOU COPY INTO YOUR JOURNAL.")
    Journal(JournalEntry57)
`},
	// area 22: the original printed "HOWEVER, YOU FIND A BOOK UNDER THE BODY. IT IS THE JOURNAL OF TYRMOR.  IT IS FACINATING AND YOU INCLUDE MUCH OF ITS INFORMATION IN YOUR JOURNAL AS ENTRY 26." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry26 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 22, Old: `    Text("HOWEVER, YOU FIND A BOOK UNDER THE BODY. IT IS THE JOURNAL OF TYRMOR.  IT IS FACINATING AND YOU INCLUDE MUCH OF ITS INFORMATION IN YOUR JOURNAL AS ENTRY 26.")
`, New: `    Text("HOWEVER, YOU FIND A BOOK UNDER THE BODY. IT IS THE JOURNAL OF TYRMOR.  IT IS FACINATING AND YOU INCLUDE MUCH OF ITS INFORMATION IN YOUR JOURNAL.")
    Journal(JournalEntry26)
`},
	// area 23: the original printed "AS YOU FREE THEM THEY TELL YOU ALL ABOUT WHAT YARASH IS DOING AND HOW YOU CAN GET HELP FROM THEIR TRIBE IN THE SWAMP. YOU COPY ALL THEY SAY AS ENTRY 35 OF YOUR JOURNAL." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry35 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 23, Old: `    Text("AS YOU FREE THEM THEY TELL YOU ALL ABOUT WHAT YARASH IS DOING AND HOW YOU CAN GET HELP FROM THEIR TRIBE IN THE SWAMP. YOU COPY ALL THEY SAY AS ENTRY 35 OF YOUR JOURNAL.")
`, New: `    Text("AS YOU FREE THEM THEY TELL YOU ALL ABOUT WHAT YARASH IS DOING AND HOW YOU CAN GET HELP FROM THEIR TRIBE IN THE SWAMP. YOU COPY ALL THEY SAY INTO YOUR JOURNAL.")
    Journal(JournalEntry35)
`},
	// area 23: the original printed "A LETTER FROM THE BOSS, TELLING YARASH TO JOIN OR DIE. (ENTRY 33)" for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry33 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 23, Old: `    Text("A LETTER FROM THE BOSS, TELLING YARASH TO JOIN OR DIE. (ENTRY 33)")
`, New: `    Text("A LETTER FROM THE BOSS, TELLING YARASH TO JOIN OR DIE.")
    Journal(JournalEntry33)
`},
	// area 23: the original printed "AN UNSENT REPLY TO THIS LETTER. (ENTRY 49)" for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry49 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 23, Old: `    Text("AN UNSENT REPLY TO THIS LETTER. (ENTRY 49)")
`, New: `    Text("AN UNSENT REPLY TO THIS LETTER.")
    Journal(JournalEntry49)
`},
	// area 23: the original printed "A MESSAGE TO THE BUCCANEERS OFFERING A BOUNTY FOR A REAL SAHUAGIN. (ENTRY 27)" for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry27 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 23, Old: `    Text("A MESSAGE TO THE BUCCANEERS OFFERING A BOUNTY FOR A REAL SAHUAGIN. (ENTRY 27)")
`, New: `    Text("A MESSAGE TO THE BUCCANEERS OFFERING A BOUNTY FOR A REAL SAHUAGIN.")
    Journal(JournalEntry27)
`},
	// area 23: the original printed "A WARNING TO AN AGENT ABOUT THE DANGERS IN THE MOUNTAINS TO THE NORTHWEST. (ENTRY 56)" for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry56 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 23, Old: `    Text("A WARNING TO AN AGENT ABOUT THE DANGERS IN THE MOUNTAINS TO THE NORTHWEST. (ENTRY 56)")
`, New: `    Text("A WARNING TO AN AGENT ABOUT THE DANGERS IN THE MOUNTAINS TO THE NORTHWEST.")
    Journal(JournalEntry56)
`},
	// area 23: the original printed "A NOTE TO A TRAPPER ABOUT KOBOLDS AND HOBGOBLINS. (ENTRY 40)" for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry40 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 23, Old: `    Text("A NOTE TO A TRAPPER ABOUT KOBOLDS AND HOBGOBLINS. (ENTRY 40)")
`, New: `    Text("A NOTE TO A TRAPPER ABOUT KOBOLDS AND HOBGOBLINS.")
    Journal(JournalEntry40)
`},
	// area 24: the original printed "YOU FIND A PIECE OF PAPER ON THE FLOOR IN THE ROOM.  YOU PLACE IT IN YOUR JOURNAL UNDER ENTRY 53" for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry53 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 24, Old: `    Text("YOU FIND A PIECE OF PAPER ON THE FLOOR IN THE ROOM.  YOU PLACE IT IN YOUR JOURNAL UNDER ENTRY 53")
`, New: `    Text("YOU FIND A PIECE OF PAPER ON THE FLOOR IN THE ROOM.  YOU PLACE IT IN YOUR JOURNAL")
    Journal(JournalEntry53)
`},
	// area 24: the original printed "ON THE DEAD BODY OF THE HALF-ORC CLERIC YOU FIND A PIECE OF PAPER.  YOU PUT THE PAPER IN YOUR JOURNAL UNDER ENTRY 25." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry25 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 24, Old: `    Text("ON THE DEAD BODY OF THE HALF-ORC CLERIC YOU FIND A PIECE OF PAPER.  YOU PUT THE PAPER IN YOUR JOURNAL UNDER ENTRY 25.")
`, New: `    Text("ON THE DEAD BODY OF THE HALF-ORC CLERIC YOU FIND A PIECE OF PAPER.  YOU PUT THE PAPER IN YOUR JOURNAL.")
    Journal(JournalEntry25)
`},
	// area 28: the original printed "AS YOU MENTION THE POOL OF RADIANCE THE COMMANDANT TELLS YOU A STORY ABOUT THAT VERY POOL.  YOU ENJOY THE STORY AND COPY IT INTO YOUR JOURNAL UNDER ENTRY 46." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry46 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 28, Old: `    Text("AS YOU MENTION THE POOL OF RADIANCE THE COMMANDANT TELLS YOU A STORY ABOUT THAT VERY POOL.  YOU ENJOY THE STORY AND COPY IT INTO YOUR JOURNAL UNDER ENTRY 46.")
`, New: `    Text("AS YOU MENTION THE POOL OF RADIANCE THE COMMANDANT TELLS YOU A STORY ABOUT THAT VERY POOL.  YOU ENJOY THE STORY AND COPY IT INTO YOUR JOURNAL.")
    Journal(JournalEntry46)
`},
	// area 29: the original printed "ON THE LEADER'S BODY, YOU FIND A CURIOUS MESSAGE THAT YOU PLACE IN YOUR JOURNAL AS ENTRY 50." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry50 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 29, Old: `    Text("ON THE LEADER'S BODY, YOU FIND A CURIOUS MESSAGE THAT YOU PLACE IN YOUR JOURNAL AS ENTRY 50.")
`, New: `    Text("ON THE LEADER'S BODY, YOU FIND A CURIOUS MESSAGE THAT YOU PLACE IN YOUR JOURNAL.")
    Journal(JournalEntry50)
`},
	// area 13: the original printed " YOU RECORD IT IN YOUR JOURNAL UNDER ENTRY " for the printed Adventurer's Journal.
	// Why: the pager shows the entries (the callers name them) (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 13, Old: `def JournalEntry() {
    Print(" YOU RECORD IT IN YOUR JOURNAL UNDER ENTRY ")
`, New: `def JournalEntry() {
    Print(" YOU RECORD IT IN YOUR JOURNAL.")
`},
	// area 13: the original printed "42." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry42 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 13, Old: `    JournalEntry()
    Print("42.")
`, New: `    JournalEntry()
    Journal(JournalEntry42)
`},
	// area 13: the original printed "28." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry28 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 13, Old: `    JournalEntry()
    Print("28.")
`, New: `    JournalEntry()
    Journal(JournalEntry28)
`},
	// area 13: the original printed "16." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry16 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 13, Old: `    JournalEntry()
    Print("16.")
`, New: `    JournalEntry()
    Journal(JournalEntry16)
`},
	// area 13: the original printed "20." for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry20 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 13, Old: `    JournalEntry()
    Print("20.")
`, New: `    JournalEntry()
    Journal(JournalEntry20)
`},
	// area 13: the original printed "16" for the printed Adventurer's Journal.
	// Why: the pager shows JournalEntry16 (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 13, Old: `    JournalEntry()
    Print("16")
`, New: `    JournalEntry()
    Journal(JournalEntry16)
`},
	// area 0: the original printed "PROCLAMATIONS ARE POSTED ON THE WALLS, IN YOUR JOURNAL YOU NOTE " for the printed Adventurer's Journal.
	// Why: the pager shows the entries (the callers name them) (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 0, Old: `    Text("PROCLAMATIONS ARE POSTED ON THE WALLS, IN YOUR JOURNAL YOU NOTE ")
`, New: `    Text("PROCLAMATIONS ARE POSTED ON THE WALLS.")
`},
	// area 0: the original printed "PROCLAMATION, $AuxRomanNum; PROCLAMATIONS LXIV, LXXVIII, CIX, AND LIX." for the printed Adventurer's Journal.
	// Why: the pager shows ProclamationLXIV, ProclamationLXXVIII, ProclamationCIX, ProclamationLIX (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 0, Old: `    Print("PROCLAMATION")
    Print($AuxRomanNum)
    exit

    OldProclamations:
    Print("PROCLAMATIONS LXIV, LXXVIII, CIX, AND LIX.")
    exit
`, New: `    exit

    OldProclamations:
    Journal(ProclamationLXIV)
    Journal(ProclamationLXXVIII)
    Journal(ProclamationCIX)
    Journal(ProclamationLIX)
    exit
`},
	// area 0: the original printed "CI." for the printed Adventurer's Journal.
	// Why: the pager shows ProclamationCI (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 0, Old: `def Proclamation1() {
    $AuxRomanNum = "CI."
`, New: `def Proclamation1() {
    $AuxRomanNum = "CI."
    Journal(ProclamationCI)
`},
	// area 0: the original printed "CXXVI AND CX." for the printed Adventurer's Journal.
	// Why: the pager shows ProclamationCXXVI, ProclamationCX (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 0, Old: `def Proclamation2() {
    $AuxRomanNum = "CXXVI AND CX."
`, New: `def Proclamation2() {
    $AuxRomanNum = "CXXVI AND CX."
    Journal(ProclamationCXXVI)
    Journal(ProclamationCX)
`},
	// area 0: the original printed "CXXXIV." for the printed Adventurer's Journal.
	// Why: the pager shows ProclamationCXXXIV (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 0, Old: `def Proclamation3() {
    $AuxRomanNum = "CXXXIV."
`, New: `def Proclamation3() {
    $AuxRomanNum = "CXXXIV."
    Journal(ProclamationCXXXIV)
`},
	// area 0: the original printed "CLIV." for the printed Adventurer's Journal.
	// Why: the pager shows ProclamationCLIV (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 0, Old: `def Proclamation4() {
    $AuxRomanNum = "CLIV."
`, New: `def Proclamation4() {
    $AuxRomanNum = "CLIV."
    Journal(ProclamationCLIV)
`},
	// area 0: the original printed "CXIV." for the printed Adventurer's Journal.
	// Why: the pager shows ProclamationCXIV (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 0, Old: `def Proclamation5() {
    $AuxRomanNum = "CXIV."
`, New: `def Proclamation5() {
    $AuxRomanNum = "CXIV."
    Journal(ProclamationCXIV)
`},
	// area 0: the original printed "CCIV." for the printed Adventurer's Journal.
	// Why: the pager shows ProclamationCCIV (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 0, Old: `def Proclamation6() {
    $AuxRomanNum = "CCIV."
`, New: `def Proclamation6() {
    $AuxRomanNum = "CCIV."
    Journal(ProclamationCCIV)
`},
	// area 0: the original printed "CXXIX." for the printed Adventurer's Journal.
	// Why: the pager shows ProclamationCXXIX (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 0, Old: `def Proclamation7() {
    $AuxRomanNum = "CXXIX."
`, New: `def Proclamation7() {
    $AuxRomanNum = "CXXIX."
    Journal(ProclamationCXXIX)
`},
	// area 0: the original printed "CCI." for the printed Adventurer's Journal.
	// Why: the pager shows ProclamationCCI (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 0, Old: `def Proclamation8() {
    $AuxRomanNum = "CCI."
`, New: `def Proclamation8() {
    $AuxRomanNum = "CCI."
    Journal(ProclamationCCI)
`},
	// area 0: the original printed "CXIV." for the printed Adventurer's Journal.
	// Why: the pager shows ProclamationCXIV (Journal), not a match on the printed reference.
	// CHANGES: POR script, journal references.
	{Area: 0, Old: `def Proclamation9() {
    $AuxRomanNum = "CXIV."
`, New: `def Proclamation9() {
    $AuxRomanNum = "CXIV."
    Journal(ProclamationCXIV)
`},
	// area 8: the original handed quest ClearTheSlums over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("THE SLUMS IMMEDIATELY TO OUR WEST NEED TO BE CLEARED OF MONSTERS.'")
`, New: `    Text("THE SLUMS IMMEDIATELY TO OUR WEST NEED TO BE CLEARED OF MONSTERS.'")
    Quest(ClearTheSlums)
`},
	// area 8: the original handed quest SokalKeep over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("SOKAL KEEP ON THORN ISLAND MUST BE CLEARED.'")
`, New: `    Text("SOKAL KEEP ON THORN ISLAND MUST BE CLEARED.'")
    Quest(SokalKeep)
`},
	// area 8: the original handed quest LoreOfOldPhlan over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("THE COUNCIL IS OFFERING A REWARD FOR BOOKS, MAPS, TOMES, ETC. WHICH PROVIDE USEFUL INFORMATION ABOUT PHLAN BEFORE THE FALL.  THE REWARD IS TIED TO THE VALUE OF THE INFORMATION.'")
`, New: `    Text("THE COUNCIL IS OFFERING A REWARD FOR BOOKS, MAPS, TOMES, ETC. WHICH PROVIDE USEFUL INFORMATION ABOUT PHLAN BEFORE THE FALL.  THE REWARD IS TIED TO THE VALUE OF THE INFORMATION.'")
    Quest(LoreOfOldPhlan)
`},
	// area 8: the original handed quest PodalPlazaAuction over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("A WEAPON OF GREAT POWER IS TO BE AUCTIONED TO OUR ENEMIES. THIS AUCTION IS TO BE HELD IN PODAL PLAZA.  FIND OUT WHAT THE WEAPON IS AND RETURN.'")
`, New: `    Text("A WEAPON OF GREAT POWER IS TO BE AUCTIONED TO OUR ENEMIES. THIS AUCTION IS TO BE HELD IN PODAL PLAZA.  FIND OUT WHAT THE WEAPON IS AND RETURN.'")
    Quest(PodalPlazaAuction)
`},
	// area 8: the original handed quest CadornasTreasure over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("JUNIOR COUNCILMAN CADORNA HAS A SPECIAL COMMISSION FOR YOU.  HIS CHAMBERS ARE THROUGH THE EAST DOOR.'")
`, New: `    Text("JUNIOR COUNCILMAN CADORNA HAS A SPECIAL COMMISSION FOR YOU.  HIS CHAMBERS ARE THROUGH THE EAST DOOR.'")
    Quest(CadornasTreasure)
`},
	// area 8: the original handed quest CadornasTreasure over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("'WHEN OLD PHLAN WAS OVERRUN, A FAMILY TREASURE WAS HIDDEN IN THE WESTERN BUILDING OF THE TEXTILE COMPLEX.")
`, New: `    Text("'WHEN OLD PHLAN WAS OVERRUN, A FAMILY TREASURE WAS HIDDEN IN THE WESTERN BUILDING OF THE TEXTILE COMPLEX.")
    Quest(CadornasTreasure)
`},
	// area 8: the original handed quest CadornasTreasure over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("'THE FAITHFUL SERVANT SENT TO FETCH IT NEVER RETURNED.  THE COMPLEX IS JUST SOUTH OF PODOL PLAZA.  BRING THE TREASURE TO ME AND YOU WILL BE WELL REWARDED.'")
`, New: `    Text("'THE FAITHFUL SERVANT SENT TO FETCH IT NEVER RETURNED.  THE COMPLEX IS JUST SOUTH OF PODOL PLAZA.  BRING THE TREASURE TO ME AND YOU WILL BE WELL REWARDED.'")
    Quest(CadornasTreasure)
`},
	// area 8: the original handed quest TempleOfBane over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("YOU ARE SUMMONED BY THE BISHOP OF BRACCIO OF TYR. REPORT TO HIM AT TYR'S TEMPLE.'")
`, New: `    Text("YOU ARE SUMMONED BY THE BISHOP OF BRACCIO OF TYR. REPORT TO HIM AT TYR'S TEMPLE.'")
    Quest(TempleOfBane)
`},
	// area 0: the original handed quest TempleOfBane over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 0, Old: `    Text("YOU ARE USHERED INTO THE BISHOP'S STUDY. BISHOP BRACCIO SPEAKS, 'ALLOW ME TO INTRODUCE DIRTEN, PRIEST OF ILMAR.")
`, New: `    Text("YOU ARE USHERED INTO THE BISHOP'S STUDY. BISHOP BRACCIO SPEAKS, 'ALLOW ME TO INTRODUCE DIRTEN, PRIEST OF ILMAR.")
    Quest(TempleOfBane)
`},
	// area 0: the original handed quest TempleOfBane over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 0, Old: `    Text("'HE IS BOUND TO RECOVER THE TEMPLE WHICH HAS BEEN DESECRATED TO BANE.  GO WITH HIM ACROSS THE RIVER AND HELP HIM TO CLEANSE THE TEMPLE.  IN PAYMENT, YOU MAY KEEP THE HIDDEN TREASURES.'")
`, New: `    Text("'HE IS BOUND TO RECOVER THE TEMPLE WHICH HAS BEEN DESECRATED TO BANE.  GO WITH HIM ACROSS THE RIVER AND HELP HIM TO CLEANSE THE TEMPLE.  IN PAYMENT, YOU MAY KEEP THE HIDDEN TREASURES.'")
    Quest(TempleOfBane)
`},
	// area 8: the original handed quest KovelMansion over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("A LARGE GROUP OF THIEVES OPERATES OUT OF THE OLD KOVEL MANSION. THE COUNCIL WILL PAY TO HAVE THEM CLEANED OUT.'")
`, New: `    Text("A LARGE GROUP OF THIEVES OPERATES OUT OF THE OLD KOVEL MANSION. THE COUNCIL WILL PAY TO HAVE THEM CLEANED OUT.'")
    Quest(KovelMansion)
`},
	// area 8: the original handed quest TheNomads over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("FIND THE NOMADS AND STOP THEM FROM JOINING FORCES WITH OUR ENEMIES.'")
`, New: `    Text("FIND THE NOMADS AND STOP THEM FROM JOINING FORCES WITH OUR ENEMIES.'")
    Quest(TheNomads)
`},
	// area 8: the original handed quest TheKoboldTribe over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("A LARGE TRIBE OF KOBOLDS IS BEING RECRUITED BY THE ENEMY.  MAKE SURE THEY DON'T JOIN THE ENEMY FORCES.'")
`, New: `    Text("A LARGE TRIBE OF KOBOLDS IS BEING RECRUITED BY THE ENEMY.  MAKE SURE THEY DON'T JOIN THE ENEMY FORCES.'")
    Quest(TheKoboldTribe)
`},
	// area 8: the original handed quest ThePoisonedRiver over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("FIND THE SOURCE OF THE RIVER'S POLLUTION AND END IT.'")
`, New: `    Text("FIND THE SOURCE OF THE RIVER'S POLLUTION AND END IT.'")
    Quest(ThePoisonedRiver)
`},
	// area 8: the original handed quest TheLizardmen over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("A TRIBE OF LIZARDMEN IS PREPARING TO JOIN THE ENEMY.  PREVENT THIS.'")
`, New: `    Text("A TRIBE OF LIZARDMEN IS PREPARING TO JOIN THE ENEMY.  PREVENT THIS.'")
    Quest(TheLizardmen)
`},
	// area 8: the original handed quest TheHeirOfBivant over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("THE HEIR TO THE HOUSE OF BIVANT MUST BE RESCUED.  WE WILL PAY GENEROUSLY FOR HIS SAFE RETURN.'")
`, New: `    Text("THE HEIR TO THE HOUSE OF BIVANT MUST BE RESCUED.  WE WILL PAY GENEROUSLY FOR HIS SAFE RETURN.'")
    Quest(TheHeirOfBivant)
`},
	// area 8: the original handed quest ZhentilEnvoys over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("LORD CADORNA HAS A DELICATE DIPLOMATIC MISSION THAT REQUIRES YOUR PARTICULAR SKILLS.  YOU ARE SUMMONED TO HIS CHAMBERS. GO THROUGH THE SOUTH DOOR.'")
`, New: `    Text("LORD CADORNA HAS A DELICATE DIPLOMATIC MISSION THAT REQUIRES YOUR PARTICULAR SKILLS.  YOU ARE SUMMONED TO HIS CHAMBERS. GO THROUGH THE SOUTH DOOR.'")
    Quest(ZhentilEnvoys)
`},
	// area 8: the original handed quest ZhentilEnvoys over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("'IT IS IMPERATIVE THAT A MESSAGE BE DELIVERED TO THE ZHENTIL OUTPOST TO THE WEST.  I AM SENDING YOU AS DIPLOMATIC ENVOYS.")
`, New: `    Text("'IT IS IMPERATIVE THAT A MESSAGE BE DELIVERED TO THE ZHENTIL OUTPOST TO THE WEST.  I AM SENDING YOU AS DIPLOMATIC ENVOYS.")
    Quest(ZhentilEnvoys)
`},
	// area 8: the original handed quest ZhentilEnvoys over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("'WHEN YOU GET TO ZHENTIAL KEEP, PRESENT THIS PACKET TO ITS COMMANDANT.  DO NOT BREAK THE SEALS OR YOU WILL BE SEIZED AS SPIES.'")
`, New: `    Text("'WHEN YOU GET TO ZHENTIAL KEEP, PRESENT THIS PACKET TO ITS COMMANDANT.  DO NOT BREAK THE SEALS OR YOU WILL BE SEIZED AS SPIES.'")
    Quest(ZhentilEnvoys)
`},
	// area 8: the original handed quest ZhentilEnvoys over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("WE NEED YOU TO DELIVER A MESSAGE TO THE ZHENTIL KEEP TO THE WEST. BE CAREFUL, THE TRAITOR CADORNA MAY HAVE TAMPERED WITH THE MESSAGE PACKET BEFORE HE FLED.'")
`, New: `    Text("WE NEED YOU TO DELIVER A MESSAGE TO THE ZHENTIL KEEP TO THE WEST. BE CAREFUL, THE TRAITOR CADORNA MAY HAVE TAMPERED WITH THE MESSAGE PACKET BEFORE HE FLED.'")
    Quest(ZhentilEnvoys)
`},
	// area 8: the original handed quest StojanowGate over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("LORD URSLINGEN WISHES URGENTLY TO SPEAK WITH YOU.  GO THROUGH THE SOUTH DOOR OF THE TRAITOR CADORNA'S OLD QUARTERS.'")
`, New: `    Text("LORD URSLINGEN WISHES URGENTLY TO SPEAK WITH YOU.  GO THROUGH THE SOUTH DOOR OF THE TRAITOR CADORNA'S OLD QUARTERS.'")
    Quest(StojanowGate)
`},
	// area 8: the original handed quest StojanowGate over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("'THANK YOU FOR COMING.  I MUST BE BRIEF FOR OUR NEED IS URGENT.  OUR ONE CHANCE TO DEFEAT THE ENEMY IS THROUGH AN ASSAULT ON THE CASTLE. TO DO THIS, WE MUST FIRST SECURE STOJANOW GATE.")
`, New: `    Text("'THANK YOU FOR COMING.  I MUST BE BRIEF FOR OUR NEED IS URGENT.  OUR ONE CHANCE TO DEFEAT THE ENEMY IS THROUGH AN ASSAULT ON THE CASTLE. TO DO THIS, WE MUST FIRST SECURE STOJANOW GATE.")
    Quest(StojanowGate)
`},
	// area 8: the original handed quest StojanowGate over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("'THE GATE IS TOO STRONG TO BE TAKEN BY STORM. A SMALL BAND MUST SLIP INSIDE AND HOLD UNTIL RELIEVED. YOU ARE THE BEST AND THE BRAVEST. THE COUNCIL HAS INSTRUCTED ME TO ENTRUST THE TASK TO YOU.'")
`, New: `    Text("'THE GATE IS TOO STRONG TO BE TAKEN BY STORM. A SMALL BAND MUST SLIP INSIDE AND HOLD UNTIL RELIEVED. YOU ARE THE BEST AND THE BRAVEST. THE COUNCIL HAS INSTRUCTED ME TO ENTRUST THE TASK TO YOU.'")
    Quest(StojanowGate)
`},
	// area 8: the original handed quest ValjevoCastle over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("YOU ARE SUMMONED TO A SPECIAL COUNCIL MEETING.  YOU ARE TO BE AWARDED A VALUABLE COMMISSION. THE CITY GUARDS WILL SHOW YOU THE WAY.'")
`, New: `    Text("YOU ARE SUMMONED TO A SPECIAL COUNCIL MEETING.  YOU ARE TO BE AWARDED A VALUABLE COMMISSION. THE CITY GUARDS WILL SHOW YOU THE WAY.'")
    Quest(ValjevoCastle)
`},
	// area 8: the original handed quest ValjevoCastle over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("IN THE COUNCIL CHAMBERS, THE ENTIRE CITY COUNCIL IS WAITING FOR YOU. THE LORD HIGH MAYOR, ULRICH EBERHARD, SPEAKS: 'OUR FORCES ARE MARSHALLED FOR THE ASSAULT ON VALJEVO CASTLE.")
`, New: `    Text("IN THE COUNCIL CHAMBERS, THE ENTIRE CITY COUNCIL IS WAITING FOR YOU. THE LORD HIGH MAYOR, ULRICH EBERHARD, SPEAKS: 'OUR FORCES ARE MARSHALLED FOR THE ASSAULT ON VALJEVO CASTLE.")
    Quest(ValjevoCastle)
`},
	// area 8: the original handed quest ValjevoCastle over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("'IT IS THE UNANIMOUS REQUEST OF THE COUNCIL THAT YOU LEAD THE ATTACK. YOUR SKILL AND BRAVERY HAVE WON THE ADMIRATION OF ALL PHLAN.  WE LOOK TO YOU TO HELP US WIN OUR FREEDOM.'")
`, New: `    Text("'IT IS THE UNANIMOUS REQUEST OF THE COUNCIL THAT YOU LEAD THE ATTACK. YOUR SKILL AND BRAVERY HAVE WON THE ADMIRATION OF ALL PHLAN.  WE LOOK TO YOU TO HELP US WIN OUR FREEDOM.'")
    Quest(ValjevoCastle)
`},
	// area 8: the original handed quest ValhingenGraveyard over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("PUTTING DOWN THE PAPERS, SHE SAYS, 'I MUST BRING TO YOUR ATTENTION THE FOLLOWING CONCERNING VALHINGEN GRAVEYARD: '")
`, New: `    Text("PUTTING DOWN THE PAPERS, SHE SAYS, 'I MUST BRING TO YOUR ATTENTION THE FOLLOWING CONCERNING VALHINGEN GRAVEYARD: '")
    Quest(ValhingenGraveyard)
`},
	// area 8: the original handed quest ValhingenGraveyard over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("THE UNDEAD ARE CROSSING THE RIVER IN HORDES AND KILLING OR CARRYING OFF OUR SETTLERS.  THE COUNCIL DEMANDS THESE ATTACKS BE STOPPED. A GENEROUS REWARD IS OFFERED.'")
`, New: `    Text("THE UNDEAD ARE CROSSING THE RIVER IN HORDES AND KILLING OR CARRYING OFF OUR SETTLERS.  THE COUNCIL DEMANDS THESE ATTACKS BE STOPPED. A GENEROUS REWARD IS OFFERED.'")
    Quest(ValhingenGraveyard)
`},
	// area 8: the original handed quest ValhingenGraveyard over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("UNDEAD FROM THE GRAVEYARD HAVE GROWN MORE DANGEROUS THAN ALL OF THE OTHER FORCES WHICH CONFRONT US.")
`, New: `    Text("UNDEAD FROM THE GRAVEYARD HAVE GROWN MORE DANGEROUS THAN ALL OF THE OTHER FORCES WHICH CONFRONT US.")
    Quest(ValhingenGraveyard)
`},
	// area 8: the original handed quest ValhingenGraveyard over by this text alone (no flag for most of the clerk's offers).
	// Why: the quest journal records it by Quest, not by matching the printed text.
	// CHANGES: POR script, quest handovers.
	{Area: 8, Old: `    Text("'I HAVE BEEN AUTHORIZED TO GIVE YOU AN ENCHANTED WEAPON IF YOU ACCEPT THE COMMISSION TO END THE GRAVEYARD MENACE.  DO YOU ACCEPT?'")
`, New: `    Text("'I HAVE BEEN AUTHORIZED TO GIVE YOU AN ENCHANTED WEAPON IF YOU ACCEPT THE COMMISSION TO END THE GRAVEYARD MENACE.  DO YOU ACCEPT?'")
    Quest(ValhingenGraveyard)
`},
	// area 13 CheckTrap: the original's CheckParty wrote found (0/1) as its fourth result.
	// Why: pool's CheckParty writes the count for every game (S-diffs 12); found is count > 0, the same branch.
	// CHANGES: none (no behaviour change).
	{Area: 13, Old: `    CheckParty(0, 19, Scratch1, Scratch1, Scratch1, Scratch1)
    if Scratch1 == 1 { goto TrapFound }
`, New: `    CheckParty(0, 19, Scratch1, Scratch1, Scratch1, Scratch1)
    if Scratch1 != 0 { goto TrapFound } # the count of members who find traps; was found == 1
`},
	// area 20 TryRunning: the original's movement census wrote found, always 0, as its fourth result, into Scratch1.
	// Why: pool's CheckParty writes the count for every game (S-diffs 12); Scratch1 is set back to the original's 0, the same branches.
	// CHANGES: none (no behaviour change).
	{Area: 20, Old: `    CheckParty(SelMove, 0, Scratch9, Scratch1, Scratch1, Scratch1)
    if AuxEncResult < Scratch1 {
`, New: `    CheckParty(SelMove, 0, Scratch9, Scratch1, Scratch1, Scratch1)
    Scratch1 = 0 # the original's fourth result after a census (found, never set); the count is the party size
    if AuxEncResult < Scratch1 {
`},
}
