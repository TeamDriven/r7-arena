// Copyright 2026 Team 254. All Rights Reserved.

package web

import (
	"github.com/TeamDriven/r7-arena/field"
	"github.com/TeamDriven/r7-arena/websocket"
	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestSetupSCC(t *testing.T) {
	web := setupTestWeb(t)

	// Admin access should succeed and display the SCC page.
	recorder := web.getHttpResponse("/setup/scc")
	assert.Equal(t, 200, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "SCC Testing")
	assert.Contains(t, recorder.Body.String(), "Red SCC")
	assert.Contains(t, recorder.Body.String(), "Blue SCC")
	assert.Contains(t, recorder.Body.String(), "Scoring Table SCC")
}

func TestSCCWebsocketHandler_SuccessAndDisconnect(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()

	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/scc/websocket", nil)
	assert.Nil(t, err)
	ws := websocket.NewTestWebsocket(conn)

	// Send an SCC update for the red alliance.
	ws.Write("sccupdate", map[string]any{
		"alliance": "red",
		"eStop1":   true,
		"eStop2":   false,
		"eStop3":   false,
	})

	// Wait and verify the arena state was updated.
	assert.Eventually(t, func() bool {
		return web.arena.Scc.IsSccConnected("red") && web.arena.AllianceStations["R1"].EStop
	}, 500*time.Millisecond, 10*time.Millisecond)

	// Close the websocket connection and verify disconnection.
	conn.Close()
	assert.Eventually(t, func() bool {
		return !web.arena.Scc.IsSccConnected("red") && !web.arena.AllianceStations["R1"].EStop
	}, 500*time.Millisecond, 10*time.Millisecond)
}

func TestSCCWebsocketHandler_InvalidMessages(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()

	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/scc/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Invalid message type
	ws.Write("invalidType", map[string]any{})
	assert.Equal(t, "Invalid message type 'invalidType'.", readWebsocketError(t, ws))

	// Missing alliance
	ws.Write("sccupdate", map[string]any{
		"eStop1": true,
		"eStop2": false,
		"eStop3": false,
	})
	assert.Equal(t, "Missing alliance string", readWebsocketError(t, ws))

	// Missing eStop1
	ws.Write("sccupdate", map[string]any{
		"alliance": "red",
		"eStop2":   false,
		"eStop3":   false,
	})
	assert.Equal(t, "Missing eStop1 boolean", readWebsocketError(t, ws))

	// Missing eStop2
	ws.Write("sccupdate", map[string]any{
		"alliance": "red",
		"eStop1":   true,
		"eStop3":   false,
	})
	assert.Equal(t, "Missing eStop2 boolean", readWebsocketError(t, ws))

	// Missing eStop3
	ws.Write("sccupdate", map[string]any{
		"alliance": "red",
		"eStop1":   true,
		"eStop2":   false,
	})
	assert.Equal(t, "Missing eStop3 boolean", readWebsocketError(t, ws))
}

func TestSCCTestingWebsocketHandler(t *testing.T) {
	web := setupTestWeb(t)

	server, wsUrl := web.startTestServer()
	defer server.Close()

	conn, _, err := gorillawebsocket.DefaultDialer.Dial(wsUrl+"/setup/scc/websocket", nil)
	assert.Nil(t, err)
	defer conn.Close()
	ws := websocket.NewTestWebsocket(conn)

	// Consume initial status message sent upon connection.
	_ = readWebsocketType(t, ws, "sccstatus")

	// Trigger an SCC status update and verify notification is received.
	web.arena.Scc.ApplyUpdate(field.SCCUpdate{
		Alliance: "blue",
		EStops:   []bool{false, true, false},
	})

	var sccStatus field.SCCNotifier
	msg := readWebsocketType(t, ws, "sccstatus")
	assert.Nil(t, mapstructure.Decode(msg, &sccStatus))
	assert.True(t, sccStatus.BlueConnected)
	assert.False(t, sccStatus.BlueEstop1)
	assert.True(t, sccStatus.BlueEstop2)
	assert.False(t, sccStatus.BlueEstop3)

	// Sending an unknown message type should return an error.
	ws.Write("unknown", nil)
	assert.Equal(t, "Invalid message type 'unknown'.", readWebsocketError(t, ws))
}
