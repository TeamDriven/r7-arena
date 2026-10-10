// Copyright 2026 Team 254. All Rights Reserved.

package field

import (
	"github.com/FRCTeam1987/r7-arena/game"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewSCC(t *testing.T) {
	arena := setupTestArena(t)
	scc := NewSCC(arena)

	assert.NotNil(t, scc)
	assert.False(t, scc.IsSccConnected("red"))
	assert.False(t, scc.IsSccConnected("blue"))
	assert.False(t, scc.IsSccConnected("scoring"))
	assert.False(t, scc.IsSccConnected("unknown"))

	notifier := scc.GenerateNotifierStatus()
	assert.False(t, notifier.RedConnected)
	assert.False(t, notifier.BlueConnected)
	assert.False(t, notifier.ScoringConnected)
	assert.False(t, notifier.RedEstop1)
	assert.False(t, notifier.RedEstop2)
	assert.False(t, notifier.RedEstop3)
	assert.False(t, notifier.BlueEstop1)
	assert.False(t, notifier.BlueEstop2)
	assert.False(t, notifier.BlueEstop3)
	assert.False(t, notifier.ScoringEstop)
}

func TestSCC_ApplyUpdate(t *testing.T) {
	arena := setupTestArena(t)
	scc := arena.Scc

	// Apply red update
	scc.ApplyUpdate(
		SCCUpdate{
			Alliance: "red",
			EStops:   []bool{true, false, false},
		},
	)
	assert.True(t, scc.IsSccConnected("red"))
	assert.True(t, arena.AllianceStations["R1"].EStop)
	assert.False(t, arena.AllianceStations["R2"].EStop)
	assert.False(t, arena.AllianceStations["R3"].EStop)

	// Apply blue update
	scc.ApplyUpdate(
		SCCUpdate{
			Alliance: "blue",
			EStops:   []bool{false, true, false},
		},
	)
	assert.True(t, scc.IsSccConnected("blue"))
	assert.False(t, arena.AllianceStations["B1"].EStop)
	assert.True(t, arena.AllianceStations["B2"].EStop)
	assert.False(t, arena.AllianceStations["B3"].EStop)

	notifier := scc.GenerateNotifierStatus()
	assert.True(t, notifier.RedConnected)
	assert.True(t, notifier.RedEstop1)
	assert.False(t, notifier.RedEstop2)
	assert.True(t, notifier.BlueConnected)
	assert.False(t, notifier.BlueEstop1)
	assert.True(t, notifier.BlueEstop2)

	// Clear stops
	scc.ApplyUpdate(
		SCCUpdate{
			Alliance: "red",
			EStops:   []bool{false, false, false},
		},
	)
	scc.ApplyUpdate(
		SCCUpdate{
			Alliance: "blue",
			EStops:   []bool{false, false, false},
		},
	)
	assert.False(t, arena.AllianceStations["R1"].EStop)
	assert.False(t, arena.AllianceStations["B2"].EStop)
}

func TestSCC_ScoringEStopAbortsMatch(t *testing.T) {
	arena := setupTestArena(t)
	scc := arena.Scc

	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true

	assert.Nil(t, arena.StartMatch())
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)

	// Trigger scoring e-stop
	scc.ApplyUpdate(
		SCCUpdate{
			Alliance: "scoring",
			EStops:   []bool{true, false, false},
		},
	)

	assert.True(t, scc.IsSccConnected("scoring"))
	assert.True(t, scc.GenerateNotifierStatus().ScoringEstop)
	assert.Equal(t, PostMatch, arena.MatchState)
}

func TestSCC_Disconnect(t *testing.T) {
	arena := setupTestArena(t)
	scc := arena.Scc

	scc.ApplyUpdate(
		SCCUpdate{
			Alliance: "red",
			EStops:   []bool{true, false, false},
		},
	)
	assert.True(t, scc.IsSccConnected("red"))
	assert.True(t, arena.AllianceStations["R1"].EStop)

	scc.Disconnect("red")
	assert.False(t, scc.IsSccConnected("red"))
	assert.False(t, arena.AllianceStations["R1"].EStop)

	scc.ApplyUpdate(
		SCCUpdate{
			Alliance: "blue",
			EStops:   []bool{false, false, true},
		},
	)
	assert.True(t, scc.IsSccConnected("blue"))
	assert.True(t, arena.AllianceStations["B3"].EStop)

	scc.Disconnect("blue")
	assert.False(t, scc.IsSccConnected("blue"))
	assert.False(t, arena.AllianceStations["B3"].EStop)
}

func TestSCC_AutonomousAStop(t *testing.T) {
	arena := setupTestArena(t)
	scc := arena.Scc

	arena.AllianceStations["R1"].Bypass = true
	arena.AllianceStations["R2"].Bypass = true
	arena.AllianceStations["R3"].Bypass = true
	arena.AllianceStations["B1"].Bypass = true
	arena.AllianceStations["B2"].Bypass = true
	arena.AllianceStations["B3"].Bypass = true

	assert.Nil(t, arena.StartMatch())
	arena.Update()
	assert.Equal(t, AutoPeriod, arena.MatchState)

	// Press stop during auto -> should set AStop
	scc.ApplyUpdate(
		SCCUpdate{
			Alliance: "red",
			EStops:   []bool{true, false, false},
		},
	)
	assert.True(t, arena.AllianceStations["R1"].AStop)
	assert.False(t, arena.AllianceStations["R1"].EStop)

	// Advance to teleop -> AStop remains, but unpressed stops do not prevent teleop enable
	arena.MatchStartTime = time.Now().Add(-game.GetDurationToTeleopStart() - time.Second)
	arena.Update() // Auto -> Pause
	arena.Update() // Pause -> Teleop
	assert.Equal(t, TeleopPeriod, arena.MatchState)
	assert.False(t, arena.AllianceStations["R1"].EStop)

	// Press stop during teleop -> should set EStop
	scc.ApplyUpdate(
		SCCUpdate{
			Alliance: "red",
			EStops:   []bool{true, false, false},
		},
	)
	assert.True(t, arena.AllianceStations["R1"].EStop)
}
