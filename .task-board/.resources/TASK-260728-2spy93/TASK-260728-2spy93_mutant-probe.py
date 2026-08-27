"""Prove the cycle-3, cycle-4 and cycle-5 mutants were accepted by the pre-fix gate.

The probe imports the shipped validator, restores each affected checker to its
pre-fix body, and runs the exact mutants the reviewers demonstrated. It writes
no project file. Exit 0 means every mutant was accepted before its fix and is
rejected after it.
"""

from __future__ import annotations

import copy
import importlib.util
from pathlib import Path

WORKTREE = Path(__file__).resolve().parent / "curator-spec-worktree"
MODULE_PATH = WORKTREE / "tools" / "validate.py"
SPEC = importlib.util.spec_from_file_location("probe_validate", MODULE_PATH)
assert SPEC is not None and SPEC.loader is not None
validate = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(validate)


class Patch:
    def __init__(self, owner, name, replacement):
        self.owner, self.name, self.replacement = owner, name, replacement
        self.original = getattr(owner, name)

    def __enter__(self):
        setattr(self.owner, self.name, self.replacement)

    def __exit__(self, *_):
        setattr(self.owner, self.name, self.original)


def patched_json(documents):
    original = validate.load_json
    resolved = {Path(path): document for path, document in documents.items()}

    def patched(path):
        document = resolved.get(Path(path))
        if document is not None:
            return copy.deepcopy(document)
        return original(path)

    return Patch(validate, "load_json", patched)


# --- pre-fix bodies, transcribed from the cycle-2 validator -----------------


def pre_fix_check_closed_member_set(name, definition, required, optional):
    if definition.get("additionalProperties") is not False:
        raise validate.ValidationFailure(
            f"common.schema.json $defs.{name} does not close additionalProperties"
        )
    expected_members = set(required) | set(optional)
    members = set(definition.get("properties", {}))
    if members != expected_members:
        raise validate.ValidationFailure(
            f"common.schema.json $defs.{name} is not its closed member set"
        )
    if set(definition.get("required", [])) != set(required):
        raise validate.ValidationFailure(
            f"common.schema.json $defs.{name} required set is not closed"
        )


def pre_fix_check_claim_driver_admission(common):
    asserting = []
    for version, path in validate.conformance_claim_schemas():
        claim = validate.load_json(path)
        member = claim.get("properties", {}).get(validate.CLAIM_DRIVER_MEMBER)
        if member is None:
            continue
        assertions = member.get("items", {}).get("oneOf")
        if not isinstance(assertions, list) or not assertions:
            raise validate.ValidationFailure(f"{path}: not a closed oneOf")
        claimed = set()
        for assertion in assertions:
            label = f"{path}: assertion"
            properties = assertion.get("properties") if isinstance(assertion, dict) else None
            if not isinstance(properties, dict):
                raise validate.ValidationFailure(f"{label} is not an object schema")
            if assertion.get("additionalProperties") is not False:
                raise validate.ValidationFailure(f"{label} does not close additionalProperties")
            if tuple(sorted(properties)) != validate.CLAIM_ASSERTION_MEMBERS or set(
                assertion.get("required", [])
            ) != set(validate.CLAIM_ASSERTION_MEMBERS):
                raise validate.ValidationFailure(f"{label} is not the closed assertion member set")
            driver = properties["driver"]
            if not isinstance(driver, dict) or set(driver) != {"const"}:
                raise validate.ValidationFailure(f"{label} does not close driver with a const")
            name = driver["const"]
            if name not in validate.ADMITTED_BUILD_DRIVERS:
                raise validate.ValidationFailure(
                    f"{label} asserts {name!r}, which is not in the admitted wire driver set"
                )
            if name in claimed:
                raise validate.ValidationFailure(f"{label} asserts {name!r} more than once")
            claimed.add(name)
            policy = validate.closed_const(
                common, properties["execution_policy"], f"{label} execution_policy"
            )
            if policy != validate.DRIVER_EXECUTION_POLICIES[name]:
                raise validate.ValidationFailure(f"{label} pairs {name!r} with {policy!r}")
        asserting.append((version, path, claimed))
    if not asserting:
        raise validate.ValidationFailure("no conformance claim schema asserts a build driver")
    _, current, claimed = asserting[-1]
    missing = sorted(set(validate.ADMITTED_BUILD_DRIVERS) - claimed)
    if missing:
        raise validate.ValidationFailure(
            f"{current}: the current claim schema does not assert every admitted wire driver"
        )


def pre_fix_check_build_artifact_closure(common):
    """The cycle-3 artifact check, transcribed from its inline body."""
    artifact = common["$defs"]["buildArtifactV1"]
    if (
        set(artifact["properties"]) != {"path", "sha256", "size"}
        or artifact.get("additionalProperties") is not False
    ):
        raise validate.ValidationFailure(
            "buildArtifactV1 is not the closed single-file artifact: "
            f"{sorted(artifact['properties'])}"
        )


def pre_fix_check_build_artifact_rejections(registry, paths):
    """The cycle-3 gate proved nothing about the artifact behaviourally."""
    return None


