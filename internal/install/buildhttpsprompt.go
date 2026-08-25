package install

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/gitcred"
)

// BuildHTTPSTokenReader reads a token without putting it in argv or config.
// Production supplies a terminal-hidden reader; tests inject a deterministic
// one without weakening the production path.
type BuildHTTPSTokenReader func() (string, error)

// PersistBuildHTTPS records an explicitly persistent source selection. For a
// keyring selection secret is the token to store in the operator helper; for
// an existing host credential it is empty because that material is already
// operator-owned and must not be copied.
type PersistBuildHTTPS func(context.Context, config.BuildHTTPSCredential, string) error

// InteractiveBuildHTTPSResolver mirrors the SSH candidate prompt. Discovery
// only describes existing material. No source is read until selected, and no
// config or keyring write occurs when the scope answer is this-run-only.
func InteractiveBuildHTTPSResolver(in io.Reader, out io.Writer, readToken BuildHTTPSTokenReader, persist PersistBuildHTTPS) BuildHTTPSResolver {
	return func(ctx context.Context, requests []BuildHTTPSRequest, candidates map[string]gitcred.HostMaterial, reader BuildHTTPSCredentialReader) (map[string]BuildHTTPSCredentials, error) {
		input := bufio.NewReader(in)
		added := map[string]BuildHTTPSCredentials{}
		for _, request := range requests {
			if _, _, covered := matchPromptedBuildHTTPS(added, request.Identity); covered {
				continue
			}
			say(out, "\nbuild_repository_https_credential_missing: %s can use HTTPS credentials (command %q of skill %q)\n",
				request.Identity, request.Command, request.Skill)
			material := candidates[request.Host]
			writeBuildHTTPSMenu(out, material)
			source, err := readBuildHTTPSChoice(input, out, material)
			if err != nil {
				return nil, err
			}

			var selected BuildHTTPSCredentials
			var secret string
			credential := config.BuildHTTPSCredential{}
			switch source {
			case config.TokenSourceGitCredentials:
				hostCredential, ok := reader.ReadHost(ctx, request.Host)
				if !ok {
					return nil, fmt.Errorf("the selected Git HTTPS credential for host %q is no longer available", request.Host)
				}
				selected = NewBuildHTTPSCredentials(hostCredential.Username, hostCredential.Secret)
				credential.Token = config.TokenSourceGitCredentials
			case config.TokenSourceKeyring:
				if readToken == nil {
					return nil, errorsForMissingTokenReader()
				}
				secret, err = readToken()
				if err != nil {
					return nil, err
				}
				if strings.TrimSpace(secret) == "" || strings.ContainsAny(secret, "\r\n") {
					return nil, fmt.Errorf("HTTPS token must be a non-empty single line")
				}
				selected = NewBuildHTTPSCredentials(config.BuildHTTPSDefaultUsername, secret)
				credential.Token = config.TokenSourceKeyring
			}

			scope, save, err := readCredentialScope(input, out, request.DefaultScope, request.Identity)
			if err != nil {
				if errors.Is(err, ErrBuildSSHAborted) {
					return nil, ErrBuildHTTPSAborted
				}
				return nil, err
			}
			credential.Scope = scope
			if save {
				if persist == nil {
					return nil, fmt.Errorf("HTTPS credential persistence is unavailable")
				}
				if err := persist(ctx, credential, secret); err != nil {
					return nil, err
				}
				say(out, "recorded build_https scope %s\n", scope)
			} else {
				say(out, "using HTTPS credential for this run only\n")
			}
			added[scope] = selected
		}
		return added, nil
	}
}

func errorsForMissingTokenReader() error {
	return fmt.Errorf("HTTPS token input is unavailable")
}

func writeBuildHTTPSMenu(out io.Writer, material gitcred.HostMaterial) {
	if material.HostCredential {
		say(out, "  1) existing Git HTTPS credential for this host (username %s)  [default]\n", material.HostUsername)
	} else {
		say(out, "no existing Git HTTPS credential was detected for this host\n")
	}
	say(out, "  t) enter a token now\n")
	say(out, "  %s) abort\n", abortToken)
}

func readBuildHTTPSChoice(reader *bufio.Reader, out io.Writer, material gitcred.HostMaterial) (string, error) {
	for {
		prompt := fmt.Sprintf("credential [t, %s]", abortToken)
		if material.HostCredential {
			prompt = fmt.Sprintf("credential [1, t, %s] (default 1)", abortToken)
		}
		answer, err := ask(reader, out, prompt)
		if err != nil {
			if errors.Is(err, ErrBuildSSHAborted) {
				return "", ErrBuildHTTPSAborted
			}
			return "", err
		}
		switch {
		case answer == abortToken:
			return "", ErrBuildHTTPSAborted
		case material.HostCredential && (answer == "" || answer == "1"):
			return config.TokenSourceGitCredentials, nil
		case answer == "t":
			return config.TokenSourceKeyring, nil
		default:
			say(out, "curator: %q is not one of the offered choices\n", answer)
		}
	}
}
