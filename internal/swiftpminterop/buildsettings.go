package swiftpminterop

import (
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/relux-works/curator/internal/swiftpmsource"
)

// errSettingShape is the verdict for a retained build-setting payload whose
// encoding this stage does not recognize. It is not a diagnostic: the caller
// turns it into the target-scoped rejection, because a setting whose shape is
// unknown is a setting whose effect on the macro and resolution axes is
// unknown.
var errSettingShape = errors.New("build setting encoding is unrecognized")

// settingKind is one decoded SwiftPM build setting: the single kind the
// manifest declared and the operands that kind carries.
//
// `swiftpmsource.decodeBuildSetting` folds every setting to one opaque record
// whose `Value` retains the raw dump-package JSON verbatim, so the kind axis is
// recoverable here without changing a canonical capture record. The observed
// shape on the accepted profile (SwiftPM 6.3.2) is
// `{"kind":{"<name>":{"_0":…,"_1":…}},"tool":"c|cxx|swift|linker"}`, with an
// optional `"condition"` member; `_0` is a string for every kind except
// `unsafeFlags`, where it is an array, and is absent for the nullary
// `strictMemorySafety`.
type settingKind struct {
	kind   string
	values []string
}

// settingDisposition is what portable mode does with one build-setting kind.
//
// The axis is closed reject-by-default, exactly like the pragma axis: a kind is
// admitted only when it is provably BOTH macro-inert (it cannot bind a
// preprocessor macro) and resolution-inert (it cannot change where a file is
// found or read). Every other kind, and every kind this table does not name,
// rejects.
type settingDisposition int

const (
	// settingReject is a kind portable mode cannot prove inert on both axes. It
	// is the zero value, so a kind the table does not name reads as a rejection
	// even before the explicit membership check.
	settingReject settingDisposition = iota
	// settingInert is proven macro-inert and resolution-inert: the setting maps
	// to a compiler flag that carries neither a `-D` nor an include-search path.
	settingInert
	// settingDefine binds a preprocessor macro, so its name is routed into both
	// macro oracles before the setting is admitted.
	settingDefine
	// settingLink names an external library or framework and is admitted only
	// when a selected external component declares it.
	settingLink
)

// settingKindDisposition is the enumerated build-setting kind axis for the
// accepted profile, verified by running `swift package dump-package` over a
// manifest that declares every kind PackageDescription vends (SwiftPM 6.3.2,
// tools versions 5.9, 6.0, and 6.2). The emitted axis has exactly fifteen
// members: the fourteen named below with a verified encoding, plus
// `swiftLanguageVersion`, which is the deprecated PackageDescription spelling
// SwiftPM 6.3.2 already serializes as `swiftLanguageMode` and which is retained
// here only against an older serializer.
//
// The two non-inert axes have exactly one non-`unsafeFlags` member each, which
// is what makes the enumeration provable rather than enumerative:
//
//   - macro axis: `define` is the only kind that reaches the compiler as `-D`.
//     It is routed rather than rejected, because a define is ordinary and
//     legitimate; only a define that binds a gated identifier rejects.
//   - resolution axis: `headerSearchPath` is the only kind that reaches the
//     compiler as `-I`. Portable mode's include closure resolves a reference
//     against the target's own roots, so a target-declared search path changes
//     a resolution this stage cannot follow, and it rejects.
//   - `unsafeFlags` is unbounded on both axes and already rejects.
//   - `linkedLibrary`/`linkedFramework` reach the linker, not the compiler, so
//     they are inert on both axes; they are still gated on component
//     declaration because an undeclared external library is
//     `artifact_toolchain_untrusted` under the accepted SwiftPM outcome, the
//     same rule `confineLinks` applies to a module-map link edge.
//   - every remaining kind maps to a Swift or Clang flag that is a language
//     mode, a feature name, a diagnostic severity, or an isolation default:
//     none of them spells `-D` or `-I`, so both axes are closed for any operand
//     value the kind can carry, which is why an open-ended operand string such
//     as an experimental-feature name does not reopen them.
var settingKindDisposition = map[string]settingDisposition{
	"define":                    settingDefine,
	"headerSearchPath":          settingReject,
	"unsafeFlags":               settingReject,
	"linkedLibrary":             settingLink,
	"linkedFramework":           settingLink,
	"interoperabilityMode":      settingInert,
	"enableUpcomingFeature":     settingInert,
	"enableExperimentalFeature": settingInert,
	"swiftLanguageMode":         settingInert,
	"swiftLanguageVersion":      settingInert,
	"treatAllWarnings":          settingInert,
	"treatWarning":              settingInert,
	"enableWarning":             settingInert,
	"disableWarning":            settingInert,
	"strictMemorySafety":        settingInert,
	"defaultIsolation":          settingInert,
}

