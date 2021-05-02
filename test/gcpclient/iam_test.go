package gcpclient

import "testing"

// TestGenerateTemporaryAccessToken ensures that privileged service account
// tokens are properly generated.
func TestGenerateTemporaryAccessToken(t *testing.T) {
	// Test that token is for the requested service account

	// Test that the token expires in ~600 seconds

	// Test that attempting to generate access token for a service account
	// that you don't have access to fails
}

// TestGetServiceAccounts ensures that listing the service accounts that you
// have access to impersonate works properly
func TestGetServiceAccounts(t *testing.T) {
}

// TestCanImpersonate ensures that validation of the users ability to impersonate
// a given service account works properly
func TestCanImpersonate(t *testing.T) {
	// Test that the function returns true when the user can impersonate the service account

	// Test that the function returns false when the user cannot impersonate the service account
}
