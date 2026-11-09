package profilegraph

import (
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"

	"github.com/relux-works/curator/internal/config"
	"github.com/relux-works/curator/internal/contextmaterialize"
	"github.com/relux-works/curator/internal/envprofile"
	"github.com/relux-works/curator/internal/protocoljson"
)

// ReadProjection consumes one bounded machine file but decodes only selected
// environment knobs. The input ledger contains its canonical restricted
// projection, never the raw configuration or unrelated values.
func ReadProjection(r *Reader, configPath, profile string) (Projection, error) {
	return ReadProjectionWithSystem(r, configPath, "", profile, false)
}

// ReadProjectionWithSystem applies the real system lock merge to selected knobs.
func ReadProjectionWithSystem(r *Reader, configPath, systemPath, profile string, optional bool) (Projection, error) {
	user, _, err := readSelection(r, configPath, "machine-selection")
	if err != nil && (!optional || !isMissing(err)) {
		return Projection{}, err
	}
	var cfg config.Environments
	if systemPath != "" {
		system, locked, err := readSelection(r, systemPath, "system-selection")
		if err != nil && !isMissing(err) {
			return Projection{}, err
		}
		cfg, err = config.MergeGraphEnvironments(user, system, locked)
		if err != nil {
			return Projection{}, refusal("graph_invalid_input")
		}
	} else {
		cfg, err = config.ParseGraphEnvironments(user)
		if err != nil {
			return Projection{}, refusal("graph_invalid_input")
		}
	}
	out := Projection{Precedence: contextmaterialize.Precedence{Winner: cfg.Precedence.Winner, Placement: cfg.Precedence.Placement}, Forms: cfg.Forms, Admission: contextmaterialize.Admission{Transitive: cfg.TransitiveSystemModules, Waivers: map[string]bool{}}, SystemPromptFiles: cfg.SystemPromptFiles[profile].Pi}
	for _, waiver := range cfg.SystemModuleWaivers {
		out.Admission.Waivers[waiver.Package] = true
	}
	for _, decl := range cfg.EffectiveOverlays(profile) {
		weight := int64(cfg.OverlayDefaultWeight)
		if decl.Weight != nil {
			weight = int64(*decl.Weight)
		}
		out.Overlays = append(out.Overlays, envprofile.OverlaySpec{Source: decl.Source, Range: decl.Range, Tag: decl.Tag, Revision: decl.Revision, Directory: decl.Directory, Weight: &weight})
	}
	return out, nil
}

func decodeRawObject(raw []byte) (map[string]json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, refusal("graph_invalid_input")
	}
	out := map[string]json.RawMessage{}
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			return nil, refusal("graph_invalid_input")
		}
		name, ok := key.(string)
		if !ok {
			return nil, refusal("graph_invalid_input")
		}
		if _, exists := out[name]; exists {
			return nil, refusal("graph_invalid_input")
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return nil, refusal("graph_invalid_input")
		}
		out[name] = value
	}
	if _, err := decoder.Token(); err != nil {
		return nil, refusal("graph_invalid_input")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, refusal("graph_invalid_input")
	}
	return out, nil
}
func readSelection(r *Reader, configPath, label string) (map[string]any, map[string]bool, error) {
	root, err := filepath.Abs(filepath.Dir(configPath))
	if err != nil {
		return nil, nil, refusal("graph_invalid_input")
	}
	rel := filepath.Base(configPath)
	raw, err := r.Read(root, label, rel)
	if err != nil {
		return nil, nil, err
	}
	delete(r.inputs, label+"\x00"+rel)
	delete(r.cache, label+"\x00"+rel)
	top, err := decodeRawObject(raw)
	if err != nil {
		return nil, nil, err
	}
	env := map[string]json.RawMessage{}
	if value, present := top["environments"]; present {
		env, err = decodeRawObject(value)
		if err != nil {
			return nil, nil, err
		}
	}
	selected := map[string]any{}
	for _, key := range []string{"overlays", "overlay_default_weight", "overlays_allowed", "precedence", "forms", "system_prompt_files", "transitive_system_modules", "system_module_waivers"} {
		if value, present := env[key]; present {
			if err := protocoljson.Validate(value); err != nil {
				return nil, nil, refusal("graph_invalid_input")
			}
			decoder := json.NewDecoder(bytes.NewReader(value))
			decoder.UseNumber()
			var decoded any
			if decoder.Decode(&decoded) != nil {
				return nil, nil, refusal("graph_invalid_input")
			}
			selected[key] = decoded
		}
	}
	locked := map[string]bool{}
	if value, ok := top["locked"]; ok {
		var keys []string
		if json.Unmarshal(value, &keys) != nil {
			return nil, nil, refusal("graph_invalid_input")
		}
		for _, key := range keys {
			locked[key] = true
		}
	}
	return selected, locked, nil
}
