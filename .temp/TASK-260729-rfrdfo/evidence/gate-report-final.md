| gate | real exit | wall s | package line |
| --- | ---: | ---: | --- |
| `gate-gofmt` | 0 | 0 | `(no output)` |
| `gate-vet` | 0 | 0 | `(no output)` |
| `gate-atomicity-structure` | 0 | 286 | `--- PASS: TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder (0.01s)` |
| `gate-race-atomicity-1` | 0 | 593 | `ok  	github.com/relux-works/curator/internal/install/atomicity	591.280s` |
| `gate-race-atomicity-2` | 0 | 561 | `ok  	github.com/relux-works/curator/internal/install/atomicity	560.828s` |
| `gate-race-atomicity-3` | 0 | 564 | `ok  	github.com/relux-works/curator/internal/install/atomicity	564.022s` |
| `gate-race-install-1` | 0 | 234 | `ok  	github.com/relux-works/curator/internal/install	232.088s` |
| `gate-race-install-2` | 0 | 235 | `ok  	github.com/relux-works/curator/internal/install	235.124s` |
| `gate-race-install-3` | 0 | 227 | `ok  	github.com/relux-works/curator/internal/install	226.191s` |
| `gate-race-r5` | 0 | 46 | `ok  	github.com/relux-works/curator/internal/install	45.432s` |
| `gate-race-revalidation` | 0 | 40 | `ok  	github.com/relux-works/curator/internal/install	39.959s` |
| `gate-race-concurrency` | 0 | 19 | `ok  	github.com/relux-works/curator/internal/install	19.248s` |
| `gate-race-activation` | 0 | 37 | `ok  	github.com/relux-works/curator/internal/install/atomicity	36.499s` |

data race markers across every gate log:
  none