def pre_fix_check_artifact_reference_targets(common):
    """The cycle-4 gate pinned the three ``$ref`` values and nothing beyond."""
    return None


# --- mutants ---------------------------------------------------------------


def requirement():
    return {
        "type": "object",
        "required": ["id", "version"],
        "properties": {"id": {"type": "string"}, "version": {"type": "object"}},
        "additionalProperties": False,
    }


def minted_with_string_requirement():
    common = validate.load_json(validate.SCHEMAS / "common.schema.json")
    name = "buildCommandV8"
    required, _ = validate.RESERVED_WIRE_SHAPES[name]
    members = {
        member: {"const": "build"} if member == "type" else {"type": "string"}
        for member in sorted(required)
    }
    members["driver"] = {"const": validate.ADMITTED_BUILD_DRIVERS[0]}
    members["toolchain"] = copy.deepcopy(validate.TOOLCHAIN_REQUIREMENT_REF)
    text = requirement()
    text["type"] = "string"
    common["$defs"][validate.TOOLCHAIN_REQUIREMENT_DEFINITION] = text
    common["$defs"][name] = {
        "type": "object",
        "required": sorted(required),
        "properties": members,
        "additionalProperties": False,
    }
    return common


def current_claim():
    _, path = validate.conformance_claim_schemas()[-1]
    return path, validate.load_json(path)


def run(label, thunk, *, expect_reject):
    try:
        thunk()
    except validate.ValidationFailure as exc:
        outcome = "REJECTED"
        detail = str(exc).splitlines()[0][:140]
    else:
        outcome = "ACCEPTED"
        detail = ""
    ok = outcome == ("REJECTED" if expect_reject else "ACCEPTED")
    print(f"[{'ok' if ok else 'FAIL'}] {label}: {outcome} {detail}")
    return ok


