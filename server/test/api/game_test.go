package api_test

import (
	"net/http"
	"testing"

	"nostr-hero/api/game"
	"nostr-hero/db"
	"nostr-hero/test"
)

func setupGameTestServer(t *testing.T) *test.TestServer {
	t.Helper()

	// Set up test environment (changes to project root for database access)
	test.SetupTestEnvironment(t)

	// Initialize database for tests
	if err := db.InitDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	ts := test.NewTestServer()
	ts.Mux.HandleFunc("/api/session/init", game.InitSessionHandler)
	ts.Mux.HandleFunc("/api/session/reload", game.ReloadSessionHandler)
	ts.Mux.HandleFunc("/api/session/state", game.GetSessionHandler)
	ts.Mux.HandleFunc("/api/session/update", game.UpdateSessionHandler)
	ts.Mux.HandleFunc("/api/session/save", game.SaveSessionHandler)
	ts.Mux.HandleFunc("/api/session/cleanup", game.CleanupSessionHandler)
	ts.Mux.HandleFunc("/api/game/action", game.GameActionHandler)
	ts.Mux.HandleFunc("/api/game/state", game.GetGameStateHandler)
	ts.Mux.HandleFunc("/api/shop/", game.ShopHandler)

	return ts
}

func TestInitSessionHandler(t *testing.T) {
	ts := setupGameTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("requires POST method", func(t *testing.T) {
		resp := ts.GET(t, "/api/session/init")
		test.AssertStatus(t, resp, http.StatusMethodNotAllowed)
	})

	t.Run("requires npub and save_id", func(t *testing.T) {
		body := map[string]interface{}{}
		resp := ts.POST(t, "/api/session/init", body)
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("fails for non-existent save", func(t *testing.T) {
		body := map[string]interface{}{
			"npub":    test.MockNpub,
			"save_id": "nonexistent_save",
		}
		resp := ts.POST(t, "/api/session/init", body)
		test.AssertStatus(t, resp, http.StatusInternalServerError)
	})
}

func TestReloadSessionHandler(t *testing.T) {
	ts := setupGameTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("requires POST method", func(t *testing.T) {
		resp := ts.GET(t, "/api/session/reload")
		test.AssertStatus(t, resp, http.StatusMethodNotAllowed)
	})

	t.Run("requires npub and save_id", func(t *testing.T) {
		body := map[string]interface{}{}
		resp := ts.POST(t, "/api/session/reload", body)
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})
}

func TestGetSessionHandler(t *testing.T) {
	ts := setupGameTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("requires GET method", func(t *testing.T) {
		resp := ts.POST(t, "/api/session/state", nil)
		test.AssertStatus(t, resp, http.StatusMethodNotAllowed)
	})

	t.Run("requires npub and save_id", func(t *testing.T) {
		resp := ts.GET(t, "/api/session/state")
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("returns 404 for non-existent session", func(t *testing.T) {
		resp := ts.GET(t, "/api/session/state?npub="+test.MockNpub+"&save_id=nonexistent")
		// Will try to load from disk, which should fail
		test.AssertStatus(t, resp, http.StatusNotFound)
	})
}

func TestUpdateSessionHandler(t *testing.T) {
	ts := setupGameTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("requires POST method", func(t *testing.T) {
		resp := ts.GET(t, "/api/session/update")
		test.AssertStatus(t, resp, http.StatusMethodNotAllowed)
	})

	t.Run("requires npub and save_id", func(t *testing.T) {
		body := map[string]interface{}{
			"save_data": test.CreateMockSaveData(),
		}
		resp := ts.POST(t, "/api/session/update", body)
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})
}

func TestSaveSessionHandler(t *testing.T) {
	ts := setupGameTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("requires POST method", func(t *testing.T) {
		resp := ts.GET(t, "/api/session/save")
		test.AssertStatus(t, resp, http.StatusMethodNotAllowed)
	})

	t.Run("requires npub and save_id", func(t *testing.T) {
		body := map[string]interface{}{}
		resp := ts.POST(t, "/api/session/save", body)
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("fails if session not in memory", func(t *testing.T) {
		body := map[string]interface{}{
			"npub":    test.MockNpub,
			"save_id": "nonexistent",
		}
		resp := ts.POST(t, "/api/session/save", body)
		test.AssertStatus(t, resp, http.StatusNotFound)
	})
}

func TestCleanupSessionHandler(t *testing.T) {
	ts := setupGameTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("requires DELETE method", func(t *testing.T) {
		resp := ts.GET(t, "/api/session/cleanup")
		test.AssertStatus(t, resp, http.StatusMethodNotAllowed)
	})

	t.Run("requires npub and save_id", func(t *testing.T) {
		resp := ts.DELETE(t, "/api/session/cleanup")
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("succeeds even for non-existent session", func(t *testing.T) {
		resp := ts.DELETE(t, "/api/session/cleanup?npub="+test.MockNpub+"&save_id=nonexistent")
		test.AssertStatus(t, resp, http.StatusOK)

		result := test.AssertJSON(t, resp)
		test.AssertSuccess(t, result)
	})
}

func TestGameActionHandler(t *testing.T) {
	ts := setupGameTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("requires POST method", func(t *testing.T) {
		resp := ts.GET(t, "/api/game/action")
		test.AssertStatus(t, resp, http.StatusMethodNotAllowed)
	})

	t.Run("requires npub and save_id", func(t *testing.T) {
		body := map[string]interface{}{
			"action": map[string]interface{}{
				"type":   "move",
				"params": map[string]interface{}{},
			},
		}
		resp := ts.POST(t, "/api/game/action", body)
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("fails for non-existent session", func(t *testing.T) {
		body := map[string]interface{}{
			"npub":    test.MockNpub,
			"save_id": "nonexistent",
			"action": map[string]interface{}{
				"type":   "move",
				"params": map[string]interface{}{},
			},
		}
		resp := ts.POST(t, "/api/game/action", body)
		test.AssertStatus(t, resp, http.StatusNotFound)
	})
}

func TestGetGameStateHandler(t *testing.T) {
	ts := setupGameTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("requires GET method", func(t *testing.T) {
		resp := ts.POST(t, "/api/game/state", nil)
		test.AssertStatus(t, resp, http.StatusMethodNotAllowed)
	})

	t.Run("requires npub and save_id", func(t *testing.T) {
		resp := ts.GET(t, "/api/game/state")
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})
}

func TestShopHandler(t *testing.T) {
	ts := setupGameTestServer(t)
	defer ts.Close()
	defer db.Close()

	t.Run("GET requires merchant ID", func(t *testing.T) {
		resp := ts.GET(t, "/api/shop/")
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("GET requires npub and save_id", func(t *testing.T) {
		resp := ts.GET(t, "/api/shop/some-merchant")
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("POST buy requires valid body", func(t *testing.T) {
		resp := ts.POST(t, "/api/shop/buy", "invalid")
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("POST sell requires valid body", func(t *testing.T) {
		resp := ts.POST(t, "/api/shop/sell", "invalid")
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})
}
