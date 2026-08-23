# TASK-260729-1vpytz: research-official-go-bootstrap-guidance

## Description
Prepare a reviewed primary-source input packet for manager-owned Go toolchain installation and verification guidance, prompted by the reachable Windows validation host having no Go executable on PATH.

## Scope
Research only for Go on macOS, Windows and later Linux. Verify current official distribution channels, supported architecture/package shapes, default and operator-trusted absolute roots, post-install verification, uninstall/upgrade cautions, and no-auto-install security wording. Inspect ssh win read-only for candidate roots, but do not install software, change PATH/profile/registry, or edit product/catalog files.

## Acceptance Criteria
The artifact cites current official primary sources, gives exact platform-specific operator steps and verification commands without auto-execution, distinguishes PATH discovery from an approved absolute root and fingerprint requirement, records the actual macOS and ssh win inventory, identifies version-family compatibility with the accepted go-v1 contract, and is directly consumable by TASK-260728-ypbuav and CocoaSkills Windows qualification.
