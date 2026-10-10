// Copyright 2020 Ken Schenke. All Rights Reserved.
// Author: kenschenke@gmail.com (Ken Schenke)
//
// Web routes for SCC interactions.

package web

import (
	"fmt"
	"github.com/FRCTeam1987/r7-arena/field"
	"github.com/FRCTeam1987/r7-arena/model"
	"github.com/FRCTeam1987/r7-arena/websocket"
	"io"
	"log"
	"net/http"
)

// Shows the SCC Testing page.
func (web *Web) sccGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	template, err := web.parseFiles("templates/setup_scc.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	data := struct {
		*model.EventSettings
		field.SCCNotifier
	}{
		web.arena.EventSettings,
		web.arena.Scc.GenerateNotifierStatus(),
	}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// The websocket endpoint for getting realtime updates from the SCC boxes.
func (web *Web) sccWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer func(ws *websocket.Websocket) {
		err := ws.Close()
		if err != nil {

		}
	}(ws)

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	alliance := ""

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, data, err := ws.Read()
		if err != nil {
			if len(alliance) > 0 {
				web.arena.Scc.Disconnect(alliance)
			}
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		switch messageType {
		case "sccupdate":
			update := data.(map[string]interface{})
			alliance = ""
			eStop1 := false
			eStop2 := false
			eStop3 := false
			var ok bool
			if alliance, ok = update["alliance"].(string); !ok {
				log.Println("Missing alliance string")
				err := ws.WriteError("Missing alliance string")
				if err != nil {
					return
				}
				continue
			}
			if eStop1, ok = update["eStop1"].(bool); !ok {
				log.Println("Missing eStop1 boolean")
				err := ws.WriteError("Missing eStop1 boolean")
				if err != nil {
					return
				}
				continue
			}
			if eStop2, ok = update["eStop2"].(bool); !ok {
				log.Println("Missing eStop2 boolean")
				err := ws.WriteError("Missing eStop2 boolean")
				if err != nil {
					return
				}
				continue
			}
			if eStop3, ok = update["eStop3"].(bool); !ok {
				log.Println("Missing eStop3 boolean")
				err := ws.WriteError("Missing eStop3 boolean")
				if err != nil {
					return
				}
				continue
			}
			web.arena.Scc.ApplyUpdate(
				field.SCCUpdate{
					Alliance: alliance,
					EStops:   []bool{eStop1, eStop2, eStop3},
				},
			)
		default:
			err := ws.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			if err != nil {
				return
			}
			continue
		}
	}
}

// The websocket endpoint for the scc testing page.
func (web *Web) sccGetTestingWebsocketHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := websocket.NewWebsocket(w, r)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer func(ws *websocket.Websocket) {
		err := ws.Close()
		if err != nil {

		}
	}(ws)

	// Subscribe the websocket to the notifiers whose messages will be passed on to the client, in a separate goroutine.
	go ws.HandleNotifiers(web.arena.SCCNotifier)

	// Loop, waiting for commands and responding to them, until the client closes the connection.
	for {
		messageType, _, err := ws.Read()
		if err != nil {
			if err == io.EOF {
				// Client has closed the connection; nothing to do here.
				return
			}
			log.Println(err)
			return
		}

		switch messageType {
		default:
			err := ws.WriteError(fmt.Sprintf("Invalid message type '%s'.", messageType))
			if err != nil {
				return
			}
			continue
		}
	}
}
