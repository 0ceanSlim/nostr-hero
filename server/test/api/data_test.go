package api_test

import (
	"net/http"
	"testing"

	"nostr-hero/api/data"
	"nostr-hero/db"
	"nostr-hero/test"
)

func setupDataTestServer(t *testing.T) *test.TestServer {
	t.Helper()

	// Set up test environment (changes to project root for database access)
	test.SetupTestEnvironment(t)

	// Initialize database for tests
	if err := db.InitDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	ts := test.NewTestServer()
	ts.Mux.HandleFunc("/api/game-data", data.GameDataHandler)
	ts.Mux.HandleFunc("/api/items", data.ItemsHandler)
	ts.Mux.HandleFunc("/api/spells/", data.SpellsHandler)
	ts.Mux.HandleFunc("/api/monsters", data.MonstersHandler)
	ts.Mux.HandleFunc("/api/locations", data.LocationsHandler)
	ts.Mux.HandleFunc("/api/npcs", data.NPCsHandler)
	ts.Mux.HandleFunc("/api/npcs/at-location", data.GetNPCsAtLocationHandler)
	ts.Mux.HandleFunc("/api/abilities", data.AbilitiesHandler)

	return ts
}

func TestGameDataHandler(t *testing.T) {
	ts := setupDataTestServer(t)
	defer ts.Close()
	defer db.Close()

	resp := ts.GET(t, "/api/game-data")
	test.AssertStatus(t, resp, http.StatusOK)

	var result map[string]interface{}
	test.ReadJSON(t, resp, &result)

	// Check that all expected fields are present
	requiredFields := []string{"items", "spells", "monsters", "locations", "packs", "music_tracks"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("Expected field %s in game-data response", field)
		}
	}

	// Check that arrays are not empty
	if items, ok := result["items"].([]interface{}); ok {
		if len(items) == 0 {
			t.Error("Expected items array to be non-empty")
		}
	} else {
		t.Error("Items should be an array")
	}
}

func TestItemsHandler(t *testing.T) {
	ts := setupDataTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("get all items", func(t *testing.T) {
		resp := ts.GET(t, "/api/items")
		test.AssertStatus(t, resp, http.StatusOK)

		items := test.AssertJSONArray(t, resp)
		if len(items) == 0 {
			t.Error("Expected items array to be non-empty")
		}
	})

	t.Run("filter by name", func(t *testing.T) {
		resp := ts.GET(t, "/api/items?name=longsword")
		test.AssertStatus(t, resp, http.StatusOK)

		items := test.AssertJSONArray(t, resp)
		// May be empty if longsword doesn't exist, but shouldn't error
		if len(items) > 0 {
			item := items[0].(map[string]interface{})
			if item["id"] != "longsword" {
				t.Errorf("Expected item ID 'longsword', got %v", item["id"])
			}
		}
	})
}

func TestSpellsHandler(t *testing.T) {
	ts := setupDataTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("get all spells", func(t *testing.T) {
		resp := ts.GET(t, "/api/spells/")
		test.AssertStatus(t, resp, http.StatusOK)

		spells := test.AssertJSONArray(t, resp)
		if len(spells) == 0 {
			t.Error("Expected spells array to be non-empty")
		}
	})

	t.Run("get spell by ID", func(t *testing.T) {
		resp := ts.GET(t, "/api/spells/fire-bolt")
		// May be 404 if fire-bolt doesn't exist, or 200 if it does
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Errorf("Unexpected status code: %d", resp.StatusCode)
		}
	})
}

func TestMonstersHandler(t *testing.T) {
	ts := setupDataTestServer(t)
	defer ts.Close()
	defer db.Close()

	resp := ts.GET(t, "/api/monsters")
	test.AssertStatus(t, resp, http.StatusOK)

	monsters := test.AssertJSONArray(t, resp)
	// May be empty if no monsters defined
	_ = monsters
}

func TestLocationsHandler(t *testing.T) {
	ts := setupDataTestServer(t)
	defer ts.Close()
	defer db.Close()

	resp := ts.GET(t, "/api/locations")
	test.AssertStatus(t, resp, http.StatusOK)

	locations := test.AssertJSONArray(t, resp)
	if len(locations) == 0 {
		t.Error("Expected locations array to be non-empty")
	}
}

func TestNPCsHandler(t *testing.T) {
	ts := setupDataTestServer(t)
	defer ts.Close()
	defer db.Close()

	resp := ts.GET(t, "/api/npcs")
	test.AssertStatus(t, resp, http.StatusOK)

	npcs := test.AssertJSONArray(t, resp)
	// NPCs should exist
	_ = npcs
}

func TestGetNPCsAtLocationHandler(t *testing.T) {
	ts := setupDataTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("requires GET method", func(t *testing.T) {
		resp := ts.POST(t, "/api/npcs/at-location", nil)
		test.AssertStatus(t, resp, http.StatusMethodNotAllowed)
	})

	t.Run("with location parameters", func(t *testing.T) {
		resp := ts.GET(t, "/api/npcs/at-location?location=kingdom&district=kingdom-center&time=720")
		test.AssertStatus(t, resp, http.StatusOK)

		npcs := test.AssertJSONArray(t, resp)
		// May be empty depending on NPC schedules
		_ = npcs
	})
}

func TestAbilitiesHandler(t *testing.T) {
	ts := setupDataTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("requires class parameter", func(t *testing.T) {
		resp := ts.GET(t, "/api/abilities")
		test.AssertStatus(t, resp, http.StatusBadRequest)

		result := test.AssertJSON(t, resp)
		test.AssertError(t, result)
	})

	t.Run("rejects invalid class", func(t *testing.T) {
		resp := ts.GET(t, "/api/abilities?class=wizard")
		test.AssertStatus(t, resp, http.StatusBadRequest)

		result := test.AssertJSON(t, resp)
		test.AssertError(t, result)
	})

	t.Run("accepts valid martial class", func(t *testing.T) {
		resp := ts.GET(t, "/api/abilities?class=fighter&level=5")
		test.AssertStatus(t, resp, http.StatusOK)

		result := test.AssertJSON(t, resp)
		test.AssertSuccess(t, result)

		if result["class"] != "fighter" {
			t.Errorf("Expected class 'fighter', got %v", result["class"])
		}
	})
}
