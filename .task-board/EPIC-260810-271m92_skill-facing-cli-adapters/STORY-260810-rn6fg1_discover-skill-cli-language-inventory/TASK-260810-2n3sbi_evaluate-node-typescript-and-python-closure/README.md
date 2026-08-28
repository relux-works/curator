# TASK-260810-2n3sbi: evaluate-node-typescript-and-python-closure

## Description
Evaluate conservative source closure for Node/TypeScript and its policy relationship with the independent Python implementation.

## Scope
Assess npm, pnpm, and Yarn lock and integrity models; package tarballs; lifecycle scripts; generated JavaScript; native addons; package-manager caches; offline installs; Python lock and requirements formats; source distributions; pure and native wheels; build backends; and runtime identity. Determine which policy and implementation layers can be shared without assuming repository co-location.

## Acceptance Criteria
The outcome recommends separate or shared Node/TypeScript and Python policy layers, proves recursive immutable source closure and offline behavior, rejects native or compiled dependency payloads and undeclared hooks, defines checkpoints and diagnostics, and records explicit unsupported cases with conformance fixtures.
