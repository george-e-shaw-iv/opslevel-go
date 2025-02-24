package opslevel_test

import (
	"testing"

	"github.com/rocktavious/autopilot"
)

func TestListTiers(t *testing.T) {
	// Arrange
	client := ATestClientSkipRequest(t, "tiers")
	// Act
	result, err := client.ListTiers()
	// Assert
	autopilot.Ok(t, err)
	autopilot.Equals(t, 4, len(result))
	autopilot.Equals(t, "tier_1", result[0].Alias)
}
