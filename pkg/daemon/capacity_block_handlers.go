package daemon

// capacity_block_handlers.go — EC2 Capacity Block REST handlers for Prism daemon.
//
// Routes:
//   POST   /api/v1/capacity-blocks          reserveCapacityBlock
//   GET    /api/v1/capacity-blocks          listCapacityBlocks
//   GET    /api/v1/capacity-blocks/{id}     describeCapacityBlock
//   DELETE /api/v1/capacity-blocks/{id}     cancelCapacityBlock
//
// Issue #63 / sub-issue 63a

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/scttfrdmn/prism/pkg/aws"
)

// handleCapacityBlocks dispatches GET / POST on /api/v1/capacity-blocks
func (s *Server) handleCapacityBlocks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListCapacityBlocks(w, r)
	case http.MethodPost:
		s.handleReserveCapacityBlock(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleCapacityBlockByID dispatches GET / DELETE on /api/v1/capacity-blocks/{id}
func (s *Server) handleCapacityBlockByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /api/v1/capacity-blocks/{id}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/capacity-blocks/")
	id = strings.TrimSuffix(id, "/")
	if id == "" {
		s.writeError(w, http.StatusBadRequest, "missing capacity block ID")
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleDescribeCapacityBlock(w, r, id)
	case http.MethodDelete:
		s.handleCancelCapacityBlock(w, r, id)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleListCapacityBlocks(w http.ResponseWriter, r *http.Request) {
	s.withAWSManager(w, r, func(m *aws.Manager) error {
		blocks, err := m.ListCapacityBlocks(r.Context())
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(blocks)
	})
}

func (s *Server) handleReserveCapacityBlock(w http.ResponseWriter, r *http.Request) {
	var req aws.CapacityBlockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.InstanceType == "" {
		s.writeError(w, http.StatusBadRequest, "instance_type is required")
		return
	}

	s.withAWSManager(w, r, func(m *aws.Manager) error {
		block, err := m.ReserveCapacityBlock(r.Context(), req)
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		return json.NewEncoder(w).Encode(block)
	})
}

func (s *Server) handleDescribeCapacityBlock(w http.ResponseWriter, r *http.Request, id string) {
	s.withAWSManager(w, r, func(m *aws.Manager) error {
		block, err := m.DescribeCapacityBlock(r.Context(), id)
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(block)
	})
}

func (s *Server) handleCancelCapacityBlock(w http.ResponseWriter, r *http.Request, id string) {
	s.withAWSManager(w, r, func(m *aws.Manager) error {
		if err := m.CancelCapacityBlock(r.Context(), id); err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}
