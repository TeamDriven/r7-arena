// Copyright 2023 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the referee interface.

var websocket;
let refCommitted = false;
let controlsAvailable = false;
let scoreIsReady = false;
let isPostMatch = false;
let isPlayoff = false;

const parseFoulInput = function (selector) {
  const value = parseInt($(selector).val(), 10);
  return Number.isFinite(value) && value >= 0 ? value : 0;
};

const sendFoulPoints = function () {
  websocket.send("foulPoints", {
    RedFoulPointsAgainst: parseFoulInput("#redFoulPointsAgainst"),
    BlueFoulPointsAgainst: parseFoulInput("#blueFoulPointsAgainst"),
  });
};

const setInputValue = function (selector, value) {
  const input = $(selector);
  if (!input.is(":focus")) {
    input.val(value || 0);
  }
};

var cycleCard = function (cardButton) {
  if (isPostMatch) {
    if (!controlsAvailable || refCommitted) {
      return;
    }

    const currentCard = $(cardButton).attr("data-card");
    const hasOldYellowCard = $(cardButton).attr("data-old-yellow-card") === "true";
    let newCard = "";
    if (currentCard === "" && hasOldYellowCard) {
      newCard = "red";
    } else if (currentCard === "") {
      newCard = "yellow";
    } else if (currentCard === "yellow") {
      newCard = "red";
    }
    websocket.send(
      "card",
      {Alliance: $(cardButton).attr("data-alliance"), TeamId: isPlayoff ? 0 : parseInt($(cardButton).attr("data-team")), Card: newCard}
    );
    $(cardButton).attr("data-card", newCard);
    return;
  }

  const isDisabled = $(cardButton).hasClass("bypassed-status");
  const team = $(cardButton).attr("data-team");
  $("#confirmBypassTitle").text(`${isDisabled ? "Enable" : "Disable"} ${team}?`);
  $("#confirmBypassAction").text(isDisabled ? "Enable" : "Disable");
  $("#confirmBypass").attr("data-station", $(cardButton).attr("data-station")?.toUpperCase());

  if (team === "0") {
    toggleBypass();
  } else {
    $("#confirmBypass").modal("show");
  }
};

const toggleBypass = function () {
  const station = $("#confirmBypass").attr("data-station");
  websocket.send("toggleBypass", station);
};

var signalVolunteers = function () {
  websocket.send("signalVolunteers");
};

var signalReset = function () {
  websocket.send("signalReset");
};

var confirmCommit = function () {
	if (!controlsAvailable || refCommitted) {
		return;
	}
	if (scoreIsReady) {
		commitAndPost();
		return;
	}

	$("#confirmCommit").modal("show");
};

var commitAndPost = function () {
	sendFoulPoints();
	websocket.send("commitAndPost");
};

var handleMatchLoad = function (data) {
  $("#matchName").text(data.Match.LongName);
  isPlayoff = data.Match.Type === matchTypePlayoff;
  for (const alliance of ["red", "blue"]) {
    const color = alliance === "red" ? "Red" : "Blue";
    $(`#${alliance}AllianceCard`).text(`Alliance ${data.Match[`Playoff${color}Alliance`]}`)
      .attr("data-old-yellow-card", data[`Playoff${color}AllianceYellowCard`])
      .attr("data-card", "");
  }
  updateCardControls();

  setTeamCard("red", 1, data.Teams["R1"], data.PlayoffRedAllianceYellowCard);
  setTeamCard("red", 2, data.Teams["R2"], data.PlayoffRedAllianceYellowCard);
  setTeamCard("red", 3, data.Teams["R3"], data.PlayoffRedAllianceYellowCard);
  setTeamCard("blue", 1, data.Teams["B1"], data.PlayoffBlueAllianceYellowCard);
  setTeamCard("blue", 2, data.Teams["B2"], data.PlayoffBlueAllianceYellowCard);
  setTeamCard("blue", 3, data.Teams["B3"], data.PlayoffBlueAllianceYellowCard);
};

