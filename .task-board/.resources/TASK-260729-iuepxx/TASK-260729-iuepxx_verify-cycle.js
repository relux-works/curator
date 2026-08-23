"use strict";

const stories = {
  K: "STORY-260728-16spsm",
  S: "STORY-260728-2abmzr",
  T: "STORY-260728-2fsqtv",
  P: "STORY-260728-2mnlp0",
  R: "STORY-260728-3tymqm",
  Q: "STORY-260728-1eye8p",
};

// Exact cross-story blocked-by links among the five cycle stories.
// Direction is blocked task -> prerequisite task.
const links = [
  ["TASK-260728-168smo", "K", "TASK-260728-2spy93", "P", "hard prerequisite"],
  ["TASK-260728-168smo", "K", "TASK-260728-1g0z69", "T", "hard prerequisite"],
  ["TASK-260728-1koh5v", "K", "TASK-260728-2bu2q6", "P", "implementation ordering"],
  ["TASK-260728-1koh5v", "K", "TASK-260728-2gbtb9", "T", "implementation ordering"],
  ["TASK-260728-1uj0bc", "K", "TASK-260728-1j72zq", "T", "implementation ordering"],
  ["TASK-260728-2uh7em", "K", "TASK-260728-ypbuav", "T", "advisory sequencing"],
  ["TASK-260728-3ar1qp", "K", "TASK-260728-1j72zq", "T", "implementation ordering"],
  ["TASK-260728-gmfxdg", "K", "TASK-260728-2bu2q6", "P", "implementation ordering"],
  ["TASK-260728-gmfxdg", "K", "TASK-260728-2gbtb9", "T", "implementation ordering"],

  ["TASK-260728-1egim2", "S", "TASK-260728-ypbuav", "T", "advisory sequencing"],
  ["TASK-260728-1yhuqi", "S", "TASK-260728-2spy93", "P", "hard prerequisite"],
  ["TASK-260728-1yhuqi", "S", "TASK-260728-1g0z69", "T", "hard prerequisite"],
  ["TASK-260728-21x3yc", "S", "TASK-260728-2bu2q6", "P", "implementation ordering"],
  ["TASK-260728-21x3yc", "S", "TASK-260728-2gbtb9", "T", "implementation ordering"],
  ["TASK-260728-2lnhci", "S", "TASK-260728-2bu2q6", "P", "implementation ordering"],
  ["TASK-260728-2lnhci", "S", "TASK-260728-2gbtb9", "T", "implementation ordering"],
  ["TASK-260728-2ztr3c", "S", "TASK-260728-1j72zq", "T", "implementation ordering"],
  ["TASK-260728-3j60e3", "S", "TASK-260728-1j72zq", "T", "implementation ordering"],

  ["TASK-260728-12pnm1", "R", "TASK-260728-2spy93", "P", "hard prerequisite"],
  ["TASK-260728-12pnm1", "R", "TASK-260728-1g0z69", "T", "hard prerequisite"],
  ["TASK-260728-13ioo0", "R", "TASK-260728-2bu2q6", "P", "implementation ordering"],
  ["TASK-260728-13ioo0", "R", "TASK-260728-2gbtb9", "T", "implementation ordering"],
  ["TASK-260728-1t59zp", "R", "TASK-260728-ypbuav", "T", "advisory sequencing"],
  ["TASK-260728-2yxdo7", "R", "TASK-260728-1j72zq", "T", "implementation ordering"],
  ["TASK-260728-gjxj1v", "R", "TASK-260728-1j72zq", "T", "implementation ordering"],
  ["TASK-260728-q283m8", "R", "TASK-260728-2bu2q6", "P", "implementation ordering"],
  ["TASK-260728-q283m8", "R", "TASK-260728-2gbtb9", "T", "implementation ordering"],

  ["TASK-260728-1e6811", "T", "TASK-260728-26e3n2", "R", "qualification ordering"],
  ["TASK-260728-1e6811", "T", "TASK-260728-1y8u4m", "S", "qualification ordering"],
  ["TASK-260728-1e6811", "T", "TASK-260728-3u1nho", "K", "qualification ordering"],
  ["TASK-260728-1j72zq", "T", "TASK-260728-2bu2q6", "P", "implementation ordering"],
  ["TASK-260728-2gbtb9", "T", "TASK-260728-2bu2q6", "P", "implementation ordering"],
  ["TASK-260728-2jaw7h", "T", "TASK-260728-2spy93", "P", "hard prerequisite"],

  ["TASK-260728-251p01", "P", "TASK-260728-12pnm1", "R", "hard prerequisite"],
  ["TASK-260728-251p01", "P", "TASK-260728-1yhuqi", "S", "hard prerequisite"],
  ["TASK-260728-251p01", "P", "TASK-260728-168smo", "K", "hard prerequisite"],
  ["TASK-260728-251p01", "P", "TASK-260728-2jaw7h", "T", "hard prerequisite"],
];

