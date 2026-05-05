package team_balancer

import (
	"testing"

	"go.codycody31.dev/squad-aegis/internal/event_manager"
)

func TestDetermineWinnerTeamIDPrefersStructuredWinnerData(t *testing.T) {
	t.Parallel()

	event := &event_manager.LogGameEventUnifiedData{
		Winner: "Eastern Irregular Militia Forces",
	}
	winnerData := map[string]interface{}{
		"team":    "2",
		"faction": "Eastern Irregular Militia Forces",
	}
	cachedTeamNames := map[int]string{
		1: "French Land Army 2020",
		2: "Irregular Motorized Platoon",
	}

	if winnerID := determineWinnerTeamID(event, winnerData, cachedTeamNames); winnerID != 2 {
		t.Fatalf("winnerID = %d, want 2", winnerID)
	}
}

func TestDetermineWinnerTeamIDFallsBackToNumericWinner(t *testing.T) {
	t.Parallel()

	event := &event_manager.LogGameEventUnifiedData{
		Winner: "1",
	}

	if winnerID := determineWinnerTeamID(event, nil, nil); winnerID != 1 {
		t.Fatalf("winnerID = %d, want 1", winnerID)
	}
}

func TestDetermineWinnerTeamIDFallsBackToCachedTeamNames(t *testing.T) {
	t.Parallel()

	event := &event_manager.LogGameEventUnifiedData{
		Winner: "Middle Eastern Insurgents",
	}
	cachedTeamNames := map[int]string{
		1: "Middle Eastern Insurgents",
		2: "United States Marine Corps 2010",
	}

	if winnerID := determineWinnerTeamID(event, nil, cachedTeamNames); winnerID != 1 {
		t.Fatalf("winnerID = %d, want 1", winnerID)
	}
}
