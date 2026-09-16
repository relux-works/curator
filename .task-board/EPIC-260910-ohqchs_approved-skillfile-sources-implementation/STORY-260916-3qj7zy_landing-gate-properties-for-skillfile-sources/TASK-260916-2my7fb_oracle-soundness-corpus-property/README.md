# TASK-260916-2my7fb: oracle-soundness-corpus-property

## Description
Every accepted frontmatter document yields exactly the reference implementation values over a generated corpus drawn from the accepted subset (scalars, folded and literal block scalars, quoting, comments, escapes limited to the corpus-backed set).

## Scope
frontmatter reader versus the reference implementation

## Acceptance Criteria
Property green over the generated corpus; the accepted subset is stated in the leaf AC and the corpus generator follows it
