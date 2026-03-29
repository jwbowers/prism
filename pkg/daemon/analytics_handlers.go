package daemon

// analytics_handlers.go — Storage analytics REST handlers for Prism daemon.
//
// Routes:
//   GET  /api/v1/storage/analytics            getAllStorageAnalytics [?period=daily|weekly|monthly]
//   GET  /api/v1/storage/analytics/{name}     getStorageAnalytics   [?period=]
//
// Issue #23 / sub-issue 23b

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	awspkg "github.com/scttfrdmn/prism/pkg/aws"
	"github.com/scttfrdmn/prism/pkg/storage"
)

// handleStorageAnalytics dispatches /api/v1/storage/analytics and sub-paths.
// Called from handleStorageOperations when the path contains "analytics".
func (s *Server) handleStorageAnalytics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse optional resource name from path: /api/v1/storage/analytics/{name}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/storage/analytics")
	path = strings.TrimPrefix(path, "/")
	resourceName := strings.TrimSuffix(path, "/")

	period := storagePeriodFromQuery(r)

	s.withAWSManager(w, r, func(m *awspkg.Manager) error {
		mgr := storage.NewAnalyticsManager(m.GetAWSConfig())

		if resourceName != "" {
			return s.getOneStorageAnalytics(w, r, mgr, resourceName, period)
		}
		return s.getAllStorageAnalytics(w, r, mgr, period)
	})
}

func (s *Server) getAllStorageAnalytics(w http.ResponseWriter, r *http.Request, mgr *storage.AnalyticsManager, period storage.AnalyticsPeriod) error {
	start, end := periodToTimeRange(period)
	req := storage.AnalyticsRequest{
		Period:    period,
		StartTime: start,
		EndTime:   end,
	}

	// Populate resources from daemon state
	st, err := s.stateManager.LoadState()
	if err != nil {
		return err
	}
	for name, vol := range st.StorageVolumes {
		resource := storage.StorageResource{Name: name}
		if vol.FileSystemID != "" {
			resource.Type = storage.StorageTypeEFS
			resource.ResourceID = vol.FileSystemID
		} else if vol.VolumeID != "" {
			resource.Type = storage.StorageTypeEBS
			resource.ResourceID = vol.VolumeID
		}
		req.Resources = append(req.Resources, resource)
	}

	analysis, err := mgr.GetStorageCostAnalysis(req)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(analysis)
}

func (s *Server) getOneStorageAnalytics(w http.ResponseWriter, r *http.Request, mgr *storage.AnalyticsManager, name string, period storage.AnalyticsPeriod) error {
	start, end := periodToTimeRange(period)
	req := storage.AnalyticsRequest{
		Period:    period,
		StartTime: start,
		EndTime:   end,
	}

	// Try to resolve resource ID from state
	st, err := s.stateManager.LoadState()
	if err != nil {
		return err
	}
	if vol, ok := st.StorageVolumes[name]; ok {
		if vol.FileSystemID != "" {
			req.Resources = []storage.StorageResource{{Name: name, Type: storage.StorageTypeEFS, ResourceID: vol.FileSystemID}}
		} else {
			req.Resources = []storage.StorageResource{{Name: name, Type: storage.StorageTypeEBS, ResourceID: vol.VolumeID}}
		}
	}

	analysis, err := mgr.GetStorageCostAnalysis(req)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(analysis)
}

// storagePeriodFromQuery extracts ?period= from the request; defaults to daily.
func storagePeriodFromQuery(r *http.Request) storage.AnalyticsPeriod {
	switch r.URL.Query().Get("period") {
	case "weekly":
		return storage.AnalyticsPeriodWeekly
	case "monthly":
		return storage.AnalyticsPeriodMonthly
	default:
		return storage.AnalyticsPeriodDaily
	}
}

// periodToTimeRange converts an AnalyticsPeriod to a [start, end) window ending now.
func periodToTimeRange(period storage.AnalyticsPeriod) (time.Time, time.Time) {
	end := time.Now().UTC()
	switch period {
	case storage.AnalyticsPeriodWeekly:
		return end.AddDate(0, 0, -7), end
	case storage.AnalyticsPeriodMonthly:
		return end.AddDate(0, -1, 0), end
	default:
		return end.AddDate(0, 0, -1), end
	}
}
