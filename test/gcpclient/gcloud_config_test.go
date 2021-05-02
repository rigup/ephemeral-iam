package gcpclient

import "testing"

// TestCheckActiveAccountSet ensures that an error is returned when the user does
// not have an active account set in their gcloud config
func TestCheckActiveAccountSet(t *testing.T) {
}

// TestGetCurrentProject ensures that the project configured in the user's
// gcloud config is fetched properly
func TestGetCurrentProject(t *testing.T) {
}

// TestGetCurrentRegion ensures that the region configured in the user's
// gcloud config is fetched properly
func TestGetCurrentRegion(t *testing.T) {
}

// TestGetCurrentZone ensures that the zone configured in the user's
// gcloud config is fetched properly
func TestGetCurrentZone(t *testing.T) {
}

// TestConfigureGcloudProxy ensures that the user's gcloud config is properly
// configured to point to the ephemeral-iam auth proxy
func TestConfigureGcloudProxy(t *testing.T) {
}

// TestUnsetGcloudProxy ensures that the user's gcloud config is returned to normal
// after being configured to point to the ephemeral-iam auth proxy
func TestUnsetGcloudProxy(t *testing.T) {
}
