package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAvailabilityManager_isStandardAZ(t *testing.T) {
	tests := []struct {
		name     string
		region   string
		az       string
		expected bool
		reason   string
	}{
		// Standard AZs - should return true
		{
			name:     "us-west-2a standard AZ",
			region:   "us-west-2",
			az:       "us-west-2a",
			expected: true,
			reason:   "Standard AZ with single letter suffix",
		},
		{
			name:     "us-west-2b standard AZ",
			region:   "us-west-2",
			az:       "us-west-2b",
			expected: true,
			reason:   "Standard AZ with single letter suffix",
		},
		{
			name:     "us-east-1a standard AZ",
			region:   "us-east-1",
			az:       "us-east-1a",
			expected: true,
			reason:   "Standard AZ in us-east-1",
		},
		{
			name:     "eu-west-1c standard AZ",
			region:   "eu-west-1",
			az:       "eu-west-1c",
			expected: true,
			reason:   "Standard AZ in EU region",
		},
		{
			name:     "ap-southeast-2a standard AZ",
			region:   "ap-southeast-2",
			az:       "ap-southeast-2a",
			expected: true,
			reason:   "Standard AZ in AP region",
		},

		// Local zones - should return false
		{
			name:     "us-west-2-lax-1a local zone",
			region:   "us-west-2",
			az:       "us-west-2-lax-1a",
			expected: false,
			reason:   "Local zone with city code (LAX)",
		},
		{
			name:     "us-east-1-bos-1a local zone",
			region:   "us-east-1",
			az:       "us-east-1-bos-1a",
			expected: false,
			reason:   "Local zone with city code (BOS)",
		},
		{
			name:     "us-west-2-lax-1b local zone",
			region:   "us-west-2",
			az:       "us-west-2-lax-1b",
			expected: false,
			reason:   "Local zone variant B",
		},

		// Wavelength zones - should return false
		{
			name:     "us-east-1-wl1-nyc-wlz-1 wavelength zone",
			region:   "us-east-1",
			az:       "us-east-1-wl1-nyc-wlz-1",
			expected: false,
			reason:   "Wavelength zone with wl identifier",
		},
		{
			name:     "us-west-2-wl1-las-wlz-1 wavelength zone",
			region:   "us-west-2",
			az:       "us-west-2-wl1-las-wlz-1",
			expected: false,
			reason:   "Wavelength zone in us-west-2",
		},

		// Edge cases - should return false
		{
			name:     "wrong region prefix",
			region:   "us-west-2",
			az:       "us-east-1a",
			expected: false,
			reason:   "AZ doesn't match manager's region",
		},
		{
			name:     "empty suffix",
			region:   "us-west-2",
			az:       "us-west-2",
			expected: false,
			reason:   "Missing letter suffix",
		},
		{
			name:     "uppercase letter",
			region:   "us-west-2",
			az:       "us-west-2A",
			expected: false,
			reason:   "Uppercase letter not valid",
		},
		{
			name:     "number suffix",
			region:   "us-west-2",
			az:       "us-west-21",
			expected: false,
			reason:   "Number instead of letter",
		},
		{
			name:     "multi-character suffix",
			region:   "us-west-2",
			az:       "us-west-2ab",
			expected: false,
			reason:   "Multiple characters instead of single letter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am := &AvailabilityManager{
				region: tt.region,
			}

			result := am.isStandardAZ(tt.az)
			assert.Equal(t, tt.expected, result,
				"isStandardAZ(%s) with region=%s: expected %v (reason: %s)",
				tt.az, tt.region, tt.expected, tt.reason)
		})
	}
}

func TestAvailabilityManager_isStandardAZ_AllRegions(t *testing.T) {
	// Test that the function works across different AWS region naming patterns
	regions := []string{
		"us-east-1",      // Standard US region
		"us-west-2",      // Standard US region
		"eu-west-1",      // EU region
		"ap-southeast-2", // AP region with multi-part name
		"ca-central-1",   // Canada region
		"sa-east-1",      // South America
	}

	for _, region := range regions {
		t.Run(region, func(t *testing.T) {
			am := &AvailabilityManager{
				region: region,
			}

			// Test standard AZ (should pass)
			standardAZ := region + "a"
			assert.True(t, am.isStandardAZ(standardAZ),
				"Standard AZ %s should be recognized", standardAZ)

			// Test local zone (should fail)
			localZone := region + "-city-1a"
			assert.False(t, am.isStandardAZ(localZone),
				"Local zone %s should be filtered out", localZone)

			// Test wavelength zone (should fail)
			wavelengthZone := region + "-wl1-city-wlz-1"
			assert.False(t, am.isStandardAZ(wavelengthZone),
				"Wavelength zone %s should be filtered out", wavelengthZone)
		})
	}
}
