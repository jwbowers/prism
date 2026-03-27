package daemon

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/scttfrdmn/prism/pkg/types"
	"github.com/scttfrdmn/prism/pkg/workshop"
)

// handleWorkshopOperations routes /api/v1/workshops (no trailing segment)
func (s *Server) handleWorkshopOperations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListWorkshops(w, r)
	case http.MethodPost:
		s.handleCreateWorkshop(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleWorkshopByID dispatches /api/v1/workshops/{id}[/{sub-resource}...]
func (s *Server) handleWorkshopByID(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/api/v1/workshops/"):]
	parts := splitPath(path)
	if len(parts) == 0 {
		s.writeError(w, http.StatusBadRequest, "Missing workshop ID")
		return
	}

	// Handle /api/v1/workshops/configs and /api/v1/workshops/from-config/{name}
	if parts[0] == "configs" {
		s.handleWorkshopConfigs(w, r, parts)
		return
	}
	if parts[0] == "from-config" {
		if len(parts) < 2 {
			s.writeError(w, http.StatusBadRequest, "Missing config name")
			return
		}
		if r.Method != http.MethodPost {
			s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		s.handleCreateWorkshopFromConfig(w, r, parts[1])
		return
	}

	workshopID := parts[0]
	if len(parts) == 1 {
		s.handleWorkshopDirectOp(w, r, workshopID)
		return
	}

	switch parts[1] {
	case "provision":
		if r.Method != http.MethodPost {
			s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		s.handleProvisionWorkshop(w, r, workshopID)
	case "dashboard":
		if r.Method != http.MethodGet {
			s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		s.handleWorkshopDashboard(w, r, workshopID)
	case "end":
		if r.Method != http.MethodPost {
			s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		s.handleEndWorkshop(w, r, workshopID)
	case "download":
		if r.Method != http.MethodGet {
			s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		s.handleWorkshopDownload(w, r, workshopID)
	case "participants":
		s.handleWorkshopParticipants(w, r, workshopID, parts)
	case "config":
		// POST /api/v1/workshops/{id}/config — save config from this workshop
		if r.Method != http.MethodPost {
			s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}
		s.handleSaveWorkshopConfig(w, r, workshopID)
	default:
		http.NotFound(w, r)
	}
}

// handleWorkshopDirectOp handles GET/PUT/DELETE on /api/v1/workshops/{id}
func (s *Server) handleWorkshopDirectOp(w http.ResponseWriter, r *http.Request, workshopID string) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetWorkshop(w, r, workshopID)
	case http.MethodPut:
		s.handleUpdateWorkshop(w, r, workshopID)
	case http.MethodDelete:
		s.handleDeleteWorkshop(w, r, workshopID)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// ── CRUD ──────────────────────────────────────────────────────────────────────

func (s *Server) handleListWorkshops(w http.ResponseWriter, r *http.Request) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}

	q := r.URL.Query()
	filter := &workshop.WorkshopFilter{
		Owner: q.Get("owner"),
	}
	if statusStr := q.Get("status"); statusStr != "" {
		st := workshop.WorkshopStatus(statusStr)
		filter.Status = &st
	}

	workshops, err := s.workshopManager.ListWorkshops(r.Context(), filter)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if workshops == nil {
		workshops = []*workshop.WorkshopEvent{}
	}
	s.writeJSON(w, http.StatusOK, map[string]interface{}{"workshops": workshops})
}

func (s *Server) handleCreateWorkshop(w http.ResponseWriter, r *http.Request) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}

	var req workshop.CreateWorkshopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	wk, err := s.workshopManager.CreateWorkshop(r.Context(), &req)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusCreated, wk)
}

func (s *Server) handleGetWorkshop(w http.ResponseWriter, r *http.Request, workshopID string) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}

	wk, err := s.workshopManager.GetWorkshop(r.Context(), workshopID)
	if err != nil {
		if err == workshop.ErrWorkshopNotFound {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, wk)
}

func (s *Server) handleUpdateWorkshop(w http.ResponseWriter, r *http.Request, workshopID string) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}

	var req workshop.UpdateWorkshopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	wk, err := s.workshopManager.UpdateWorkshop(r.Context(), workshopID, &req)
	if err != nil {
		if err == workshop.ErrWorkshopNotFound {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, wk)
}

