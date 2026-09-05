# TASK-260905-1wi9j6: interactive-mode-review-follow-ups

## Description
Follow-ups from review-findings-agm-1 on the landed LaunchModeInteractive (skill-agents-management main 3edbde8): (1) split the per-plugin composition negatives in claude/interactive_test.go, codex/interactive_test.go, and internal/regress/interactive_test.go into prefix-only and servers-only subcases so a Prefix&&Servers narrowing is caught per plugin; (2) decide and, if adopted, add a core ErrModelMissing so an empty Model.ID is refused in one place for every mode (pre-existing gap: exec mode admits --model ""); (3) record ErrParameterNotInteractive in a Decision 0013 erratum line when the decision is next touched (environments 1.1 batch). Schedule with implementation stage (b) when the launcher consumes the mode.

## Scope
(define task scope)

## Acceptance Criteria
Subcases split and green under make test regress; ErrModelMissing decision recorded (adopted with tests, or declined with reason); Decision 0013 erratum line filed or scheduled; PR landed by fast-forward of the reviewed head.
