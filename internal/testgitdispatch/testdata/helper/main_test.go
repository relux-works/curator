package main

import "testing"

// The fixture refuses both admitted SCP spellings, without mistaking Windows
// drive paths or a colon in a local path segment for network endpoints.
func TestEndpointClassificationPreservesLocalPaths(t *testing.T) {
	for _, value := range []string{"host.invalid:repo", "git@host.invalid:repo", "https://host.invalid/repo", "ssh://host.invalid/repo"} {
		if !networkEndpoint(value) {
			t.Errorf("network endpoint passed as local: %q", value)
		}
	}
	for _, value := range []string{`C:\fixture\repo`, "C:/fixture/repo", "file:///C:/fixture/repo", "/tmp/fixture:repo", "./fixture:repo"} {
		if networkEndpoint(value) {
			t.Errorf("local path treated as network: %q", value)
		}
	}
}
