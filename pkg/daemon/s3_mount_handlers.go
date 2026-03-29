package daemon

// s3_mount_handlers.go — S3 mount/unmount REST handlers for Prism daemon.
//
// Routes:
//   GET    /api/v1/instances/{name}/s3-mounts          listS3Mounts
//   POST   /api/v1/instances/{name}/s3-mounts          mountS3Bucket
//   DELETE /api/v1/instances/{name}/s3-mounts/{path}   unmountS3Bucket
//
// Issue #22 / sub-issue 22b

import (
	"encoding/json"
	"net/http"
	"strings"

	awspkg "github.com/scttfrdmn/prism/pkg/aws"
	"github.com/scttfrdmn/prism/pkg/storage"
)

type s3MountRequest struct {
	BucketName string                `json:"bucket_name"`
	MountPath  string                `json:"mount_path"`
	Method     storage.S3MountMethod `json:"method,omitempty"` // defaults to mountpoint
	ReadOnly   bool                  `json:"read_only,omitempty"`
}

// handleInstanceS3Mounts dispatches /instances/{name}/s3-mounts and sub-paths.
// Called from routeInstanceOperation when parts[1]=="s3-mounts".
func (s *Server) handleInstanceS3Mounts(w http.ResponseWriter, r *http.Request, instanceName string) {
	path := r.URL.Path
	base := "/api/v1/instances/" + instanceName + "/s3-mounts"
	suffix := strings.TrimPrefix(path, base)
	suffix = strings.TrimSuffix(suffix, "/")

	if suffix == "" {
		switch r.Method {
		case http.MethodGet:
			s.handleListS3Mounts(w, r, instanceName)
		case http.MethodPost:
			s.handleMountS3Bucket(w, r, instanceName)
		default:
			s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
		return
	}

	// DELETE /instances/{name}/s3-mounts/{mountPath}
	mountPath := strings.TrimPrefix(suffix, "/")
	if mountPath == "" {
		s.writeError(w, http.StatusBadRequest, "missing mount path")
		return
	}
	if r.Method == http.MethodDelete {
		s.handleUnmountS3Bucket(w, r, instanceName, "/"+mountPath)
	} else {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleListS3Mounts(w http.ResponseWriter, r *http.Request, instanceName string) {
	s.withAWSManager(w, r, func(m *awspkg.Manager) error {
		// Load instance ID from state
		st, err := s.stateManager.LoadState()
		if err != nil {
			return err
		}
		inst, ok := st.Instances[instanceName]
		if !ok {
			return writeNotFound(w, "instance", instanceName)
		}
		s3mgr := storage.NewS3Manager(m.GetAWSConfig())
		mounts, err := s3mgr.ListInstanceS3Mounts(r.Context(), m, inst.ID)
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "application/json")
		return json.NewEncoder(w).Encode(mounts)
	})
}

func (s *Server) handleMountS3Bucket(w http.ResponseWriter, r *http.Request, instanceName string) {
	var req s3MountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.BucketName == "" || req.MountPath == "" {
		s.writeError(w, http.StatusBadRequest, "bucket_name and mount_path are required")
		return
	}
	if req.Method == "" {
		req.Method = storage.S3MountMethodMountpoint
	}

	s.withAWSManager(w, r, func(m *awspkg.Manager) error {
		st, err := s.stateManager.LoadState()
		if err != nil {
			return err
		}
		inst, ok := st.Instances[instanceName]
		if !ok {
			return writeNotFound(w, "instance", instanceName)
		}
		s3mgr := storage.NewS3Manager(m.GetAWSConfig())
		result, err := s3mgr.MountS3BucketOnInstance(r.Context(), m, inst.ID, instanceName,
			req.BucketName, req.MountPath, req.Method, req.ReadOnly)
		if err != nil {
			return err
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		return json.NewEncoder(w).Encode(result)
	})
}

func (s *Server) handleUnmountS3Bucket(w http.ResponseWriter, r *http.Request, instanceName, mountPath string) {
	s.withAWSManager(w, r, func(m *awspkg.Manager) error {
		st, err := s.stateManager.LoadState()
		if err != nil {
			return err
		}
		inst, ok := st.Instances[instanceName]
		if !ok {
			return writeNotFound(w, "instance", instanceName)
		}
		s3mgr := storage.NewS3Manager(m.GetAWSConfig())
		if err := s3mgr.UnmountS3BucketFromInstance(r.Context(), m, inst.ID, mountPath); err != nil {
			return err
		}
		w.WriteHeader(http.StatusNoContent)
		return nil
	})
}

// writeNotFound is a tiny helper — returns an error so withAWSManager writes it.
func writeNotFound(w http.ResponseWriter, kind, name string) error {
	return &notFoundError{kind: kind, name: name}
}

type notFoundError struct{ kind, name string }

func (e *notFoundError) Error() string {
	return e.kind + " '" + e.name + "' not found"
}
