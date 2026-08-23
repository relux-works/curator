#!/usr/bin/env python3
"""Independent closed-schema oracle; imports no Curator code or mutable state."""
import hashlib
import json
import sys
from pathlib import Path


def exact(obj, path, *keys):
    assert isinstance(obj, dict), f"{path} must be an object"
    assert set(obj) == set(keys), f"{path} has missing or unknown fields: {set(obj) ^ set(keys)}"


def record_id(label, payload):
    ccj = json.dumps(payload, ensure_ascii=False, separators=(",", ":"), sort_keys=True)
    return "sha256:" + hashlib.sha256(label.encode() + b"\0" + ccj.encode()).hexdigest(), ccj


def envelope(label, payload):
    identifier, _ = record_id(label, payload)
    return {"id": identifier, "label": label, "payload": payload}


def canonical_record(label, payload):
    identifier, ccj = record_id(label, payload)
    return {"ccj": ccj, "id": identifier, "label": label}


def validate_shared_record(record, label, fields, path):
    exact(record, path, "id", "label", "payload")
    assert record["label"] == label
    exact(record["payload"], f"{path}.payload", *fields)
    assert record_id(label, record["payload"])[0] == record["id"]


def derive_p10_shared(case):
    fixture = case["input"]
    assert case["id"] == "P10" and len(fixture["packages"]) == 1 and len(fixture["targets"]) == 2
    assert fixture["reuse_attempt"] is not None
    artifact_id = record_id("nodesource-test-v1", {"seed": "python-p10-artifact"})[0]
    declaration_id = record_id("nodesource-test-v1", {"seed": "python-p10-declaration"})[0]
    product = envelope("curator-node-v1", {
        "kind": "command_product", "logical_key": "python:portable", "payload": {
            "command_key": "portable", "declaration_digest": declaration_id,
            "entry_point_contract": "python-module", "profile": "python-source-v1", "skill_key": "python-protocol",
        },
    })
    package = envelope("curator-node-v1", {
        "kind": "package_instance", "logical_key": "python:portable@1", "payload": {
            "artifact_manifest_id": artifact_id, "ecosystem": "python", "lock_instance_key": "portable@1",
            "name": "portable", "origin": "pylock://portable/1", "profile": "python-source-v1",
            "trust_role": "dependency_input", "version": "1",
        },
    })
    requires = envelope("curator-edge-v1", {
        "edge_key": "python:portable:requires", "from_node_id": product["id"], "kind": "requires",
        "payload": {"dependency_kind": "runtime", "origin": {"field": "packages[portable]"}, "scope": "runtime"},
        "to_node_id": package["id"],
    })
    graph_records = sorted([product, package, requires], key=lambda value: value["id"])
    capture = envelope("curator-capture-graph-v1", {
        "artifact_manifest_ids": [artifact_id], "edge_ids": [requires["id"]],
        "node_ids": sorted([product["id"], package["id"]]), "policy_ids": ["curator-artifact-policy-v1"],
        "profile_id": "python-source-v1", "root_node_ids": [product["id"]], "schema_id": "closure-capture-graph-v1",
    })
    validate_shared_record(capture, "curator-capture-graph-v1",
                           ("artifact_manifest_ids", "edge_ids", "node_ids", "policy_ids", "profile_id", "root_node_ids", "schema_id"), "P10.capture")
    assert capture["payload"]["node_ids"] == sorted(capture["payload"]["node_ids"])
    assert capture["payload"]["edge_ids"] == sorted(capture["payload"]["edge_ids"])
    assert set(capture["payload"]["root_node_ids"]) <= set(capture["payload"]["node_ids"])
    assert requires["payload"]["from_node_id"] in capture["payload"]["node_ids"]
    assert requires["payload"]["to_node_id"] in capture["payload"]["node_ids"]

    outcomes, binding_ids, active_ids, bindings_by_key = [], [], [], {}
    for target in fixture["targets"]:
        exact(target, "P10.target", "interpreter", "platform", "abi")
        target_key = f'{target["interpreter"]}/{target["platform"]}/{target["abi"]}'
        architecture, libc, triple = "x86_64", "glibc", "x86_64-unknown-linux-gnu"
        if target["platform"] == "darwin":
            architecture, libc, triple = "arm64", "libSystem", "arm64-apple-darwin"
        platform = envelope("curator-node-v1", {
            "kind": "target_platform", "logical_key": "python-target:" + target_key, "payload": {
                "abi": target["abi"], "architecture": architecture, "language_modes": {"python": target["interpreter"]},
                "libc": libc, "minimum_runtime": "profile-v1", "os": target["platform"], "runtime": target["interpreter"],
                "sdk_id": "python-runtime-v1", "target_triple": triple, "tuning": {},
            },
        })
        selection = envelope("curator-selection-context-v1", {
            "default_features": True, "evaluator_ids": ["python-marker-v1"], "features": [],
            "markers": {"python_abi": target["abi"], "python_interpreter": target["interpreter"], "sys_platform": target["platform"]},
            "peer_context": {}, "platform_roles": {"target": platform["id"]}, "product_node_ids": [product["id"]],
            "schema_id": "closure-selection-context-v1",
        })
        target_edge = envelope("curator-edge-v1", {
            "edge_key": "python:portable:target", "from_node_id": product["id"], "kind": "targets",
            "payload": {"binding_role": "target", "origin": {"field": "selection.platform_roles.target"}},
            "to_node_id": platform["id"],
        })
        binding = envelope("curator-selection-binding-v1", {
            "binding_edge_ids": [target_edge["id"]], "binding_node_ids": [platform["id"]],
            "captured_graph_id": capture["id"], "schema_id": "closure-selection-binding-v1",
            "selection_context_id": selection["id"],
        })
        active = envelope("curator-active-graph-v1", {
            "captured_graph_id": capture["id"], "edge_activations": [],
            "node_activations": sorted([{"node_id": product["id"], "state": "selected"}, {"node_id": package["id"], "state": "selected"}], key=lambda value: value["node_id"]),
            "non_ordering_sccs": [], "schema_id": "closure-active-graph-v1",
            "selection_binding_id": binding["id"], "selection_context_id": selection["id"],
        })
        validate_shared_record(selection, "curator-selection-context-v1",
                               ("default_features", "evaluator_ids", "features", "markers", "peer_context", "platform_roles", "product_node_ids", "schema_id"), f"P10.{target_key}.selection")
        validate_shared_record(binding, "curator-selection-binding-v1",
                               ("binding_edge_ids", "binding_node_ids", "captured_graph_id", "schema_id", "selection_context_id"), f"P10.{target_key}.binding")
        validate_shared_record(active, "curator-active-graph-v1",
                               ("captured_graph_id", "edge_activations", "node_activations", "non_ordering_sccs", "schema_id", "selection_binding_id", "selection_context_id"), f"P10.{target_key}.active")
        assert selection["payload"]["platform_roles"]["target"] == platform["id"]
        assert binding["payload"]["captured_graph_id"] == capture["id"]
        assert binding["payload"]["selection_context_id"] == selection["id"]
        assert binding["payload"]["binding_node_ids"] == [platform["id"]]
        assert binding["payload"]["binding_edge_ids"] == [target_edge["id"]]
        assert active["payload"]["captured_graph_id"] == capture["id"]
        assert active["payload"]["selection_context_id"] == selection["id"]
        assert active["payload"]["selection_binding_id"] == binding["id"]
        assert active["payload"]["node_activations"] == sorted(active["payload"]["node_activations"], key=lambda value: value["node_id"])
        target_graph_records = sorted(graph_records + [platform, target_edge], key=lambda value: value["id"])
        assert set(binding["payload"]["binding_node_ids"] + binding["payload"]["binding_edge_ids"]) <= {record["id"] for record in target_graph_records}
        payload = {
            "active_graph": active, "case_id": "P10", "capture_graph": capture, "decision": "admit", "diagnostic": None,
            "graph_records": target_graph_records, "schema_id": "python-protocol-target-outcome-v1", "selection_binding": binding,
            "selection_context": selection, "target_key": target_key,
        }
        outcomes.append(canonical_record("python-protocol-target-outcome-v1", payload))
        binding_ids.append(binding["id"])
        active_ids.append(active["id"])
        assert target_key not in bindings_by_key
        bindings_by_key[target_key] = binding["id"]

    attempt = fixture["reuse_attempt"]
    exact(attempt, "P10.reuse_attempt", "from_target", "to_target")
    source, destination = bindings_by_key[attempt["from_target"]], bindings_by_key[attempt["to_target"]]
    assert source != destination
    diagnostic = {"code": "closure_target_identity_changed", "fields": {"from_binding_id": source, "to_binding_id": destination}, "subject": "python-target-binding"}
    exact(diagnostic, "P10.reuse.diagnostic", "code", "fields", "subject")
    exact(diagnostic["fields"], "P10.reuse.diagnostic.fields", "from_binding_id", "to_binding_id")
    reuse = canonical_record("python-protocol-cross-target-reuse-v1", {
        "case_id": "P10", "diagnostic": diagnostic, "from_selection_binding_id": source,
        "schema_id": "python-protocol-cross-target-reuse-v1", "to_selection_binding_id": destination,
    })
    summary = {
        "decision": "admit", "diagnostic_code": "", "capture_graph_id": capture["id"],
        "binding_ids": sorted(binding_ids), "active_graph_ids": sorted(active_ids),
        "reuse_diagnostic_code": "closure_target_identity_changed",
    }
    return outcomes, reuse, summary