const handleMatchTime = function (data) {
	isPostMatch = matchStates[data.MatchState] === "POST_MATCH";
	updateCardControls();
	controlsAvailable = isPostMatch;
	if (!controlsAvailable) {
		refCommitted = false;
	}

	let title = "Red/Yellow Cards";
	if (!isPostMatch) {
		title = matchStates[data.MatchState] === "PRE_MATCH" ? "Bypass" : "Disable";
	}
	$("#teamTitle").text(title);
	updateUIMode();
};

const updateCardControls = function () {
  $(".playoff-alliance-card").toggle(isPlayoff && isPostMatch);
  $(".station-card").toggle(!isPlayoff || !isPostMatch);
};

const handleRealtimeScore = function (data) {
  if (isPlayoff) {
    $("#redAllianceCard").attr("data-card", data.PlayoffRedAllianceCard || "");
    $("#blueAllianceCard").attr("data-card", data.PlayoffBlueAllianceCard || "");
  } else {
    $(".station-card").attr("data-card", "");
    for (const [teamId, card] of Object.entries({...data.RedCards, ...data.BlueCards})) {
      $(`[data-team="${teamId}"]`).attr("data-card", card);
    }
  }

  setInputValue("#redFoulPointsAgainst", data.Red.Score.FoulPointsAgainst);
  setInputValue("#blueFoulPointsAgainst", data.Blue.Score.FoulPointsAgainst);
};

const handleScoringStatus = function (data) {
	refCommitted = data.RefereeScoreReady;
	scoreIsReady = data.ScoringReady;
	$("#scoringPanelStatus").text(`Scoring ${data.NumScoringPanelsReady}/${data.NumScoringPanels}`);
	$("#scoringPanelStatus").attr("data-present", data.NumScoringPanels > 0);
	$("#scoringPanelStatus").attr("data-ready", data.ScoringReady);
	$("#commitButton").toggleClass("disabled", !scoreIsReady);
	updateUIMode();
};

const handleArenaStatus = function (data) {
  setTeamBypassedStatus("red1", data.AllianceStations["R1"]?.Bypass);
  setTeamBypassedStatus("red2", data.AllianceStations["R2"]?.Bypass);
  setTeamBypassedStatus("red3", data.AllianceStations["R3"]?.Bypass);
  setTeamBypassedStatus("blue1", data.AllianceStations["B1"]?.Bypass);
  setTeamBypassedStatus("blue2", data.AllianceStations["B2"]?.Bypass);
  setTeamBypassedStatus("blue3", data.AllianceStations["B3"]?.Bypass);
};

const setTeamBypassedStatus = function (station, bypassed) {
  const cardButton = $(`#${station}Card`);
  cardButton.toggleClass("bypassed-status", bypassed && !isPostMatch);
};

const updateUIMode = function () {
	$("input[type=number]").prop("disabled", false);
	$(".control-button").attr("data-enabled", controlsAvailable);
	$("#commitButton").attr("data-enabled", controlsAvailable && !refCommitted);
};

const setTeamCard = function (alliance, position, team, allianceYellowCard) {
  const cardButton = $(`#${alliance}${position}Card`);
  if (team === null) {
    cardButton.text("-");
    cardButton.attr("data-team", 0);
    cardButton.attr("data-old-yellow-card", "");
  } else {
    cardButton.text(team.Id);
    cardButton.attr("data-team", team.Id);
    cardButton.attr("data-old-yellow-card", isPlayoff ? allianceYellowCard : team.YellowCard);
  }
  cardButton.attr("data-card", "");
};

$(function () {
  $("input[type=number]").on("input", sendFoulPoints);

  websocket = new CheesyWebsocket("/panels/referee/websocket", {
    matchLoad: function (event) {
      handleMatchLoad(event.data);
    },
    matchTime: function (event) {
      handleMatchTime(event.data);
    },
    realtimeScore: function (event) {
      handleRealtimeScore(event.data);
    },
    scoringStatus: function (event) {
      handleScoringStatus(event.data);
    },
    arenaStatus: function (event) {
      handleArenaStatus(event.data);
    },
  });
});
