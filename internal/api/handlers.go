package api

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/fleet"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/history"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/verification"
)

type SingleResponse struct {
	Data any `json:"data"`
}

type CollectionResponse struct {
	Data any            `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) handleListFleets(w http.ResponseWriter, r *http.Request) {
	pagination := parsePagination(r)
	fleets, total, err := s.fleetService.ListFleets(r.Context(), pagination.Limit, pagination.Offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list fleets")
		return
	}

	if fleets == nil {
		fleets = []fleet.Fleet{}
	}

	writeJSON(w, http.StatusOK, CollectionResponse{
		Data: fleets,
		Meta: PaginationMeta{
			Total:  total,
			Limit:  pagination.Limit,
			Offset: pagination.Offset,
		},
	})
}

func (s *Server) handleGetFleet(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	tag := r.PathValue("tag")

	flt, err := s.fleetService.GetFleet(r.Context(), fleet.FleetID{Owner: owner, Tag: tag})
	if err != nil {
		if errors.Is(err, fleet.ErrFleetNotFound) {
			writeError(w, http.StatusNotFound, "FLEET_NOT_FOUND", "Fleet was not found")
			return
		}
		if errors.Is(err, cap85.ErrInvalidContractID) || errors.Is(err, cap85.ErrInvalidTag) {
			writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get fleet")
		return
	}

	writeJSON(w, http.StatusOK, SingleResponse{Data: flt})
}

func (s *Server) handleListMembers(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	tag := r.PathValue("tag")

	activeOnly := true
	if aStr := r.URL.Query().Get("active_only"); aStr != "" {
		if strings.ToLower(aStr) == "false" || aStr == "0" {
			activeOnly = false
		}
	}

	pagination := parsePagination(r)
	members, total, err := s.fleetService.ListMembers(r.Context(), fleet.FleetID{Owner: owner, Tag: tag}, activeOnly, pagination.Limit, pagination.Offset)
	if err != nil {
		if errors.Is(err, cap85.ErrInvalidContractID) || errors.Is(err, cap85.ErrInvalidTag) {
			writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list members")
		return
	}

	if members == nil {
		members = []fleet.FleetMember{}
	}

	writeJSON(w, http.StatusOK, CollectionResponse{
		Data: members,
		Meta: PaginationMeta{
			Total:  total,
			Limit:  pagination.Limit,
			Offset: pagination.Offset,
		},
	})
}

func (s *Server) handleFleetHistory(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	tag := r.PathValue("tag")

	fleetID := fleet.FleetID{Owner: owner, Tag: tag}
	pagination := parsePagination(r)

	historyList, total, err := s.verifier.ListVerifications(r.Context(), fleetID, pagination.Limit, pagination.Offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list verification history")
		return
	}

	if historyList == nil {
		historyList = []verification.VerificationResult{}
	}

	writeJSON(w, http.StatusOK, CollectionResponse{
		Data: historyList,
		Meta: PaginationMeta{
			Total:  total,
			Limit:  pagination.Limit,
			Offset: pagination.Offset,
		},
	})
}

func (s *Server) handleFleetReleases(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	tag := r.PathValue("tag")

	fleetID := history.FleetID{Owner: owner, Tag: tag}
	pagination := parsePagination(r)

	releases, total, err := s.releases.ListReleases(r.Context(), fleetID, pagination.Limit, pagination.Offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list releases")
		return
	}

	if releases == nil {
		releases = []history.FleetRelease{}
	}

	writeJSON(w, http.StatusOK, CollectionResponse{
		Data: releases,
		Meta: PaginationMeta{
			Total:  total,
			Limit:  pagination.Limit,
			Offset: pagination.Offset,
		},
	})
}

func (s *Server) handleVerifyFleet(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	tag := r.PathValue("tag")

	fleetID := fleet.FleetID{Owner: owner, Tag: tag}
	flt, err := s.fleetService.GetFleet(r.Context(), fleetID)
	if err != nil {
		if errors.Is(err, fleet.ErrFleetNotFound) {
			writeError(w, http.StatusNotFound, "FLEET_NOT_FOUND", "Fleet was not found")
			return
		}
		if errors.Is(err, cap85.ErrInvalidContractID) || errors.Is(err, cap85.ErrInvalidTag) {
			writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query fleet")
		return
	}

	var expectedHash []byte
	if hashStr := r.URL.Query().Get("expected_wasm"); hashStr != "" {
		h, err := hex.DecodeString(hashStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "Invalid expected_wasm hex string")
			return
		}
		expectedHash = h
	} else {
		expectedHash = flt.CurrentWASMHash
	}

	if len(expectedHash) == 0 {
		writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", "expected_wasm parameter is required when fleet has no indexed current WASM hash")
		return
	}

	result, err := s.verifier.VerifyFleet(r.Context(), fleetID, expectedHash)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify fleet")
		return
	}

	writeJSON(w, http.StatusOK, SingleResponse{Data: result})
}

func (s *Server) handleInspectContract(w http.ResponseWriter, r *http.Request) {
	contractID := r.PathValue("contract_id")

	inspection, err := s.fleetService.InspectContract(r.Context(), contractID)
	if err != nil {
		if errors.Is(err, cap85.ErrInvalidContractID) {
			writeError(w, http.StatusBadRequest, "INVALID_ARGUMENT", err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to inspect contract")
		return
	}

	writeJSON(w, http.StatusOK, SingleResponse{Data: inspection})
}

type HealthResponse struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Timestamp string `json:"timestamp"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:    "UP",
		Service:   "soroban-fleet-registry",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}
