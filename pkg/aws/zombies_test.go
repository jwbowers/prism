package aws

import (
	"math"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func zombieApproxEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

// ── isPrismManaged ────────────────────────────────────────────────────────

func TestIsPrismManaged(t *testing.T) {
	m := &Manager{}

	tests := []struct {
		name string
		tags []ec2types.Tag
		want bool
	}{
		{
			name: "new namespaced tag",
			tags: []ec2types.Tag{
				{Key: aws.String("prism:managed"), Value: aws.String("true")},
			},
			want: true,
		},
		{
			name: "legacy tag",
			tags: []ec2types.Tag{
				{Key: aws.String("Prism"), Value: aws.String("true")},
			},
			want: true,
		},
		{
			name: "both tags present",
			tags: []ec2types.Tag{
				{Key: aws.String("prism:managed"), Value: aws.String("true")},
				{Key: aws.String("Prism"), Value: aws.String("true")},
				{Key: aws.String("Name"), Value: aws.String("research-1")},
			},
			want: true,
		},
		{
			name: "tag present but wrong value",
			tags: []ec2types.Tag{
				{Key: aws.String("prism:managed"), Value: aws.String("false")},
			},
			want: false,
		},
		{
			name: "unrelated tags only",
			tags: []ec2types.Tag{
				{Key: aws.String("Name"), Value: aws.String("my-instance")},
				{Key: aws.String("Environment"), Value: aws.String("prod")},
			},
			want: false,
		},
		{
			name: "empty tags",
			tags: []ec2types.Tag{},
			want: false,
		},
		{
			name: "nil tags",
			tags: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.isPrismManaged(tt.tags)
			if got != tt.want {
				t.Errorf("isPrismManaged() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ── estimateInstanceMonthlyCost ───────────────────────────────────────────

func TestEstimateInstanceMonthlyCost(t *testing.T) {
	// Known instance types — verify hourly × 24 × 30
	knownTypes := map[string]float64{
		"t3.micro":   0.010,
		"t3.small":   0.021,
		"t3.medium":  0.042,
		"t3.large":   0.083,
		"t3.xlarge":  0.166,
		"t3.2xlarge": 0.333,
		"m5.large":   0.096,
		"m5.xlarge":  0.192,
		"r5.large":   0.126,
		"r5.xlarge":  0.252,
		"c7g.large":  0.072,
		"c7g.xlarge": 0.145,
	}
	for instanceType, hourly := range knownTypes {
		want := hourly * 24 * 30
		got := estimateInstanceMonthlyCost(instanceType)
		if !zombieApproxEqual(got, want) {
			t.Errorf("%s: got %.9f, want %.9f", instanceType, got, want)
		}
	}
}

func TestEstimateInstanceMonthlyCost_Unknown(t *testing.T) {
	// Unknown types fall back to $0.10/hr
	unknown := []string{"p4d.24xlarge", "trn1.2xlarge", "inf2.xlarge", ""}
	want := 0.10 * 24 * 30
	for _, instanceType := range unknown {
		got := estimateInstanceMonthlyCost(instanceType)
		if !zombieApproxEqual(got, want) {
			t.Errorf("%q: got %.9f, want %.9f (default)", instanceType, got, want)
		}
	}
}

func TestEstimateInstanceMonthlyCost_Positive(t *testing.T) {
	// All estimates should be positive
	types := []string{"t3.micro", "m5.4xlarge", "r5.4xlarge", "c7g.8xlarge", "unknown-type"}
	for _, instanceType := range types {
		got := estimateInstanceMonthlyCost(instanceType)
		if got <= 0 {
			t.Errorf("%s: cost should be positive, got %.4f", instanceType, got)
		}
	}
}
