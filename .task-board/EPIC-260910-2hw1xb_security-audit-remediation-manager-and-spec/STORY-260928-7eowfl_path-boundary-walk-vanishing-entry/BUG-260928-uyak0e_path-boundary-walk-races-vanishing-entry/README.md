# BUG-260928-uyak0e: path-boundary-walk-races-vanishing-entry

## Description
internal/pathboundary walk: an entry listed by readdir that vanishes before lstat makes the whole path source fail regular_types (seen with .git/objects/maintenance.lock during TestPathInstallCapturesDirtyUntrackedInsideGit on macOS, run 36467710280).

## Scope
(define bug scope / affected area)

## Acceptance Criteria
Specified behaviour for a vanished entry (ENOENT after readdir) is implemented with a row; test fixtures disable git auto-maintenance (gc.auto=0, maintenance.auto=false); 20 repeated local runs green