corpus = json.loads((Path(__file__).parent / "python_protocol_shared_records.json").read_text())
exact(corpus, "corpus", "schema_id", "records", "cases")
assert corpus["schema_id"] == "node-python-protocol-golden-v3"

records = {record["id"]: record for record in corpus["records"]}
for index, record in enumerate(corpus["records"]):
    exact(record, f"records[{index}]", "id", "label", "name", "payload")
    assert record_id(record["label"], record["payload"])[0] == record["id"], record["name"]

capture = next(record["payload"] for record in corpus["records"] if record["name"] == "cgp05.capture")
assert capture["schema_id"] == "closure-capture-graph-v1"
assert capture["node_ids"] == sorted(capture["node_ids"])
assert capture["edge_ids"] == sorted(capture["edge_ids"])
assert all(records[value]["label"] == "curator-node-v1" for value in capture["node_ids"])
assert all(records[value]["label"] == "curator-edge-v1" for value in capture["edge_ids"])
for edge_id in capture["edge_ids"]:
    edge = records[edge_id]["payload"]
    assert edge["from_node_id"] in capture["node_ids"] and edge["to_node_id"] in capture["node_ids"]


def derive(case):
    exact(case, case.get("id", "case"), "id", "input", "expected")
    fixture = case["input"]
    exact(fixture, f'{case["id"]}.input', "schema_id", "packages", "lock", "artifact", "build", "targets", "reuse_attempt")
    assert fixture["schema_id"] == "python-protocol-fixture-v2"
    exact(fixture["lock"], f'{case["id"]}.lock', "format_supported", "hashes_complete", "graph_complete", "local_path_escape")
    exact(fixture["artifact"], f'{case["id"]}.artifact', "class", "record_valid")
    exact(fixture["build"], f'{case["id"]}.build', "dependencies_locked", "metadata_matches", "network_attempted", "native_build")
    assert isinstance(fixture["packages"], list) and fixture["packages"]
    assert isinstance(fixture["targets"], list) and fixture["targets"]

    packages = {}
    for index, package in enumerate(fixture["packages"]):
        exact(package, f'{case["id"]}.packages[{index}]', "name", "version", "dependencies")
        assert package["name"] and package["version"] and isinstance(package["dependencies"], list)
        package_id = record_id("python-protocol-package-node-v1", {"kind": "package_instance", "name": package["name"], "version": package["version"]})[0]
        assert package["name"] not in packages
        packages[package["name"]] = package_id
    node_ids = sorted(packages.values())
    edge_ids = []
    for package in fixture["packages"]:
        for dependency in package["dependencies"]:
            assert dependency in packages
            edge_ids.append(record_id("python-protocol-requires-edge-v1", {"from_node_id": packages[package["name"]], "kind": "requires", "to_node_id": packages[dependency]})[0])
    edge_ids.sort()
    capture_payload = {"schema_id": "python-protocol-capture-graph-v1", "node_ids": node_ids, "edge_ids": edge_ids}
    capture_id = record_id("python-protocol-capture-graph-v1", capture_payload)[0]

    target_bindings = {}
    binding_ids, active_ids = [], []
    bindings, active_graphs = [], []
    for index, target in enumerate(fixture["targets"]):
        exact(target, f'{case["id"]}.targets[{index}]', "interpreter", "platform", "abi")
        assert all(target.values())
        target_key = f'{target["interpreter"]}/{target["platform"]}/{target["abi"]}'
        target_id = record_id("python-protocol-target-node-v1", {"kind": "target_platform", **target})[0]
        binding_payload = {"schema_id": "python-protocol-selection-binding-v1", "captured_graph_id": capture_id, "target_node_id": target_id}
        binding_id = record_id("python-protocol-selection-binding-v1", binding_payload)[0]
        active_payload = {"schema_id": "python-protocol-active-graph-v1", "captured_graph_id": capture_id, "selection_binding_id": binding_id, "node_ids": node_ids, "edge_ids": edge_ids}
        active_id = record_id("python-protocol-active-graph-v1", active_payload)[0]
        assert target_key not in target_bindings
        target_bindings[target_key] = binding_id
        binding_ids.append(binding_id)
        active_ids.append(active_id)
        bindings.append({"label": "python-protocol-selection-binding-v1", "id": binding_id, "payload": binding_payload})
        active_graphs.append({"label": "python-protocol-active-graph-v1", "id": active_id, "payload": active_payload})
    binding_ids.sort()
    active_ids.sort()
    bindings.sort(key=lambda record: record["id"])
    active_graphs.sort(key=lambda record: record["id"])

    lock, artifact, build = fixture["lock"], fixture["artifact"], fixture["build"]
    diagnostic = None
    process_started = False
    if not lock["format_supported"]:
        diagnostic = {"code": "closure_lock_format_unsupported", "phase": "C1.resolve"}
    elif lock["local_path_escape"]:
        diagnostic = {"code": "closure_local_path_escape", "phase": "C1.resolve"}
    elif not lock["hashes_complete"]:
        diagnostic = {"code": "closure_integrity_missing", "phase": "C1.resolve"}
    elif not lock["graph_complete"]:
        diagnostic = {"code": "closure_graph_incomplete", "phase": "C1.resolve"}
    elif artifact["class"] in {"native.shared-library", "python.bytecode"}:
        diagnostic = {"code": "artifact_compiled_dependency_forbidden", "phase": "C3.admit"}
    elif not artifact["record_valid"]:
        diagnostic = {"code": "closure_metadata_mismatch", "phase": "C3.admit"}
    elif not build["dependencies_locked"]:
        diagnostic = {"code": "closure_build_dependency_unlocked", "phase": "C5.plan"}
    elif build["native_build"]:
        diagnostic = {"code": "closure_native_build_unsupported", "phase": "C5.plan"}
    else:
        process_started = True
        if build["network_attempted"]:
            diagnostic = {"code": "closure_network_attempted", "phase": "C6.offline"}
        elif not build["metadata_matches"]:
            diagnostic = {"code": "closure_metadata_mismatch", "phase": "C7.publish"}

    admitted = diagnostic is None
    if not admitted:
        binding_ids, active_ids = [], []
        bindings, active_graphs = [], []
    diagnostic_record = None
    if diagnostic is not None:
        diagnostic_payload = {"schema_id": "python-protocol-diagnostic-v1", **diagnostic}
        diagnostic_id = record_id("python-protocol-diagnostic-v1", diagnostic_payload)[0]
        diagnostic_record = {"label": "python-protocol-diagnostic-v1", "id": diagnostic_id, "payload": diagnostic_payload}
    reuse_diagnostic = None
    reuse_code = ""
    attempt = fixture["reuse_attempt"]
    if attempt is not None:
        exact(attempt, f'{case["id"]}.reuse_attempt', "from_target", "to_target")
        source, destination = target_bindings[attempt["from_target"]], target_bindings[attempt["to_target"]]
        assert source != destination
        reuse_code = "closure_target_identity_changed"
        reuse_payload = {"schema_id": "python-protocol-diagnostic-v1", "code": reuse_code, "phase": "C4.close", "from_binding_id": source, "to_binding_id": destination}
        reuse_id = record_id("python-protocol-diagnostic-v1", reuse_payload)[0]
        reuse_diagnostic = {"label": "python-protocol-diagnostic-v1", "id": reuse_id, "payload": reuse_payload}

    payload = {
        "schema_id": "python-protocol-outcome-v2", "case_id": case["id"],
        "decision": "admit" if admitted else "reject", "diagnostic": diagnostic_record,
        "capture_graph": {"label": "python-protocol-capture-graph-v1", "id": capture_id, "payload": capture_payload},
        "bindings": bindings, "active_graphs": active_graphs, "reuse_diagnostic": reuse_diagnostic,
        "process_started": process_started, "published": admitted,
    }
    outcome_id, ccj = record_id("python-protocol-outcome-v1", payload)
    summary = {
        "decision": payload["decision"], "diagnostic_code": "" if diagnostic is None else diagnostic["code"],
        "capture_graph_id": capture_id, "binding_ids": binding_ids, "active_graph_ids": active_ids,
        "reuse_diagnostic_code": reuse_code,
    }
    return {"ccj": ccj, "id": outcome_id, "label": "python-protocol-outcome-v1", "summary": summary}