// disposeBuildSettings applies the enumerated kind axis to one selected
// target's build settings and returns the macro names its `define` settings
// bind.
//
// The returned set is the second input to the target's macro oracle. Both
// closing rules for the identifier positions the compiler expands —
// `atPositionIdentifiers` at a `#define`, and `rejectMacroDefinedModuleNames`
// over the scanned closure — previously read an oracle populated only by source
// `#define`s, and a SwiftPM `.define` build setting is a macro the compiler
// binds that no source line spells. Verified on the accepted Darwin profile:
// `-Dprotocol=import` with `@ protocol SecretKit;` builds SecretKit and reads
// its header, and `-DNoSuchKitXYZ=SecretKit` with `@import NoSuchKitXYZ;` does
// the same while the closure would retain `NoSuchKitXYZ` as the module read.
//
// The setting is checked on both halves of what it binds. Its NAME goes through
// the two rules above; its BODY goes through analyzeSettingDefineBody, which
// calls the same analyzer a source `#define` body calls, because a replacement
// list is itself a channel and the build-setting spelling of one reaches the
// compiler identically. Together with the source route this closes the macro
// layer across both of the pinned compiler's macro-binding inputs.
//
// Only conditions the destination selects are considered, and only for a
// selected target: a pruned or unselected setting never reaches a compiler
// invocation for this destination, so it can rebind nothing.
func (state *closeState) disposeBuildSettings(entry declaredTarget) (map[string]bool, error) {
	defines := map[string]bool{}
	for _, setting := range entry.target.Settings {
		if !conditionSelected(setting.Condition, state.markers) {
			continue
		}
		if setting.Unsafe {
			return nil, failFields(CodeUnsafeSettingForbidden, map[string]string{"target": entry.key, "setting": setting.Kind}, "selected C-family target uses unsafe flags")
		}
		decoded, err := decodeSettingKind(setting)
		if err != nil {
			return nil, failFields(CodeUnsafeSettingForbidden, map[string]string{"target": entry.key, "setting": setting.Kind}, "selected target declares a build setting whose shape portable mode cannot model, so its effect on macros and header resolution is unbounded")
		}
		disposition, known := settingKindDisposition[decoded.kind]
		if !known {
			return nil, failFields(CodeUnsafeSettingForbidden, map[string]string{"target": entry.key, "kind": decoded.kind}, "selected target declares a build setting kind portable mode cannot prove is both macro-inert and resolution-inert")
		}
		switch disposition {
		case settingDefine:
			for _, value := range decoded.values {
				name, rest := splitLeadingIdentifier(strings.TrimLeft(value, " \t"))
				if name == "" {
					return nil, failFields(CodeUnsafeSettingForbidden, map[string]string{"target": entry.key, "kind": decoded.kind, "value": value}, "selected target declares a define build setting that names no identifier this stage can bound")
				}
				if atPositionIdentifiers[name] {
					return nil, failFields(CodeHeaderInputUndeclared, map[string]string{"target": entry.key, "kind": decoded.kind, "macro": name}, "build setting binds an identifier the compiler expands in the `@`-keyword position, where a definition that reaches `import` performs an unconfined module import")
				}
				if err := analyzeSettingDefineBody(entry, decoded.kind, value, rest); err != nil {
					return nil, err
				}
				defines[name] = true
			}
		case settingLink:
			for _, value := range decoded.values {
				if !state.linkDeclared(Link{Name: value, Framework: decoded.kind == "linkedFramework"}, nil) {
					return nil, failFields(CodeToolchainUntrusted, map[string]string{"target": entry.key, "link": value, "framework": boolText(decoded.kind == "linkedFramework")}, "build setting links a library or framework that no selected external component declares")
				}
			}
		case settingInert:
		default:
			return nil, failFields(CodeUnsafeSettingForbidden, map[string]string{"target": entry.key, "kind": decoded.kind}, "selected target declares a build setting kind portable mode names as neither macro-inert nor resolution-inert")
		}
	}
	return defines, nil
}

