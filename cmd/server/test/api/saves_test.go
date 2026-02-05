package api_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"pubkey-quest/cmd/server/api"
	"pubkey-quest/cmd/server/db"
	"pubkey-quest/cmd/server/test"
)

func setupSavesTestServer(t *testing.T) *test.TestServer {
	t.Helper()

	// Set up test environment (changes to project root for database access)
	test.SetupTestEnvironment(t)

	// Initialize database for tests
	if err := db.InitDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	ts := test.NewTestServer()
	ts.Mux.HandleFunc("/api/saves/", api.SavesHandler)

	return ts
}

// cleanupTestSaves removes any test save files created during tests
func cleanupTestSaves(t *testing.T, npub string) {
	t.Helper()
	savesDir := filepath.Join(api.SavesDirectory, npub)
	os.RemoveAll(savesDir)
}

func TestSavesHandler_GetSaves(t *testing.T) {
	ts := setupSavesTestServer(t)
	defer ts.Close()
	defer db.Close()
	defer cleanupTestSaves(t, test.MockNpub)

	t.Run("returns empty array for new user", func(t *testing.T) {
		resp := ts.GET(t, "/api/saves/"+test.MockNpub)
		test.AssertStatus(t, resp, http.StatusOK)

		saves := test.AssertJSONArray(t, resp)
		if len(saves) != 0 {
			t.Errorf("Expected empty saves array, got %d saves", len(saves))
		}
	})

	t.Run("requires npub in URL", func(t *testing.T) {
		resp := ts.GET(t, "/api/saves/")
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})
}

func TestSavesHandler_CreateSave(t *testing.T) {
	ts := setupSavesTestServer(t)
	defer ts.Close()
	defer db.Close()
	defer cleanupTestSaves(t, test.MockNpub)

	t.Run("creates new save", func(t *testing.T) {
		saveData := test.CreateMockSaveData()
		resp := ts.POST(t, "/api/saves/"+test.MockNpub, saveData)
		test.AssertStatus(t, resp, http.StatusOK)

		result := test.AssertJSON(t, resp)
		test.AssertSuccess(t, result)

		if _, ok := result["save_id"].(string); !ok {
			t.Error("Expected save_id in response")
		}
	})

	t.Run("overwrites existing save with same ID", func(t *testing.T) {
		// First create a save
		saveData := test.CreateMockSaveData()
		resp := ts.POST(t, "/api/saves/"+test.MockNpub, saveData)
		test.AssertStatus(t, resp, http.StatusOK)

		result := test.AssertJSON(t, resp)
		saveID := result["save_id"].(string)

		// Now update it with the same ID
		saveData["id"] = saveID
		saveData["d"] = "UpdatedCharacter"
		resp2 := ts.POST(t, "/api/saves/"+test.MockNpub, saveData)
		test.AssertStatus(t, resp2, http.StatusOK)

		result2 := test.AssertJSON(t, resp2)
		test.AssertSuccess(t, result2)

		// Verify it's the same save ID
		if result2["save_id"] != saveID {
			t.Errorf("Expected same save_id, got %v", result2["save_id"])
		}
	})

	t.Run("rejects invalid save data", func(t *testing.T) {
		resp := ts.POST(t, "/api/saves/"+test.MockNpub, "invalid json")
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})
}

func TestSavesHandler_DeleteSave(t *testing.T) {
	ts := setupSavesTestServer(t)
	defer ts.Close()
	defer db.Close()
	defer cleanupTestSaves(t, test.MockNpub)

	t.Run("requires save ID", func(t *testing.T) {
		resp := ts.DELETE(t, "/api/saves/"+test.MockNpub)
		test.AssertStatus(t, resp, http.StatusBadRequest)
	})

	t.Run("returns 404 for non-existent save", func(t *testing.T) {
		resp := ts.DELETE(t, "/api/saves/"+test.MockNpub+"/nonexistent_save")
		test.AssertStatus(t, resp, http.StatusNotFound)
	})

	t.Run("deletes existing save", func(t *testing.T) {
		// First create a save
		saveData := test.CreateMockSaveData()
		createResp := ts.POST(t, "/api/saves/"+test.MockNpub, saveData)
		test.AssertStatus(t, createResp, http.StatusOK)

		result := test.AssertJSON(t, createResp)
		saveID := result["save_id"].(string)

		// Now delete it
		deleteResp := ts.DELETE(t, "/api/saves/"+test.MockNpub+"/"+saveID)
		test.AssertStatus(t, deleteResp, http.StatusOK)

		deleteResult := test.AssertJSON(t, deleteResp)
		test.AssertSuccess(t, deleteResult)

		// Verify it's deleted by trying to delete again
		deleteResp2 := ts.DELETE(t, "/api/saves/"+test.MockNpub+"/"+saveID)
		test.AssertStatus(t, deleteResp2, http.StatusNotFound)
	})
}

func TestSavesHandler_FullCycle(t *testing.T) {
	ts := setupSavesTestServer(t)
	defer ts.Close()
	defer db.Close()
	defer cleanupTestSaves(t, test.MockNpub)

	// 1. Start with no saves
	resp := ts.GET(t, "/api/saves/"+test.MockNpub)
	test.AssertStatus(t, resp, http.StatusOK)
	saves := test.AssertJSONArray(t, resp)
	if len(saves) != 0 {
		t.Fatalf("Expected 0 saves, got %d", len(saves))
	}

	// 2. Create a save
	saveData := test.CreateMockSaveData()
	resp = ts.POST(t, "/api/saves/"+test.MockNpub, saveData)
	test.AssertStatus(t, resp, http.StatusOK)
	result := test.AssertJSON(t, resp)
	saveID := result["save_id"].(string)

	// 3. Verify save exists
	resp = ts.GET(t, "/api/saves/"+test.MockNpub)
	test.AssertStatus(t, resp, http.StatusOK)
	saves = test.AssertJSONArray(t, resp)
	if len(saves) != 1 {
		t.Fatalf("Expected 1 save, got %d", len(saves))
	}

	// 4. Delete the save
	resp = ts.DELETE(t, "/api/saves/"+test.MockNpub+"/"+saveID)
	test.AssertStatus(t, resp, http.StatusOK)

	// 5. Verify save is deleted
	resp = ts.GET(t, "/api/saves/"+test.MockNpub)
	test.AssertStatus(t, resp, http.StatusOK)
	saves = test.AssertJSONArray(t, resp)
	if len(saves) != 0 {
		t.Fatalf("Expected 0 saves after delete, got %d", len(saves))
	}
}