assert [case["id"] for case in corpus["cases"]] == [f"P{index:02d}" for index in range(1, 14)]

# Prove the independent decoder rejects both omissions and extensions in every
# security-relevant nested schema rather than merely accepting the golden file.
for nested in ("lock", "artifact", "build"):
    for add_unknown in (False, True):
        mutated = json.loads(json.dumps(corpus["cases"][0]))
        if add_unknown:
            mutated["input"][nested]["unknown_field"] = True
        else:
            del mutated["input"][nested][next(iter(mutated["input"][nested]))]
        rejected = False
        try:
            derive(mutated)
        except (AssertionError, KeyError):
            rejected = True
        assert rejected, f"Python accepted malformed nested {nested} schema"

p10_case = next(case for case in corpus["cases"] if case["id"] == "P10")
p10_outcomes, p10_reuse, _ = derive_p10_shared(p10_case)
p10_payload = json.loads(p10_outcomes[0]["ccj"])
shared_schemas = {
    "capture_graph": ("curator-capture-graph-v1", ("artifact_manifest_ids", "edge_ids", "node_ids", "policy_ids", "profile_id", "root_node_ids", "schema_id")),
    "selection_context": ("curator-selection-context-v1", ("default_features", "evaluator_ids", "features", "markers", "peer_context", "platform_roles", "product_node_ids", "schema_id")),
    "selection_binding": ("curator-selection-binding-v1", ("binding_edge_ids", "binding_node_ids", "captured_graph_id", "schema_id", "selection_context_id")),
    "active_graph": ("curator-active-graph-v1", ("captured_graph_id", "edge_activations", "node_activations", "non_ordering_sccs", "schema_id", "selection_binding_id", "selection_context_id")),
}
for name, (label, fields) in shared_schemas.items():
    for add_unknown in (False, True):
        mutated = json.loads(json.dumps(p10_payload[name]))
        if add_unknown:
            mutated["payload"]["unknown_field"] = True
        else:
            del mutated["payload"]["schema_id"]
        rejected = False
        try:
            validate_shared_record(mutated, label, fields, f"P10.mutated.{name}")
        except AssertionError:
            rejected = True
        assert rejected, f"Python accepted malformed shared {name} schema"

