// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the Field Testing page.

var websocket;

// Handles a websocket message to update the SCC status.
var handleUpdate = function(data) {
  if (data.RedConnected) {
    $("#redConnected").addClass("scc-indicator-connected");
  } else {
    $("#redConnected").removeClass("scc-indicator-connected");
  }
  if (data.BlueConnected) {
    $("#blueConnected").addClass("scc-indicator-connected");
  } else {
    $("#blueConnected").removeClass("scc-indicator-connected");
  }
  if (data.ScoringConnected) {
    $("#scoringConnected").addClass("scc-indicator-connected");
  } else {
    $("#scoringConnected").removeClass("scc-indicator-connected");
  }

  if (data.RedEstop1) {
    $("#redEstop1").addClass("scc-indicator-pushed");
  } else {
    $("#redEstop1").removeClass("scc-indicator-pushed");
  }
  if (data.RedEstop2) {
    $("#redEstop2").addClass("scc-indicator-pushed");
  } else {
    $("#redEstop2").removeClass("scc-indicator-pushed");
  }
  if (data.RedEstop3) {
    $("#redEstop3").addClass("scc-indicator-pushed");
  } else {
    $("#redEstop3").removeClass("scc-indicator-pushed");
  }

  if (data.BlueEstop1) {
    $("#blueEstop1").addClass("scc-indicator-pushed");
  } else {
    $("#blueEstop1").removeClass("scc-indicator-pushed");
  }
  if (data.BlueEstop2) {
    $("#blueEstop2").addClass("scc-indicator-pushed");
  } else {
    $("#blueEstop2").removeClass("scc-indicator-pushed");
  }
  if (data.BlueEstop3) {
    $("#blueEstop3").addClass("scc-indicator-pushed");
  } else {
    $("#blueEstop3").removeClass("scc-indicator-pushed");
  }

  if (data.ScoringEstop) {
    $("#scoringEstop").addClass("scc-indicator-pushed");
  } else {
    $("#scoringEstop").removeClass("scc-indicator-pushed");
  }
};

$(function() {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/setup/scc/websocket", {
    sccstatus: function(event) { handleUpdate(event.data); }
  });
});