func (s *Server) handleDeleteWorkshop(w http.ResponseWriter, r *http.Request, workshopID string) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}

	if err := s.workshopManager.DeleteWorkshop(r.Context(), workshopID); err != nil {
		if err == workshop.ErrWorkshopNotFound {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ── Lifecycle ─────────────────────────────────────────────────────────────────

func (s *Server) handleProvisionWorkshop(w http.ResponseWriter, r *http.Request, workshopID string) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}

	result, toProvision, err := s.workshopManager.ProvisionWorkshop(r.Context(), workshopID)
	if err != nil {
		if err == workshop.ErrWorkshopNotFound {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// In test mode or when no AWS manager, mark participants as provisioned without launching
	if s.testMode || s.awsManager == nil {
		for _, p := range toProvision {
			_ = s.workshopManager.UpdateParticipantInstance(r.Context(), workshopID,
				p.UserID, "mock-"+p.UserID, "ws-"+p.UserID, "provisioned")
			result.Provisioned++
		}
		s.writeJSON(w, http.StatusOK, result)
		return
	}

	// Get workshop for template info
	wk, err := s.workshopManager.GetWorkshop(r.Context(), workshopID)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Launch instances for pending participants
	hoursUntilEnd := int(time.Until(wk.EndTime).Hours()) + 1
	for _, p := range toProvision {
		instanceName := "ws-" + workshopID[:8] + "-" + p.UserID
		if len(instanceName) > 30 {
			instanceName = instanceName[:30]
		}

		launchReq := types.LaunchRequest{
			Template: wk.Template,
			Name:     instanceName,
			Hours:    hoursUntilEnd,
		}
		_, launchErr := s.awsManager.LaunchInstanceWithContext(r.Context(), launchReq)
		if launchErr != nil {
			result.Errors = append(result.Errors, "participant "+p.UserID+": "+launchErr.Error())
			continue
		}
		_ = s.workshopManager.UpdateParticipantInstance(r.Context(), workshopID,
			p.UserID, "", instanceName, "provisioned")
		result.Provisioned++
	}

	s.writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleWorkshopDashboard(w http.ResponseWriter, r *http.Request, workshopID string) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}

	dash, err := s.workshopManager.GetDashboard(r.Context(), workshopID)
	if err != nil {
		if err == workshop.ErrWorkshopNotFound {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, dash)
}

func (s *Server) handleEndWorkshop(w http.ResponseWriter, r *http.Request, workshopID string) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}

	result, withInstances, err := s.workshopManager.EndWorkshop(r.Context(), workshopID)
	if err != nil {
		if err == workshop.ErrWorkshopNotFound {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Stop all participant instances (best-effort)
	if s.awsManager != nil && !s.testMode {
		for _, p := range withInstances {
			if p.InstanceName != "" {
				if stopErr := s.awsManager.StopInstance(p.InstanceName); stopErr != nil {
					result.Errors = append(result.Errors, "stop "+p.InstanceName+": "+stopErr.Error())
				} else {
					result.Stopped++
				}
			}
		}
	} else {
		result.Stopped = len(withInstances)
	}

	s.writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleWorkshopDownload(w http.ResponseWriter, r *http.Request, workshopID string) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}

	manifest, err := s.workshopManager.GetDownloadManifest(r.Context(), workshopID)
	if err != nil {
		if err == workshop.ErrWorkshopNotFound {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, manifest)
}

// ── Participants ──────────────────────────────────────────────────────────────

func (s *Server) handleWorkshopParticipants(w http.ResponseWriter, r *http.Request, workshopID string, parts []string) {
	if len(parts) == 2 {
		// /api/v1/workshops/{id}/participants
		switch r.Method {
		case http.MethodGet:
			s.handleListWorkshopParticipants(w, r, workshopID)
		case http.MethodPost:
			s.handleAddWorkshopParticipant(w, r, workshopID)
		default:
			s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}
	// /api/v1/workshops/{id}/participants/{userID}
	userID := parts[2]
	if r.Method == http.MethodDelete {
		s.handleRemoveWorkshopParticipant(w, r, workshopID, userID)
	} else {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleListWorkshopParticipants(w http.ResponseWriter, r *http.Request, workshopID string) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}
	wk, err := s.workshopManager.GetWorkshop(r.Context(), workshopID)
	if err != nil {
		if err == workshop.ErrWorkshopNotFound {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]interface{}{"participants": wk.Participants})
}

func (s *Server) handleAddWorkshopParticipant(w http.ResponseWriter, r *http.Request, workshopID string) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}

	var p workshop.WorkshopParticipant
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}
	if p.UserID == "" {
		p.UserID = p.Email
	}

	wk, err := s.workshopManager.AddParticipant(r.Context(), workshopID, p)
	if err != nil {
		switch err {
		case workshop.ErrWorkshopNotFound:
			s.writeError(w, http.StatusNotFound, err.Error())
		case workshop.ErrWorkshopFull:
			s.writeError(w, http.StatusConflict, err.Error())
		case workshop.ErrWorkshopEnded:
			s.writeError(w, http.StatusGone, err.Error())
		default:
			s.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	s.writeJSON(w, http.StatusCreated, wk)
}

func (s *Server) handleRemoveWorkshopParticipant(w http.ResponseWriter, r *http.Request, workshopID, userID string) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}

	if err := s.workshopManager.RemoveParticipant(r.Context(), workshopID, userID); err != nil {
		switch err {
		case workshop.ErrWorkshopNotFound:
			s.writeError(w, http.StatusNotFound, err.Error())
		case workshop.ErrParticipantNotFound:
			s.writeError(w, http.StatusNotFound, err.Error())
		default:
			s.writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	s.writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

// ── Config Templates ──────────────────────────────────────────────────────────

func (s *Server) handleWorkshopConfigs(w http.ResponseWriter, r *http.Request, parts []string) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}
	// GET /api/v1/workshops/configs
	if len(parts) == 1 && r.Method == http.MethodGet {
		configs, err := s.workshopManager.ListConfigs(r.Context())
		if err != nil {
			s.writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if configs == nil {
			configs = []*workshop.WorkshopConfig{}
		}
		s.writeJSON(w, http.StatusOK, map[string]interface{}{"configs": configs})
		return
	}
	s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
}

func (s *Server) handleSaveWorkshopConfig(w http.ResponseWriter, r *http.Request, workshopID string) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}
	if body.Name == "" {
		s.writeError(w, http.StatusBadRequest, "config name is required")
		return
	}

	config, err := s.workshopManager.SaveConfig(r.Context(), workshopID, body.Name)
	if err != nil {
		if err == workshop.ErrWorkshopNotFound {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.writeJSON(w, http.StatusCreated, config)
}

func (s *Server) handleCreateWorkshopFromConfig(w http.ResponseWriter, r *http.Request, configName string) {
	if s.workshopManager == nil {
		s.writeError(w, http.StatusServiceUnavailable, "Workshop manager unavailable")
		return
	}

	var req workshop.CreateWorkshopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	wk, err := s.workshopManager.CreateFromConfig(r.Context(), configName, &req)
	if err != nil {
		if err == workshop.ErrConfigNotFound {
			s.writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.writeJSON(w, http.StatusCreated, wk)
}