// analyzeSettingDefineBody runs one `.define` build setting's replacement list
// through the same phase-4 analyzer a source `#define` body goes through.
//
// The name check above closes the `@`-keyword position; it says nothing about
// what the setting binds the name TO, and the body is where round 9 finding M's
// channels live. Verified on the accepted Darwin profile, a build-setting body
// really does deliver them: `-D'A=__asm__'` with `A(".incbin \"payload.bin\"");`
// produces an object byte-identical to the direct `__asm__(…)` control with the
// named file's bytes present at `0x180`, and `-D'A=_Pragma'` with
// `A("clang module import SecretKit")` builds an undeclared module and reads its
// header. The source spellings of both — `#define A __asm__` and
// `#define A _Pragma` — already reject, so this is a route gap, not a grammar
// gap, and it is closed by routing the second route into the first route's
// analyzer rather than by growing a second one.
//
// A body that resolves to a module import is rejected outright rather than
// confined: a build setting is not an admitted source file, so the reference
// belongs to no scanned unit, has no directory to resolve against, and cannot
// join the include worklist. Under the reject-by-default posture the construct
// is refused; the ordinary `-DFEATURE=1` body carries no reference at all.
func analyzeSettingDefineBody(entry declaredTarget, kind, value, rest string) error {
	parameters, body, ok := splitSettingDefine(rest)
	if !ok {
		return failFields(CodeUnsafeSettingForbidden, map[string]string{"target": entry.key, "kind": kind, "value": value}, "selected target declares a define build setting whose operand is not a macro name, optional parameter list, and replacement list this stage can separate")
	}
	references, err := analyzeMacroBody(entry.pkg, entry.name, entry.pkg, settingSourceLabel(kind, value), body, parameters)
	if errors.Is(err, errParameterPaste) {
		return failFields(CodeHeaderInputUndeclared, map[string]string{"target": entry.key, "kind": kind, "value": value}, "%s", err.Error())
	}
	if err != nil {
		return err
	}
	if len(references) > 0 {
		return failFields(CodeHeaderInputUndeclared, map[string]string{"target": entry.key, "kind": kind, "value": value, "module": references[0].Spelling}, "define build setting body performs a Clang module import that belongs to no admitted source, so this stage can neither attribute nor confine the read")
	}
	return nil
}

// splitSettingDefine separates a `.define` operand into the function-like
// parameter set and the replacement list the compiler binds.
//
// The operand SwiftPM serializes is the `-D` argument verbatim: `NAME`,
// `NAME=VALUE`, or the function-like `NAME(a,b)=VALUE`. A define with no `=`
// binds the replacement list `1`, which carries no channel, so an absent body
// is the empty body and stays admitted. Anything else between the name and the
// body is a spelling this stage cannot bound, which under the reject-by-default
// posture is a rejection rather than a guess.
func splitSettingDefine(rest string) (map[string]bool, string, bool) {
	parameters, tail, ok := readMacroParameters(rest)
	if !ok {
		return nil, "", false
	}
	switch {
	case tail == "":
		return parameters, "", true
	case strings.HasPrefix(tail, "="):
		return parameters, tail[1:], true
	}
	return nil, "", false
}

// settingSourceLabel names the build setting in a body rejection the way a file
// path names a source one, so a diagnostic raised inside the shared analyzer
// still says which input carried the channel.
func settingSourceLabel(kind, value string) string {
	return "<build setting " + kind + " " + value + ">"
}

// decodeSettingKind recovers one setting's real SwiftPM kind and operands from
// the raw dump-package JSON `swiftpmsource` retains verbatim.
//
// A `Value` that is not a JSON object is a directly constructed setting record
// rather than a decoded manifest one; its declared `Kind` and `Value` are then
// the kind and its single operand, which is the identity this function had
// before the kind axis existed.
func decodeSettingKind(setting swiftpmsource.BuildSetting) (settingKind, error) {
	trimmed := strings.TrimSpace(setting.Value)
	if !strings.HasPrefix(trimmed, "{") {
		return settingKind{kind: setting.Kind, values: []string{setting.Value}}, nil
	}
	var envelope map[string]json.RawMessage
	if err := json.Unmarshal([]byte(trimmed), &envelope); err != nil {
		return settingKind{}, errSettingShape
	}
	raw, ok := envelope["kind"]
	if !ok {
		return settingKind{}, errSettingShape
	}
	var kinds map[string]json.RawMessage
	if err := json.Unmarshal(raw, &kinds); err != nil || len(kinds) != 1 {
		return settingKind{}, errSettingShape
	}
	for name, payload := range kinds {
		values, err := decodeSettingOperands(payload)
		if err != nil {
			return settingKind{}, err
		}
		return settingKind{kind: name, values: values}, nil
	}
	return settingKind{}, errSettingShape
}

// decodeSettingOperands reads the `_0`, `_1`, … members Swift's synthesized
// enum encoding emits, in ordinal order. A member is a string or an array of
// strings; anything else is a shape this stage cannot bound.
func decodeSettingOperands(payload json.RawMessage) ([]string, error) {
	var members map[string]json.RawMessage
	if err := json.Unmarshal(payload, &members); err != nil {
		return nil, errSettingShape
	}
	keys := make([]string, 0, len(members))
	for key := range members {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	values := []string{}
	for _, key := range keys {
		var single string
		if err := json.Unmarshal(members[key], &single); err == nil {
			values = append(values, single)
			continue
		}
		var list []string
		if err := json.Unmarshal(members[key], &list); err != nil {
			return nil, errSettingShape
		}
		values = append(values, list...)
	}
	return values, nil
}
