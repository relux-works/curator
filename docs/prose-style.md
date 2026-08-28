# Curator Documentation Prose Style

Binding style rules for all Curator documentation. Synthesized from The Go Programming Language (Donovan, Kernighan), The Swift Programming Language, and engineering prose practice. Where this guide disagrees with a source book, this guide wins.

## Voice and sentences

Write in the active voice and name the actor. The software is the usual grammatical subject: the installer copies, the audit gate rejects, `curator install` resolves. Use "you" for the reader's actions and choices. Reserve the passive for rules where the agent is irrelevant: "Symlinks are not followed during hashing."

Open with a short, flat claim; let elaboration follow. Load-bearing sentences stay under roughly ten words. Longer sentences are additive chains joined by commas and "and", never nested subordination.

Lead with the goal, then the mechanism: "To pin a skill to a tag, add a `tag` field." The reader learns why before how.

State one idea per sentence. When a second idea appears, start a second sentence.

## Paragraphs

The first sentence of a paragraph carries its claim; everything after elaborates. A paragraph introduces at most one new concept. The shape is: claim, explanation, consequence or example.

Give the definition before the consequences. Name the concept, define it in one or two sentences, then use it: "A compiled command links a skill binary to an executable shim in `.agents/bin/`. `curator install` rebuilds a missing or drifted shim."

State the problem before the solution. First the drawback, then the fix.

Introduce every code or command block with a sentence that ends in a colon and says what the block shows. After a non-trivial block, add one sentence that interprets the result: what the reader should observe.

Show an install command:

```bash
curator install
```

The command installs declared skills into `.agents/skills/`.

## Terminology

Define a term at first use, then repeat it verbatim. Never rotate synonyms: a skill is "installed", not sometimes "deployed" or "provisioned". Use a pronoun only when the referent is in the same sentence; otherwise repeat the term. In technical text, repetition is precision, not a defect.

Set every identifier, command, flag, file name, and literal in code style: `curator install`, `Skillfile.json`, `.agents/skills/`.

When a term must appear before its full treatment, give a working approximation and a precise pointer: "For now, treat compiled commands as executable shims; docs/compiled-commands.md defines them."

## Punctuation and typography

Do not use em-dashes or en-dashes in prose. Replace them with a comma, a colon, parentheses, or a new sentence.

Do not use Russian guillemets. Prefer code style for identifiers; use plain double quotes only for genuine quotations and coined readings.

Use semicolons to bind two halves of one thought; the second clause explains or completes the first. Use parentheses for one-sentence asides (a concrete example, a practical note). No exclamation points, no ellipses for drama.

## Lists vs prose

Reasoning lives in prose. A list is acceptable only for genuinely parallel enumerable items: flags, file names, supported platforms, numbered install steps. Each bullet is a complete, grammatically parallel statement. Never use a list to carry an argument, and never stack lists of three-item fragments. Short enumerations (two to four items) stay inline with commas.

Numbered steps are for procedures the reader executes. Each step names the command and its observable result.

Comparative overviews may use a structured breakdown (table or per-item lead-ins with short points) when readers need to scan; reasoning inside each item stays prose.

Collapsible `<details>` blocks are allowed for parallel variants (install methods, command groups); the summary line names the variant exactly.

## Tone

State facts flatly. No marketing register: no "powerful", "seamless", "robust", "blazingly fast", no superlatives about the project itself. State costs and limitations as plainly as features, next to the claim they limit: "The whitelist strips tests from context; a skill that needs its tests at runtime must declare them as runtime roots."

Hedge only with calibrated words that carry frequency information: "usually", "in practice", "rarely". Never hedge to soften a hard rule.

Give prescriptions directly, with the rationale attached: "Always pin a tag or revision. Branch refs drift and break reproducibility."

Cross-reference with a precise target: "see Reconciliation and repair in docs/compiled-commands.md", never "as we will see later".

## Blacklist

A reviewer rejects a document that contains any of the following:

- Antithesis constructions: "it's not X, it's Y", "this isn't about X". State what the thing is; do not stage a contrast with a strawman.
  Bad: "Curator is not just a package manager, it is a skill runner."
  Good: "Curator installs skill packages into project repositories."
- Em-dashes and en-dashes used as rhetorical glue.
  Bad: "`curator` — a skill manager".
  Good: "`curator` is a skill manager."
- Russian guillemets.
  Bad: «Skillfile».
  Good: `Skillfile`.
- Chains of triple enumerations and adjective triples ("fast, simple, reliable").
  Bad: "The installer is fast, simple, and reliable."
  Good: "The installer is deterministic. The same `Skillfile.json` produces the same tree."
- A closing paragraph that restates the section just written.
  Bad: "In summary, this section introduced the installation commands."
  Good: Omit the summary paragraph.
- Filler openers: "Let's dive in", "it is important to note", "in today's world".
  Bad: "Let's dive into configuration."
  Good: "Configure the installer in `Skillfile.json`."
- Bullet lists that carry reasoning instead of parallel facts.
  Bad: Bullet list explaining why content hashing prevents security attacks.
  Good: Paragraph of prose explaining content hashing and its security impact.
- Marketing adjectives applied to the project ("powerful", "seamless", "robust", "blazingly fast").
  Bad: "A powerful and seamless skill manager."
  Good: "A tool that manages skill packages."

## Worked contrast

Bad (slop):

> Curator is not just another package manager — it's a powerful, seamless solution for skill management. Let's dive into why it's a game-changer: reproducibility, flexibility, and simplicity.

Good:

> Curator installs skill packages from git repositories into project repositories. An install is reproducible: the same `Skillfile.json` produces the same content-hashed tree on every machine.
