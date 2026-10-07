package api

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/fleet"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/history"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/verification"
	"github.com/soroban-fleet/soroban-fleet-registry/migrations"
	"github.com/stellar/go-stellar-sdk/strkey"
)

func getTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("SFR_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres@localhost:5433/sfr_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("cannot connect to db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("db not reachable: %v", err)
	}
	if err := migrations.Up(db); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}
	return db
}

func makeContractID(seed byte) string {
	var b [32]byte
	b[0] = seed
	res, _ := strkey.Encode(strkey.VersionByteContract, b[:])
	return res
}

func TestAPI_Endpoints(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE fleet_verifications, fleet_releases, fleet_members, fleets CASCADE")

	fleetRepo := fleet.NewPostgresRepository(db)
	releaseRepo := history.NewPostgresRepository(db)
	verifier := verification.NewEngine(db, fleetRepo, nil, nil, nil)
	fleetService := fleet.NewService(fleetRepo, nil)

	server := NewServer(":8080", fleetService, verifier, releaseRepo, nil)
	handler := server.Handler()

	ownerID := makeContractID(1)
	tag := "vault-v1"
	fID := fleet.FleetID{Owner: ownerID, Tag: tag}
	wasmBytes := make([]byte, 32)
	wasmBytes[0] = 0xAA

	// 1. Seed fleet, member, and release
	_ = fleetRepo.UpsertFleet(context.Background(), &fleet.Fleet{
		ID:                fID,
		CurrentWASMHash:   wasmBytes,
		FirstSeenLedger:   100,
		LastSeenLedger:    100,
		LastIndexedLedger: 100,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	})

	cid := makeContractID(2)
	_ = fleetRepo.UpsertMember(context.Background(), &fleet.FleetMember{
		ContractID:      cid,
		FleetID:         fID,
		WASMHash:        wasmBytes,
		FirstSeenLedger: 100,
		LastSeenLedger:  100,
		Active:          true,
	})

	tx, _ := db.BeginTx(context.Background(), nil)
	_ = releaseRepo.RecordReleaseTx(context.Background(), tx, ownerID, tag, nil, wasmBytes, 100, "tx1", time.Now().UTC())
	_ = tx.Commit()

	// Test 1: GET /v1/fleets
	req := httptest.NewRequest("GET", "/v1/fleets?limit=10&offset=0", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var coll CollectionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &coll); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if coll.Meta.Total != 1 {
		t.Errorf("expected total 1, got %d", coll.Meta.Total)
	}

	// Test 2: GET /v1/fleets/{owner}/{tag}
	req = httptest.NewRequest("GET", "/v1/fleets/"+ownerID+"/"+tag, nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Test 3: GET /v1/fleets/{owner}/{tag} - NOT FOUND
	req = httptest.NewRequest("GET", "/v1/fleets/"+ownerID+"/nonexistent", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	var errResp ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &errResp)
	if errResp.Error.Code != "FLEET_NOT_FOUND" {
		t.Errorf("expected FLEET_NOT_FOUND error code, got %s", errResp.Error.Code)
	}

	// Test 4: GET /v1/fleets/{owner}/{tag}/members
	req = httptest.NewRequest("GET", "/v1/fleets/"+ownerID+"/"+tag+"/members?active_only=true", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Test 5: GET /v1/fleets/{owner}/{tag}/releases
	req = httptest.NewRequest("GET", "/v1/fleets/"+ownerID+"/"+tag+"/releases", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Test 6: GET /v1/fleets/{owner}/{tag}/verify
	expectedHex := hex.EncodeToString(wasmBytes)
	req = httptest.NewRequest("GET", "/v1/fleets/"+ownerID+"/"+tag+"/verify?expected_wasm="+expectedHex, nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Test 7: GET /v1/fleets/{owner}/{tag}/history
	req = httptest.NewRequest("GET", "/v1/fleets/"+ownerID+"/"+tag+"/history", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Test 8: GET /v1/contracts/{contract_id}
	req = httptest.NewRequest("GET", "/v1/contracts/"+cid, nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Test 9: GET /v1/contracts/{contract_id} with invalid strkey -> 400
	req = httptest.NewRequest("GET", "/v1/contracts/invalid_id", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid contract ID, got %d", rec.Code)
	}
	var errResp2 ErrorResponse
	_ = json.Unmarshal(rec.Body.Bytes(), &errResp2)
	if errResp2.Error.Code != "INVALID_ARGUMENT" {
		t.Errorf("expected INVALID_ARGUMENT, got %s", errResp2.Error.Code)
	}
}
