# Review execution evidence

## readiness.log

```text
git version 2.54.0 (Apple Git-157)
go version go1.25.5 darwin/arm64
```

## focused.log

```text
ok  	github.com/relux-works/curator/internal/envprofile	25.436s
ok  	github.com/relux-works/curator/internal/transaction	1.041s
ok  	github.com/relux-works/curator/internal/managerlock	0.615s
```

## R1.log

```text
ok  	github.com/relux-works/curator/internal/envprofile	5.979s
ok  	github.com/relux-works/curator/internal/transaction	0.648s
ok  	github.com/relux-works/curator/internal/managerlock	0.260s
```

## R2.log

```text
ok  	github.com/relux-works/curator/internal/envprofile	3.866s
ok  	github.com/relux-works/curator/internal/transaction	0.181s [no tests to run]
ok  	github.com/relux-works/curator/internal/managerlock	0.197s
```

## R3.log

```text
ok  	github.com/relux-works/curator/internal/envprofile	3.189s
ok  	github.com/relux-works/curator/internal/transaction	0.158s [no tests to run]
ok  	github.com/relux-works/curator/internal/managerlock	0.149s [no tests to run]
```

## R4.log

```text
ok  	github.com/relux-works/curator/internal/envprofile	13.007s
ok  	github.com/relux-works/curator/internal/transaction	0.154s [no tests to run]
ok  	github.com/relux-works/curator/internal/managerlock	0.145s [no tests to run]
```

## R5.log

```text
ok  	github.com/relux-works/curator/internal/envprofile	0.973s
ok  	github.com/relux-works/curator/internal/transaction	0.153s [no tests to run]
ok  	github.com/relux-works/curator/internal/managerlock	0.144s [no tests to run]
```

## M3.log

```text
--- FAIL: TestMaterializeCaptureSameFragmentDrift (0.93s)
    --- FAIL: TestMaterializeCaptureSameFragmentDrift/aba (0.93s)
        materialize_capture_test.go:409: capture = [non-nil captured output omitted], <nil>; want environment_materialize_capture_drift and no output
FAIL
FAIL	github.com/relux-works/curator/internal/envprofile	1.254s
FAIL
```

## free-hunt.log

```text
ok  	github.com/relux-works/curator/internal/envprofile	3.197s
```
