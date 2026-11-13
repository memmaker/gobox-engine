package engine

import (
	"gobox/rules1e"
	"testing"
)

func TestAuditCDeadStaysDead(t *testing.T) {
	e := &Engine{}
	e.initVM()
	c := &rules1e.Character{}
	c.Status = rules1e.StatusDead
	e.state().ApplyDamage(c, 3)
	if c.Status != rules1e.StatusDead {
		t.Fatalf("status %d", c.Status)
	}
}

func TestAuditCFireResistSmall(t *testing.T) {
	e := &Engine{}
	e.initVM()
	e.state().Rv.DmgType, e.state().Rv.DiceN, e.state().Rv.Damage = rules1e.DmgFire|rules1e.DmgMagic, 1, 1
	e.state().ResistFirePerDie(&rules1e.Character{}, &rules1e.Affect{}, 0)
	if e.state().Rv.Damage > 1 {
		t.Fatalf("damage %d", e.state().Rv.Damage)
	}
}

func TestAuditCSpeedNoExtraAttacks(t *testing.T) {
	e := &Engine{}
	e.initVM()
	c := &rules1e.Character{Combat: &rules1e.CombatInfo{Attacked: 1}}
	c.BaseAttack[0], c.Attacks[0] = 2, 0
	e.state().SetSpeed(c)
	if c.Attacks[0] != 0 {
		t.Fatalf("attacks %d", c.Attacks[0])
	}
}
