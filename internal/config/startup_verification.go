package config

import (
	"bytes"
	"encoding/json"
	"net"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/relux-works/curator/internal/startupverification"
)

// validateStartupJSON preserves the numeric spelling before legacy config
// projections turn JSON numbers into float64. Unrelated fields keep their
// established representation.
func validateStartupJSON(payload []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(payload, &fields); err != nil {
		return err
	}
	if raw, present := fields["startup_verification"]; present {
		var value any
		d := json.NewDecoder(bytes.NewReader(raw))
		d.UseNumber()
		if err := d.Decode(&value); err != nil {
			return startupverification.Invalid("malformed level object")
		}
		_, err := ParseStartupVerification(value)
		return err
	}
	return nil
}

// StartupVerificationSchemaVersion is the unsigned startup-policy carrier.
const StartupVerificationSchemaVersion = 11

type StartupVerification struct {
	Level int `json:"level"`
}
type GrantServiceReference struct {
	Endpoint string `json:"endpoint"`
}
type StartupVerificationProjection struct {
	Schema       string                 `json:"schema"`
	Level        int                    `json:"level"`
	Origin       string                 `json:"origin"`
	GrantService *GrantServiceReference `json:"grant_service,omitempty"`
}

// ParseStartupVerification accepts the closed object without numeric coercion.
// JSON numbers must spell an integer; 3.0 and 3e0 are not admitted aliases.
func ParseStartupVerification(raw any) (StartupVerification, error) {
	level, err := startupverification.ParseLevelObject(raw)
	return StartupVerification{Level: level}, err
}

func (c *Config) StartupProjection() StartupVerificationProjection {
	p := StartupVerificationProjection{Schema: "curator.startup-verification/1", Level: c.StartupVerification.Level, Origin: c.StartupVerificationOrigin}
	if c.Extensions != nil && c.Extensions.GrantService != nil {
		v := *c.Extensions.GrantService
		p.GrantService = &v
	}
	return p
}

func parseGrantService(raw any) (*GrantServiceReference, error) {
	object, ok := launcherMembers(raw, "endpoint")
	if !ok {
		return nil, startupverification.Invalid("grant_service requires only endpoint")
	}
	rawEndpoint, ok := object["endpoint"].(string)
	if !ok || !validGrantEndpoint(rawEndpoint) {
		return nil, startupverification.Invalid("grant_service endpoint must be an absolute HTTPS locator without credentials, query or fragment")
	}
	return &GrantServiceReference{Endpoint: rawEndpoint}, nil
}

func validGrantEndpoint(raw string) bool {
	if raw == "" || len(raw) > 4096 || !utf8.ValidString(raw) || strings.ContainsFunc(raw, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) || strings.ContainsAny(raw, "\\#") {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Opaque != "" || u.User != nil || u.Hostname() == "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return false
	}
	host := u.Hostname()
	if strings.Contains(host, ":") && !strings.HasPrefix(u.Host, "[") {
		return false
	}
	if net.ParseIP(host) == nil && (len(host) > 253 || !registryHostRE.MatchString(host)) {
		return false
	}
	if strings.HasSuffix(u.Host, ":") {
		return false
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		if err != nil || n < 1 || n > 65535 {
			return false
		}
	}
	if strings.ContainsFunc(u.Path, unicode.IsControl) || strings.ContainsRune(u.Path, '\\') {
		return false
	}
	for _, part := range strings.Split(u.Path, "/") {
		if part == "." || part == ".." {
			return false
		}
	}
	return true
}