reuse_payload = json.loads(p10_reuse["ccj"])
exact(reuse_payload["diagnostic"], "P10.reuse.diagnostic", "code", "fields", "subject")
for add_unknown in (False, True):
    mutated = json.loads(json.dumps(reuse_payload["diagnostic"]))
    if add_unknown:
        mutated["unknown_field"] = True
    else:
        del mutated["subject"]
    rejected = False
    try:
        exact(mutated, "P10.mutated.diagnostic", "code", "fields", "subject")
    except AssertionError:
        rejected = True
    assert rejected, "Python accepted malformed shared diagnostic schema"

for case in corpus["cases"]:
	if case["id"] == "P10":
		target_outcomes, reuse_negative, summary = derive_p10_shared(case)
		if "--emit-expected" in sys.argv:
			print(json.dumps({"id": case["id"], "expected": summary}, separators=(",", ":"), sort_keys=True))
		else:
			exact(case["expected"], f'{case["id"]}.expected', "decision", "diagnostic_code", "capture_graph_id", "binding_ids", "active_graph_ids", "reuse_diagnostic_code")
			assert case["expected"] == summary, case["id"]
			print(json.dumps({"case_id": case["id"], "outcome": None, "target_outcomes": target_outcomes, "reuse_negative": reuse_negative}, separators=(",", ":"), sort_keys=True))
		continue
	actual = derive(case)
	if "--emit-expected" in sys.argv:
		print(json.dumps({"id": case["id"], "expected": actual["summary"]}, separators=(",", ":"), sort_keys=True))
	else:
		exact(case["expected"], f'{case["id"]}.expected', "decision", "diagnostic_code", "capture_graph_id", "binding_ids", "active_graph_ids", "reuse_diagnostic_code")
		assert case["expected"] == actual["summary"], case["id"]
		print(json.dumps({"case_id": case["id"], "outcome": {"ccj": actual["ccj"], "id": actual["id"], "label": actual["label"]}, "target_outcomes": [], "reuse_negative": None}, separators=(",", ":"), sort_keys=True))
