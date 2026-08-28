# frozen_string_literal: true

require "digest"
require "json"
require "set"

path = ARGV.fetch(0, ".research/260811_cross-language-closure-graph-and-checkpoints.md")
lines = File.readlines(path, chomp: true)

def canonical(value)
  case value
  when Hash
    "{" + value.keys.map(&:to_s).sort.map { |key| JSON.generate(key) + ":" + canonical(value.fetch(key)) }.join(",") + "}"
  when Array
    "[" + value.map { |item| canonical(item) }.join(",") + "]"
  when String
    JSON.generate(value)
  when Integer
    value.to_s
  when TrueClass
    "true"
  when FalseClass
    "false"
  when NilClass
    "null"
  else
    raise "unsupported canonical value #{value.class}"
  end
end

records = {}
by_id = {}
errors = []

lines.each_index do |index|
  next unless lines[index].start_with?("name=")

  name = lines[index].delete_prefix("name=")
  label_line = lines[index + 1]
  payload_line = lines[index + 2]
  hash_line = lines[index + 3]

  unless label_line&.start_with?("label=") && payload_line&.start_with?("{") && hash_line&.match?(/\Asha256:[0-9a-f]{64}\z/)
    errors << "#{name}: malformed four-line record"
    next
  end

  label = label_line.delete_prefix("label=")
  begin
    payload = JSON.parse(payload_line)
  rescue JSON::ParserError => error
    errors << "#{name}: JSON parse failed: #{error.message}"
    next
  end

  ccj = canonical(payload)
  errors << "#{name}: payload is not canonical JSON" unless ccj == payload_line
  actual_id = "sha256:" + Digest::SHA256.hexdigest(label + "\0" + ccj)
  errors << "#{name}: expected #{hash_line}, derived #{actual_id}" unless actual_id == hash_line
  errors << "#{name}: duplicate fixture name" if records.key?(name)

  record = { "label" => label, "payload" => payload, "id" => actual_id }
  records[name] = record
  by_id[actual_id] ||= record
end

required_names = %w[
  cgp05.capture
  cgp05.platform.darwin
  cgp05.platform.linux
  cgp05.selection.darwin
  cgp05.selection.linux
  cgp05.targets.darwin
  cgp05.targets.linux
  cgp05.binding.darwin
  cgp05.binding.linux
  cgp05.active.darwin
  cgp05.active.linux
  cgp05.plan.darwin
  cgp05.plan.linux
  cgp05.c4.darwin
  cgp05.c4.linux
  cgp05.c5.darwin
  cgp05.c5.linux
  cgp10.artifact-manifest
  cgp10.product
  cgp10.source
  cgp10.action
  cgp10.output
  cgp10.declares
  cgp10.reads
  cgp10.produces
  cgp10.publishes
  cgp10.capture
  cgp10.platform
  cgp10.toolchain
  cgp10.selection
  cgp10.uses-tool
  cgp10.targets.product
  cgp10.targets.action
  cgp10.targets.toolchain
  cgp10.targets.output
  cgp10.binding
  cgp10.active
  cgp10.plan
  cgp10.c4
  cgp10.c5
  cgp10.closure
  cgp10.expected-cache-input
  cgp10.observation.one
  cgp10.observation.two
  cgp10.execution.one
  cgp10.execution.two
  cgp10.publication.one
  cgp10.publication.two
]

missing = required_names.reject { |name| records.key?(name) }
errors << "missing records: #{missing.join(', ')}" unless missing.empty?

resolve = lambda do |id, expected_label = nil|
  record = by_id[id]
  errors << "unresolved fixture record #{id}" unless record
  if record && expected_label && record["label"] != expected_label
    errors << "wrong record label for #{id}: #{record['label']}, expected #{expected_label}"
  end
  record
end

