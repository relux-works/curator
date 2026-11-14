package contextlock

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/relux-works/curator/internal/skillintegrity"
	"github.com/relux-works/curator/internal/startupverification"
)

// SkillIntegritySchemaVersion extends lock4 without adopting lock5 skill plans.
const SkillIntegritySchemaVersion = 6

func hasSettingsCarrier(version int) bool {
	return version == SchemaVersion4 || version == SchemaVersion5 || version == SkillIntegritySchemaVersion
}

// SkillIdentity projects an already validated lock member's evidence identity.
func SkillIdentity(member Member) skillintegrity.Identity {
	return skillintegrity.Identity{Name: member.Name, Source: member.Source, Commit: member.Commit, StateSHA256: member.StateHash, Directory: member.Directory}
}

func (lock *Lock) validateIntegrity() error {
	if lock.SchemaVersion != SkillIntegritySchemaVersion {
		if lock.StartupVerificationLevel != nil || lock.SkillIntegrity != nil {
			return fmt.Errorf("skill_integrity_manifest_invalid: integrity fields on legacy lock")
		}
		return nil
	}
	if lock.StartupVerificationLevel != nil {
		if _, err := startupverification.EffectiveLevel(*lock.StartupVerificationLevel); err != nil {
			return err
		}
	}
	if lock.SkillIntegrity == nil {
		return fmt.Errorf("skill_integrity_evidence_missing: lock6 requires skill_integrity")
	}
	index := 0
	var files int
	var total int64
	for _, member := range lock.Members {
		if member.Kind != KindSkill {
			continue
		}
		if index >= len(lock.SkillIntegrity) {
			return fmt.Errorf("skill_integrity_evidence_missing: skill evidence is incomplete")
		}
		entry := lock.SkillIntegrity[index]
		index++
		if err := entry.Validate(SkillIdentity(member)); err != nil {
			return err
		}
		for _, file := range entry.Manifest.Files {
			files++
			total += file.Size
			if files > skillintegrity.MaxFiles || total > skillintegrity.MaxBytes {
				return fmt.Errorf("skill_integrity_manifest_invalid: closure evidence budget exceeded")
			}
		}
	}
	if index != len(lock.SkillIntegrity) {
		return fmt.Errorf("skill_integrity_manifest_invalid: extra or unordered skill evidence")
	}
	return nil
}

func (lock *Lock) parseIntegrity(rawFloor, rawEntries json.RawMessage, present map[string]json.RawMessage) error {
	if lock.SchemaVersion != SkillIntegritySchemaVersion {
		if _, ok := present["startup_verification_level"]; ok {
			return fmt.Errorf("startup_verification_schema_unsupported: profile floor requires lock6")
		}
		if _, ok := present["skill_integrity"]; ok {
			return fmt.Errorf("skill_integrity_manifest_invalid: integrity fields require lock6")
		}
		return nil
	}
	if _, ok := present["startup_verification_level"]; ok {
		level, err := startupverification.ParseLevelObject(map[string]any{"level": json.Number(string(rawFloor))})
		if err != nil {
			return err
		}
		lock.StartupVerificationLevel = &level
	}
	decoder := json.NewDecoder(bytes.NewReader(rawEntries))
	if token, err := decoder.Token(); err != nil || token != json.Delim('[') {
		return fmt.Errorf("skill_integrity_evidence_missing: lock6 requires an integrity array")
	}
	lock.SkillIntegrity = []skillintegrity.Entry{}
	var skills []Member
	for _, member := range lock.Members {
		if member.Kind == KindSkill {
			skills = append(skills, member)
		}
	}
	for decoder.More() {
		if len(lock.SkillIntegrity) >= len(skills) {
			return fmt.Errorf("skill_integrity_manifest_invalid: extra skill evidence")
		}
		var raw json.RawMessage
		if decoder.Decode(&raw) != nil {
			return fmt.Errorf("skill_integrity_manifest_invalid: invalid skill evidence")
		}
		entry, err := skillintegrity.DecodeEntry(raw, SkillIdentity(skills[len(lock.SkillIntegrity)]))
		if err != nil {
			return err
		}
		lock.SkillIntegrity = append(lock.SkillIntegrity, entry)
	}
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("skill_integrity_manifest_invalid: invalid integrity array")
	}
	return nil
}

func integrityInputBudget(payload []byte) error {
	// Project only the schema token before allocating members/evidence. Older
	// native lock carriers retain their existing input-size behavior.
	var head struct {
		Schema json.RawMessage `json:"schema_version"`
	}
	if json.Unmarshal(payload, &head) == nil && string(head.Schema) == fmt.Sprint(SkillIntegritySchemaVersion) && int64(len(payload)) > skillintegrity.MaxBytes {
		return fmt.Errorf("skill_integrity_manifest_invalid: lock6 input exceeds evidence budget")
	}
	return nil
}
