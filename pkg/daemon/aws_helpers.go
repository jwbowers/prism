package daemon

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/scttfrdmn/prism/pkg/aws"
)

// createAWSManagerFromRequest creates a new AWS manager from the request context
func (s *Server) createAWSManagerFromRequest(r *http.Request) (*aws.Manager, error) {
	// Get AWS credentials from request context
	options := aws.ManagerOptions{}
	hasProfile := false
	hasRegion := false

	if profile := getAWSProfile(r.Context()); profile != "" {
		options.Profile = profile
		hasProfile = true
	}
	if region := getAWSRegion(r.Context()); region != "" {
		options.Region = region
		hasRegion = true
	}

	// If no profile/region in request context, use daemon's initialized manager
	// This allows daemon to respect environment variables (AWS_PROFILE, AWS_REGION)
	// when clients don't send explicit credentials in HTTP headers
	if !hasProfile && !hasRegion {
		if s.awsManager != nil {
			log.Printf("No AWS credentials in request headers, using daemon's initialized manager")
			return s.awsManager, nil
		}
		log.Printf("No AWS credentials in request headers and daemon manager not initialized, using AWS SDK defaults")
	}

	// Create AWS manager with credentials from request headers
	return aws.NewManager(options)
}

// withAWSManager runs the given function with an AWS manager created from the request
// and handles error cases, writing the error to the response
func (s *Server) withAWSManager(w http.ResponseWriter, r *http.Request,
	fn func(*aws.Manager) error) {
	// Track this as an AWS operation specifically with operation name derived from URL
	opType := "AWS-" + extractOperationType(r.URL.Path)
	opID := s.statusTracker.StartOperationWithType(opType)

	// Record start time for duration tracking
	startTime := time.Now()

	// Ensure operation is completed and log duration
	defer func() {
		s.statusTracker.EndOperationWithType(opType)
		log.Printf("AWS operation %d (%s) completed in %v", opID, opType, time.Since(startTime))
	}()

	// Create AWS manager with credentials from the request
	awsManager, err := s.createAWSManagerFromRequest(r)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("Failed to initialize AWS manager: %v", err))
		return
	}

	// Call the provided function with the AWS manager
	if err := fn(awsManager); err != nil {
		s.writeError(w, http.StatusInternalServerError,
			fmt.Sprintf("AWS operation failed: %v", err))
		return
	}
}
