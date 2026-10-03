// Copyright 2020 Ken Schenke. All Rights Reserved.
// Author: kenschenke@gmail.com (Ken Schenke)
//
// Functions for interacting with SCC units

package field

import (
	"fmt"
)

type SCCStatus struct {
	Connected bool   `json:"connected"`
	EStops    []bool `json:"eStops"`
}

type SCCUpdate struct {
	Alliance string `json:"alliance"`
	EStops   []bool `json:"eStops"`
}

type SCC struct {
	status map[string]*SCCStatus
	arena  *Arena
}

type SCCNotifier struct {
	RedConnected     bool
	RedEstop1        bool
	RedEstop2        bool
	RedEstop3        bool
	BlueConnected    bool
	BlueEstop1       bool
	BlueEstop2       bool
	BlueEstop3       bool
	ScoringConnected bool
	ScoringEstop     bool
}

func NewSCC(arena *Arena) *SCC {
	scc := new(SCC)
	scc.arena = arena
	scc.status = make(map[string]*SCCStatus)

	red := new(SCCStatus)
	red.EStops = []bool{false, false, false}

	blue := new(SCCStatus)
	blue.EStops = []bool{false, false, false}

	scoring := new(SCCStatus)
	scoring.EStops = []bool{false, false, false}

	scc.status["red"] = red
	scc.status["blue"] = blue
	scc.status["scoring"] = scoring

	return scc
}

func (scc *SCC) ApplyUpdate(update SCCUpdate) {
	status, ok := scc.status[update.Alliance]
	if ok {
		alliance := "R"
		if update.Alliance == "blue" {
			alliance = "B"
		} else if update.Alliance == "scoring" {
			alliance = "S"
		}

		updated := false
		if !status.Connected {
			updated = true
		}
		for i := range status.EStops {
			if status.EStops[i] != update.EStops[i] {
				updated = true
			}
		}

		status.Connected = true
		status.EStops = update.EStops

		if alliance == "R" || alliance == "B" {
			scc.updateEstop(alliance, 1, update.EStops[0])
			scc.updateEstop(alliance, 2, update.EStops[1])
			scc.updateEstop(alliance, 3, update.EStops[2])
		} else if update.EStops[0] {
			scc.arena.AbortMatch()
		}

		if updated {
			scc.arena.SCCNotifier.Notify()
		}
	}
}

func (scc *SCC) updateEstop(alliance string, station int, newValue bool) {
	code := fmt.Sprintf("%s%d", alliance, station)
	if scc.arena.AllianceStations[code].Estop == false || newValue {
		scc.arena.handleEstop(code, newValue)
	}
}

func (scc *SCC) Disconnect(alliance string) {
	status, ok := scc.status[alliance]
	if ok {
		status.Connected = false
		// Mark all 3 eStops as off
		code := "R"
		if alliance == "blue" {
			code = "B"
		}
		status.EStops[0] = false
		status.EStops[1] = false
		status.EStops[2] = false
		scc.updateEstop(code, 1, false)
		scc.updateEstop(code, 2, false)
		scc.updateEstop(code, 3, false)
		scc.arena.SCCNotifier.Notify()
	}
}

func (scc *SCC) GenerateNotifierStatus() SCCNotifier {
	return SCCNotifier{
		scc.status["red"].Connected,
		scc.status["red"].EStops[0],
		scc.status["red"].EStops[1],
		scc.status["red"].EStops[2],
		scc.status["blue"].Connected,
		scc.status["blue"].EStops[0],
		scc.status["blue"].EStops[1],
		scc.status["blue"].EStops[2],
		scc.status["scoring"].Connected,
		scc.status["scoring"].EStops[0],
	}
}

func (scc *SCC) IsSccConnected(station string) bool {
	status, ok := scc.status[station]
	if ok {
		return status.Connected
	}
	return false
}