def main() -> int:
    results = []
    path, claim = current_claim()
    newer = path.with_name("conformance-claim-v99.schema.json")

    # Mutant 1: non-object toolchainRequirementV1.
    common = minted_with_string_requirement()

    def mutant_one():
        with patched_json({validate.SCHEMAS / "common.schema.json": common}):
            validate.validate_additional_driver_boundary()

    with Patch(validate, "check_closed_member_set", pre_fix_check_closed_member_set):
        results.append(run("pre-fix  | string toolchainRequirementV1", mutant_one, expect_reject=False))
    results.append(run("post-fix | string toolchainRequirementV1", mutant_one, expect_reject=True))

    # Mutant 2: newest claim schema without build_drivers.
    stripped = copy.deepcopy(claim)
    del stripped["properties"][validate.CLAIM_DRIVER_MEMBER]

    def mutant_two():
        with Patch(validate, "conformance_claim_schemas", lambda: [(3, path), (99, newer)]):
            with patched_json({newer: stripped}):
                validate.validate_additional_driver_boundary()

    with Patch(validate, "check_claim_driver_admission", pre_fix_check_claim_driver_admission):
        results.append(run("pre-fix  | newest claim drops build_drivers", mutant_two, expect_reject=False))
    results.append(run("post-fix | newest claim drops build_drivers", mutant_two, expect_reject=True))

    # Mutant 3: reserved driver reached through prefixItems.
    driver = validate.RESERVED_BUILD_DRIVERS[0]
    smuggled = copy.deepcopy(claim)
    member = smuggled["properties"][validate.CLAIM_DRIVER_MEMBER]
    reserved = copy.deepcopy(member["items"]["oneOf"][0])
    reserved["properties"]["driver"] = {"const": driver}
    member["prefixItems"] = [reserved]

    def mutant_three():
        with Patch(validate, "conformance_claim_schemas", lambda: [(3, path), (99, newer)]):
            with patched_json({newer: smuggled}):
                validate.validate_additional_driver_boundary()

    with Patch(validate, "check_claim_driver_admission", pre_fix_check_claim_driver_admission):
        results.append(run(f"pre-fix  | {driver} via prefixItems", mutant_three, expect_reject=False))
    results.append(run(f"post-fix | {driver} via prefixItems", mutant_three, expect_reject=True))

    # The prefixItems escape is real under the compiled Draft 2020-12 validator.
    registry, _ = validate.schema_registry()
    instance = validate.load_json(
        validate.SUITE / "schema-cases" / "conformance-claim-v3" / "valid.json"
    )
    validate.set_at(instance, ("build_drivers", 0), driver)
    frozen_errors = list(validate.Draft202012Validator(claim, registry=registry).iter_errors(instance))
    escaped_errors = list(
        validate.Draft202012Validator(smuggled, registry=registry).iter_errors(instance)
    )
    real = bool(frozen_errors) and not escaped_errors
    print(
        f"[{'ok' if real else 'FAIL'}] draft2020-12 | shipped claim rejects {driver}"
        f" ({len(frozen_errors)} errors), prefixItems variant accepts it"
        f" ({len(escaped_errors)} errors)"
    )
    results.append(real)

    # Mutants 4 and 5: the cycle-4 artifact escapes. Both keep the three
    # property names and additionalProperties: false, which is all the pre-fix
    # check compared.
    artifact_common = validate.load_json(validate.SCHEMAS / "common.schema.json")
    union = copy.deepcopy(artifact_common)
    union["$defs"]["buildArtifactV1"]["type"] = ["object", "string"]
    empty_required = copy.deepcopy(artifact_common)
    empty_required["$defs"]["buildArtifactV1"]["required"] = []

    artifact_patches = (
        Patch(validate, "check_build_artifact_closure", pre_fix_check_build_artifact_closure),
        Patch(validate, "check_build_artifact_rejections", pre_fix_check_build_artifact_rejections),
    )

    def artifact_mutant(mutant):
        def thunk():
            with patched_json({validate.SCHEMAS / "common.schema.json": mutant}):
                validate.validate_additional_driver_boundary()

        return thunk

    for label, mutant in (
        ("buildArtifactV1 typed object-or-string", union),
        ("buildArtifactV1 requires nothing", empty_required),
    ):
        thunk = artifact_mutant(mutant)
        with artifact_patches[0], artifact_patches[1]:
            results.append(run(f"pre-fix  | {label}", thunk, expect_reject=False))
        results.append(run(f"post-fix | {label}", thunk, expect_reject=True))

    # Both escapes are real under the compiled Draft 2020-12 validator: the
    # shipped receipt schema rejects them, the mutated one accepts, and every
    # generated positive case keeps validating either way.
    receipt = "build-receipt-v2"
    positive = validate.load_json(validate.SUITE / "schema-cases" / receipt / "valid.json")
    launcher = copy.deepcopy(positive)
    launcher["artifact"] = "bin/golden-tool-launcher"
    nothing = copy.deepcopy(positive)
    nothing["artifact"] = {}

    def receipt_errors(common_document, instance):
        with patched_json({validate.SCHEMAS / "common.schema.json": common_document}):
            registry, paths = validate.schema_registry()
            schema = validate.load_json(paths[f"{receipt}.schema.json"])
            return list(
                validate.Draft202012Validator(schema, registry=registry).iter_errors(instance)
            )

    for label, mutant_document, escape in (
        ("a launcher string", union, launcher),
        ("an empty artifact object", empty_required, nothing),
    ):
        shipped = receipt_errors(artifact_common, escape)
        widened = receipt_errors(mutant_document, escape)
        positives_hold = not receipt_errors(mutant_document, positive)
        real = bool(shipped) and not widened and positives_hold
        print(
            f"[{'ok' if real else 'FAIL'}] draft2020-12 | shipped {receipt} rejects {label}"
            f" ({len(shipped)} errors), mutated variant accepts it ({len(widened)} errors),"
            f" positive case still validates: {positives_hold}"
        )
        results.append(real)

    # Mutants 6, 7 and 8: the cycle-5 identity-target escapes. Each leaves the
    # three artifact $ref values untouched and every sampled negative still
    # rejected, so the cycle-4 gate accepted all three.
    long_path = copy.deepcopy(artifact_common)
    long_path["$defs"]["portablePath"]["maxLength"] = 4097
    upper_digest = copy.deepcopy(artifact_common)
    upper_digest["$defs"]["sha256"]["pattern"] = "^sha256:[0-9a-fA-F]{64}$"
    big_size = copy.deepcopy(artifact_common)
    big_size["$defs"]["nonNegativeSafeInteger"]["maximum"] = validate.SAFE_INTEGER + 1

    published = copy.deepcopy(positive["artifact"])
    target_mutants = (
        (
            "portablePath.maxLength 4096 -> 4097",
            long_path,
            {**published, "path": "a" * 4097},
        ),
        (
            "sha256.pattern admits uppercase hex",
            upper_digest,
            {**published, "sha256": "sha256:" + "A" * 64},
        ),
        (
            "nonNegativeSafeInteger.maximum + 1",
            big_size,
            {**published, "size": validate.SAFE_INTEGER + 1},
        ),
    )

    for label, mutant, admitted in target_mutants:
        thunk = artifact_mutant(mutant)
        with Patch(
            validate,
            "check_artifact_reference_targets",
            pre_fix_check_artifact_reference_targets,
        ):
            results.append(run(f"pre-fix  | {label}", thunk, expect_reject=False))
        results.append(run(f"post-fix | {label}", thunk, expect_reject=True))

        escape = copy.deepcopy(positive)
        escape["artifact"] = admitted
        shipped = receipt_errors(artifact_common, escape)
        widened = receipt_errors(mutant, escape)
        positives_hold = not receipt_errors(mutant, positive)
        real = bool(shipped) and not widened and positives_hold
        print(
            f"[{'ok' if real else 'FAIL'}] draft2020-12 | shipped {receipt} rejects the"
            f" artifact {label} admits ({len(shipped)} errors), mutated variant accepts it"
            f" ({len(widened)} errors), positive case still validates: {positives_hold}"
        )
        results.append(real)

    print("PROBE", "PASS" if all(results) else "FAIL")
    return 0 if all(results) else 1


if __name__ == "__main__":
    raise SystemExit(main())