function aggregate(currentLinks) {
  const map = new Map();
  for (const [, from, , to, classification] of currentLinks) {
    if (from === to) continue;
    const key = `${from}>${to}`;
    const value = map.get(key) || { from, to, count: 0, classes: new Set() };
    value.count += 1;
    value.classes.add(classification);
    map.set(key, value);
  }
  return [...map.values()].sort((a, b) => `${a.from}>${a.to}`.localeCompare(`${b.from}>${b.to}`));
}

function hasCycle(vertices, edges) {
  const adjacency = new Map(vertices.map((v) => [v, []]));
  const indegree = new Map(vertices.map((v) => [v, 0]));
  for (const { from, to } of edges) {
    adjacency.get(from).push(to);
    indegree.set(to, indegree.get(to) + 1);
  }
  const queue = vertices.filter((v) => indegree.get(v) === 0);
  let seen = 0;
  while (queue.length) {
    const current = queue.shift();
    seen += 1;
    for (const next of adjacency.get(current)) {
      indegree.set(next, indegree.get(next) - 1);
      if (indegree.get(next) === 0) queue.push(next);
    }
  }
  return seen !== vertices.length;
}

function permutations(values) {
  if (values.length <= 1) return [values];
  const result = [];
  for (let i = 0; i < values.length; i += 1) {
    const rest = values.slice(0, i).concat(values.slice(i + 1));
    for (const suffix of permutations(rest)) result.push([values[i], ...suffix]);
  }
  return result;
}

const five = ["K", "S", "R", "T", "P"];
const current = aggregate(links);
if (links.length !== 37) throw new Error(`expected 37 source links, got ${links.length}`);
if (current.length !== 14) throw new Error(`expected 14 story edges, got ${current.length}`);
if (!hasCycle(five, current)) throw new Error("current graph should be cyclic");

let minimum = Infinity;
let minimumCuts = [];
for (const order of permutations(five)) {
  const position = Object.fromEntries(order.map((value, index) => [value, index]));
  const cut = current.filter(({ from, to }) => position[from] > position[to]);
  const cost = cut.reduce((sum, edge) => sum + edge.count, 0);
  if (cost < minimum) {
    minimum = cost;
    minimumCuts = [{ order, cut }];
  } else if (cost === minimum) {
    minimumCuts.push({ order, cut });
  }
}
if (minimum !== 7) throw new Error(`expected minimum cut weight 7, got ${minimum}`);

const candidateCut = new Set(["T>K", "T>S", "T>R", "P>K", "P>S", "P>R", "P>T"]);
const candidate = current.filter(({ from, to }) => !candidateCut.has(`${from}>${to}`));
if (hasCycle(five, candidate)) throw new Error("seven-link candidate should be acyclic");

// These three directions are forced by the existing task semantics for each
// language: accepted language design before integration, qualified protocol
// before preflight implementation, and preflight before language implementation.
for (const language of ["K", "S", "R"]) {
  const mandatory = [
    { from: "P", to: language },
    { from: language, to: "T" },
    { from: "T", to: "P" },
  ];
  if (!hasCycle([language, "T", "P"], mandatory)) {
    throw new Error(`mandatory phase graph for ${language} should prove impossibility`);
  }
}

const proposedParents = new Map([
  ["TASK-260728-12pnm1", "P"],
  ["TASK-260728-1yhuqi", "P"],
  ["TASK-260728-168smo", "P"],
  ["TASK-260728-1g0z69", "P"],
  ["TASK-260728-2jaw7h", "P"],
  ["TASK-260728-1e6811", "Q"],
]);
const reparentedLinks = links.map(([fromTask, fromStory, toTask, toStory, classification]) => [
  fromTask,
  proposedParents.get(fromTask) || fromStory,
  toTask,
  proposedParents.get(toTask) || toStory,
  classification,
]);
const reparented = aggregate(reparentedLinks);
if (hasCycle(["K", "S", "R", "T", "P", "Q"], reparented)) {
  throw new Error("phase-aligned reparent proposal should be acyclic");
}

console.log(JSON.stringify({
  source_link_count: links.length,
  current_story_edge_count: current.length,
  current_story_edges: current.map(({ from, to, count, classes }) => ({
    edge: `${stories[from]} -> ${stories[to]}`,
    source_links: count,
    classes: [...classes].sort(),
  })),
  minimum_unlink_count: minimum,
  minimum_cut_story_edges: [...candidateCut].sort(),
  minimum_cut_is_acyclic: !hasCycle(five, candidate),
  mandatory_phase_cycle_proved_for: ["K", "S", "R"],
  link_unlink_only_solution_preserving_mandatory_edges: false,
  phase_aligned_reparent_story_edges: reparented.map(({ from, to, count }) => ({
    edge: `${stories[from]} -> ${stories[to]}`,
    source_links: count,
  })),
  phase_aligned_reparent_is_acyclic: !hasCycle(["K", "S", "R", "T", "P", "Q"], reparented),
}, null, 2));
