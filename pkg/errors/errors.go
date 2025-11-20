// Package errors provides enhanced error messages with actionable guidance
//
// This package wraps errors with contextual information, suggestions for
// resolution, and documentation links to improve the developer experience
// when encountering failures.
package errors

import "github.com/scttfrdmn/prism/pkg/aws"

// Enhance wraps an error with actionable guidance and documentation links
//
// This function analyzes the error message and adds contextual help based
// on common error patterns. It delegates to the AWS error enhancement
// system which provides detailed guidance for AWS-specific errors.
//
// Example:
//
//	err := someAWSOperation()
//	enhancedErr := errors.Enhance(err)
//	fmt.Println(enhancedErr) // Shows suggestions and docs links
func Enhance(err error) error {
	return aws.EnhanceError(err, "operation")
}

// EnhanceWithOperation wraps an error with operation-specific context
//
// This variant allows specifying the operation name for better error messages.
//
// Example:
//
//	err := launchInstance()
//	enhancedErr := errors.EnhanceWithOperation(err, "launch instance")
func EnhanceWithOperation(err error, operation string) error {
	return aws.EnhanceError(err, operation)
}
