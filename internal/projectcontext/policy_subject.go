package projectcontext

import (
	"os"
	"path/filepath"

	"github.com/relux-works/curator/internal/managerlock"
)

// PolicySubject is a read-only projection of protected registration identity.
// It is never accepted back as a caller assertion of trust.
type PolicySubject struct {
	Registration       Registration
	RegistrationSHA256 string
	RepositoryRoot     string
	PrimaryRemote      string
	RootCommit         string
}

// ReadPolicySubject borrows the mutation lock and requires the literal checkout
// root, current profile/lock and retained filesystem identity. It creates no
// registration, consent, home, lease or projection state.
func (s Store) ReadPolicySubject(held managerlock.Held, checkout, root, environment string, selected Profile) (PolicySubject, error) {
	op, err := s.readHeld(held)
	if err != nil {
		return PolicySubject{}, err
	}
	r, ok := op.state.Registrations[checkout]
	if !ok || !r.Active || r.Lapsed || r.Record.ProjectRoot != root || r.Record.Environment != environment || r.Record.Profile != selected {
		return PolicySubject{}, fail("workspace_untrusted", "no matching current registration")
	}
	if !validRoot(root) {
		return PolicySubject{}, fail("workspace_trust_path_invalid", "select one exact checkout root")
	}
	for p := root; ; p = filepath.Dir(p) {
		st, e := os.Lstat(p)
		if e != nil || !st.IsDir() {
			return PolicySubject{}, fail("workspace_trust_symlink", "checkout path is linked or unavailable")
		}
		if p == filepath.Dir(p) {
			break
		}
	}
	if err = verifyBinding(r); err != nil {
		return PolicySubject{}, fail("workspace_trust_repository_changed", "registration identity changed")
	}
	if shallow, e := gitLocal(root, "rev-parse", "--is-shallow-repository"); e != nil || shallow != "false" {
		return PolicySubject{}, fail("workspace_trust_repository_unverifiable", "complete root history is required")
	}
	p, _, err := profile(s.Home, selected.Name)
	if err != nil || p != selected {
		return PolicySubject{}, fail("project_context_drift", "selected profile lock changed")
	}
	// Preserve the literal configured spelling: normalized aliases never mint
	// destination consent. Wider/sole-remote evaluation is not qualified here.
	remote, err := gitLocal(root, "config", "--get-all", "remote.origin.url")
	if err != nil {
		return PolicySubject{}, fail("workspace_trust_repository_unverifiable", "exact origin remote unavailable")
	}
	if err = held.Validate(s.Home); err != nil {
		return PolicySubject{}, err
	}
	return PolicySubject{r.Record, identity(r.Record), r.Binding.Repository.Root, remote, r.Binding.Repository.RootCommit}, nil
}
