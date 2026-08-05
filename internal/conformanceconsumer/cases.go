package conformanceconsumer

import (
	"encoding/json"
	"fmt"
)

// Case is one implementation-neutral external-repository observation from the
// shared rc.5 case manifest. Expected remains raw JSON so the protocol corpus,
// rather than this consumer, owns the outcome vocabulary.
type Case struct {
	ID       string          `json:"id"`
	Category string          `json:"category"`
	Source   string          `json:"source"`
	Expected json.RawMessage `json:"expected"`
}

type caseManifest struct {
	Cases []Case `json:"cases"`
}

// Cases authenticates and decodes the shared case manifest.
func (c *Corpus) Cases() ([]Case, error) {
	payload, _, err := c.Read("case-manifest.json")
	if err != nil {
		return nil, err
	}
	var document caseManifest
	if err := json.Unmarshal(payload, &document); err != nil {
		return nil, fmt.Errorf("decode case manifest: %w", err)
	}
	seen := make(map[string]struct{}, len(document.Cases))
	for _, item := range document.Cases {
		if item.ID == "" || item.Category == "" || item.Source == "" || len(item.Expected) == 0 {
			return nil, fmt.Errorf("case manifest contains an incomplete case")
		}
		if _, ok := seen[item.ID]; ok {
			return nil, fmt.Errorf("case manifest contains duplicate case %q", item.ID)
		}
		seen[item.ID] = struct{}{}
	}
	return append([]Case(nil), document.Cases...), nil
}
