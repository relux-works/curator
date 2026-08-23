# STORY-260728-16spsm: kotlin-external-build-driver

## Description
Select and implement closed local and external Kotlin CLI build drivers across curator-spec, Curator, csk and shared interoperability.

## Scope
Architecture decision between Kotlin/JVM runtime packaging and Kotlin/Native, paired local/repository driver identifiers, build-root and external-target models, toolchain preflight, offline dependency and plugin policy, manager implementations, platform qualification, conformance and guidance.

## Acceptance Criteria
The Kotlin artifact/runtime model and driver pair are explicitly selected before implementation; Gradle or other package-controlled scripts and plugins are never a generic build escape hatch; local skill and external repository fixtures match across both managers for artifact, cache, rollback and launch behavior; platform and shared gates pass.