unless missing.any?
  %w[cgp05 cgp10].each do |prefix|
    capture = records.fetch("#{prefix}.capture")["payload"]
    node_ids = capture.fetch("node_ids")
    edge_ids = capture.fetch("edge_ids")

    errors << "#{prefix}: capture node IDs are not sorted" unless node_ids == node_ids.sort
    errors << "#{prefix}: capture edge IDs are not sorted" unless edge_ids == edge_ids.sort
    node_ids.each { |id| resolve.call(id, "curator-node-v1") }
    edge_ids.each { |id| resolve.call(id, "curator-edge-v1") }

    edge_ids.each do |id|
      edge = by_id.fetch(id)["payload"]
      errors << "#{prefix}: capture edge has dangling from endpoint" unless node_ids.include?(edge["from_node_id"])
      errors << "#{prefix}: capture edge has dangling to endpoint" unless node_ids.include?(edge["to_node_id"])
    end
    errors << "#{prefix}: capture root is dangling" unless (capture.fetch("root_node_ids") - node_ids).empty?
  end

  binding_cases = [
    ["cgp05.binding.darwin", "cgp05.selection.darwin", "cgp05.capture"],
    ["cgp05.binding.linux", "cgp05.selection.linux", "cgp05.capture"],
    ["cgp10.binding", "cgp10.selection", "cgp10.capture"]
  ]

  binding_cases.each do |binding_name, selection_name, capture_name|
    binding = records.fetch(binding_name)["payload"]
    selection = records.fetch(selection_name)["payload"]
    capture = records.fetch(capture_name)["payload"]
    binding_node_ids = binding.fetch("binding_node_ids")
    binding_edge_ids = binding.fetch("binding_edge_ids")

    errors << "#{binding_name}: binding node IDs are not sorted" unless binding_node_ids == binding_node_ids.sort
    errors << "#{binding_name}: binding edge IDs are not sorted" unless binding_edge_ids == binding_edge_ids.sort

    binding_node_ids.each do |id|
      record = resolve.call(id, "curator-node-v1")
      next unless record

      errors << "#{binding_name}: forbidden binding node kind" unless %w[target_platform toolchain_component].include?(record["payload"]["kind"])
    end
    binding_edge_ids.each { |id| resolve.call(id, "curator-edge-v1") }

    all_node_ids = (capture.fetch("node_ids") + binding_node_ids).to_set
    binding_edge_ids.each do |id|
      edge = by_id.fetch(id)["payload"]
      errors << "#{binding_name}: forbidden binding edge kind" unless %w[targets uses_tool requires provides_interop].include?(edge["kind"])
      errors << "#{binding_name}: binding edge has dangling from endpoint" unless all_node_ids.include?(edge["from_node_id"])
      errors << "#{binding_name}: binding edge has dangling to endpoint" unless all_node_ids.include?(edge["to_node_id"])
    end

    target_id = selection.fetch("platform_roles").fetch("target")
    target_record = resolve.call(target_id, "curator-node-v1")
    errors << "#{binding_name}: target platform is absent from binding" unless binding_node_ids.include?(target_id)
    errors << "#{binding_name}: target role resolves to wrong kind" unless target_record && target_record["payload"]["kind"] == "target_platform"

    target_edges = binding_edge_ids.filter_map { |id| by_id[id]&.fetch("payload") }
      .select { |edge| edge["kind"] == "targets" && edge["to_node_id"] == target_id }
    errors << "#{binding_name}: no explicit targets edge" if target_edges.empty?
  end

  cgp05_capture_id = records.fetch("cgp05.capture")["id"]
  errors << "CGP05: Darwin binding changed capture" unless records.fetch("cgp05.binding.darwin")["payload"]["captured_graph_id"] == cgp05_capture_id
  errors << "CGP05: Linux binding changed capture" unless records.fetch("cgp05.binding.linux")["payload"]["captured_graph_id"] == cgp05_capture_id
  %w[platform selection binding active plan c4 c5].each do |kind|
    errors << "CGP05: #{kind} branch IDs unexpectedly equal" if records.fetch("cgp05.#{kind}.darwin")["id"] == records.fetch("cgp05.#{kind}.linux")["id"]
  end

  action_id = records.fetch("cgp10.action")["id"]
  slot_edges = %w[cgp10.reads cgp10.produces cgp10.uses-tool].map { |name| records.fetch(name)["payload"] }
  errors << "CGP10: read slot is not bound exactly once" unless slot_edges.count { |edge| edge["kind"] == "reads" && edge["from_node_id"] == action_id && edge["payload"]["read_slot"] == "src" } == 1
  errors << "CGP10: write slot is not bound exactly once" unless slot_edges.count { |edge| edge["kind"] == "produces" && edge["from_node_id"] == action_id && edge["payload"]["write_slot"] == "cli" } == 1
  errors << "CGP10: tool slot is not bound exactly once" unless slot_edges.count { |edge| edge["kind"] == "uses_tool" && edge["from_node_id"] == action_id && edge["payload"]["tool_slot"] == "compiler" } == 1

  c4 = records.fetch("cgp10.c4")
  c5 = records.fetch("cgp10.c5")
  closure = records.fetch("cgp10.closure")
  errors << "CGP10: C4 active reference mismatch" unless c4["payload"]["payload"]["active_graph_id"] == records.fetch("cgp10.active")["id"]
  errors << "CGP10: C4 capture reference mismatch" unless c4["payload"]["payload"]["captured_graph_id"] == records.fetch("cgp10.capture")["id"]
  errors << "CGP10: C4 binding reference mismatch" unless c4["payload"]["payload"]["selection_binding_id"] == records.fetch("cgp10.binding")["id"]
  errors << "CGP10: C5 predecessor mismatch" unless c5["payload"]["previous_checkpoint_id"] == c4["id"]
  errors << "CGP10: C5 plan reference mismatch" unless c5["payload"]["payload"]["build_plan_id"] == records.fetch("cgp10.plan")["id"]
  errors << "CGP10: closure reference mismatch" unless closure["payload"]["c5_checkpoint_id"] == c5["id"]

  %w[one two].each do |branch|
    observation = records.fetch("cgp10.observation.#{branch}")
    execution = records.fetch("cgp10.execution.#{branch}")
    publication = records.fetch("cgp10.publication.#{branch}")

    errors << "CGP10 #{branch}: expected output mismatch" unless observation["payload"]["expected_output_node_id"] == records.fetch("cgp10.output")["id"]
    errors << "CGP10 #{branch}: produces edge mismatch" unless observation["payload"]["produces_edge_id"] == records.fetch("cgp10.produces")["id"]
    errors << "CGP10 #{branch}: execution closure mismatch" unless execution["payload"]["closure_id"] == closure["id"]
    errors << "CGP10 #{branch}: execution observation mismatch" unless execution["payload"]["produced_observation_ids"] == [observation["id"]]
    errors << "CGP10 #{branch}: publication execution mismatch" unless publication["payload"]["execution_receipt_id"] == execution["id"]
    errors << "CGP10 #{branch}: expected cache input mismatch" unless publication["payload"]["expected_cache_input_id"] == records.fetch("cgp10.expected-cache-input")["id"]
  end

  errors << "CGP10: observation IDs unexpectedly equal" if records.fetch("cgp10.observation.one")["id"] == records.fetch("cgp10.observation.two")["id"]
  errors << "CGP10: execution IDs unexpectedly equal" if records.fetch("cgp10.execution.one")["id"] == records.fetch("cgp10.execution.two")["id"]
  errors << "CGP10: publication IDs unexpectedly equal" if records.fetch("cgp10.publication.one")["id"] == records.fetch("cgp10.publication.two")["id"]
end

abort(errors.join("\n")) unless errors.empty?

puts "canonical_goldens=pass labeled_records=#{records.length} cgp05_target_branches=2 cgp10_observation_branches=2"
puts "canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true"
