package gitcred

import (
	"context"
	"strings"
	"testing"
)

func TestValidProviderAdmitsOpaqueIdentifiersOnly(t *testing.T) {
	for _, id := range []string{
		"team-ssh",
		"team-https",
		"anonymous",
		"a",
		"ANON-1.x_y",
		strings.Repeat("a", 128),
	} {
		if !ValidProvider(id) {
			t.Errorf("ValidProvider(%q) = false", id)
		}
	}
	for _, id := range []string{
		"",
		"/bin/sh",
		"./helper",
		"../escape",
		`C:\Windows\helper.exe`,
		"C:/Windows/helper.exe",
		"team ssh",
		"team/ssh",
		"team:ssh",
		"team@ssh",
		"team;rm",
		"team|pipe",
		"team&bg",
		"$(helper)",
		"`helper`",
		"ssh://example.org/kit",
		"trailing.",
		"con",
		"COM1",
		strings.Repeat("a", 129),
		"ünicode",
	} {
		if ValidProvider(id) {
			t.Errorf("ValidProvider(%q) = true", id)
		}
	}
}

func TestProviderUsernameNamespacesOpaqueIdentifiersOnly(t *testing.T) {
	if got, want := ProviderUsername("team-https"), "curator-provider-https:team-https"; got != want {
		t.Fatalf("ProviderUsername = %q, want %q", got, want)
	}
	for _, id := range []string{"", "../escape", "/bin/sh", "team ssh", "team/ssh", "ssh://example.org/kit"} {
		if got := ProviderUsername(id); got != "" {
			t.Fatalf("ProviderUsername(%q) = %q, want empty", id, got)
		}
	}
	if ProviderNamespacePrefix == NamespacePrefix {
		t.Fatal("provider entries must live under their own namespace, disjoint from scope entries")
	}
}

func TestReadProviderGoesThroughGitCredentialFill(t *testing.T) {
	access, dir := fakeAccess(t, "")
	seedStore(t, dir, credEntry{Host: "git.example.org", Username: ProviderUsername("team-https"), Password: "s3cret"})

	credential, ok := access.ReadProvider(context.Background(), "team-https", "git.example.org")
	if !ok {
		t.Fatal("ReadProvider found nothing")
	}
	if credential.Username != ProviderUsername("team-https") || credential.Secret != "s3cret" {
		t.Fatalf("ReadProvider = %+v", credential)
	}
	call := singleCall(t, dir)
	if want := "protocol=https\nhost=git.example.org\nusername=" + ProviderUsername("team-https") + "\n\n"; call.Stdin != want {
		t.Fatalf("request payload = %q, want %q", call.Stdin, want)
	}
}

func TestReadProviderRejectsNearMissAnswers(t *testing.T) {
	access, dir := fakeAccess(t, modeIgnoreUsername)
	seedStore(t, dir, credEntry{Host: "git.example.org", Username: "oauth2", Password: "s3cret"})

	if _, ok := access.ReadProvider(context.Background(), "team-https", "git.example.org"); ok {
		t.Fatal("ReadProvider accepted a credential the helper answered for another username")
	}
}

func TestReadProviderRefusesInvalidReferencesBeforeAnyCall(t *testing.T) {
	access, dir := fakeAccess(t, "")
	for _, refs := range [][2]string{
		{"../escape", "git.example.org"},
		{"team ssh", "git.example.org"},
		{"", "git.example.org"},
		{"team-https", ""},
		{"team-https", "bad\nhost"},
	} {
		if _, ok := access.ReadProvider(context.Background(), refs[0], refs[1]); ok {
			t.Fatalf("ReadProvider(%q, %q) accepted an invalid reference", refs[0], refs[1])
		}
	}
	if calls := recordedCalls(t, dir); len(calls) != 0 {
		t.Fatalf("invalid references reached the credential machinery %d times", len(calls))
	}
}

func TestReadProviderFindsNothingWhenAbsent(t *testing.T) {
	access, dir := fakeAccess(t, "")
	seedStore(t, dir, credEntry{Host: "git.example.org", Username: "oauth2", Password: "s3cret"})
	if _, ok := access.ReadProvider(context.Background(), "team-https", "git.example.org"); ok {
		t.Fatal("ReadProvider reported material for an unconfigured provider")
	}
}

func TestReadHostExcludesProviderEntries(t *testing.T) {
	access, dir := fakeAccess(t, "")
	seedStore(t, dir, credEntry{Host: "git.example.org", Username: ProviderUsername("team-https"), Password: "s3cret"})

	if _, ok := access.ReadHost(context.Background(), "git.example.org"); ok {
		t.Fatal("ReadHost reported a provider-held entry as the operator's own credential")
	}
}
