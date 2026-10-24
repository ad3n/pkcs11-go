# Fork patches

Base: upstream v1.2.0-rc4 (`1cb640f1fbaa1e801f012ff047f84e2c41f709fe`). Module: `github.com/ad3n/pkcs11-go`. Go 1.25 is the minimum for the b.Loop benchmarks; validation used Go 1.27.3, Apple M3 Pro, darwin/arm64.

- `ULongToBytes` and integer `NewAttribute` call `ckULong`: encode into one owned Go byte slice, preserving native endian, ABI width and overflow rejection. No C pointer to Go storage is needed by this conversion. Returned slices do not alias; decoding remains bounded and allocation-free.
- On Apple platforms, `ck_memzero` uses the existing volatile byte-write fallback. Apple SDK does not provide explicit_bzero. Windows/BSD/glibc branches are unchanged. No secret lifetime is extended.

Validation commands (all passed except explicitly noted):

```sh
go test ./...
go vet ./...
PKCS11_MODULE=/opt/homebrew/lib/softhsm/libsofthsm2.so go test -race -gcflags=all=-d=checkptr=2 -skip '^TestIntegrationRSAOAEP$' ./...
git diff --check
cc -fsanitize=address,undefined -Icryptoki -Iinternal/headers testdata/memzero.c cryptoki/shim.c -o /private/tmp/crypto-fork-validation/memzero
/private/tmp/crypto-fork-validation/memzero
```

Unfiltered integration testing fails TestIntegrationRSAOAEP: this upstream test requires OAEP SHA-256 and the installed SoftHSM2 returns CKR_ARGUMENTS_BAD. The filtered integration run excludes exactly that test; unsupported mechanisms elsewhere retain upstream skips. SoftHSM uses temporary tokens and removes them. Hardware/network HSM, PQC providers and non-macOS ABI execution were unavailable. Betteralign and betteralign -apply returned four existing layout diagnostics; rewrites were reverted to preserve public types and avoid unrelated layout changes. Final betteralign retains those diagnostics.

Benchmarks use b.ReportAllocs and b.Loop. The baseline has the macOS compile fix only, so both sides build on the same machine; candidate adds ULONG encoding optimization. Both use Go 1.27.3, default GOMAXPROCS=12, default 1s benchmark time, five sequential samples, without other validation workloads. Benchstat was unavailable. These are exported binding API measurements, not a claim about hardware HSM throughput; crypto11 benchmarks separately cover lookup and RSA-PSS.

Command in each tree:

```sh
go test -run '^$' -bench 'Benchmark(ULongToBytes|NewIntegerAttribute)' -benchmem -count=5 ./cryptoki
```

Baseline:

```text
# github.com/eclipse-keypont/pkcs11-go/cryptoki.test
ld: warning: duplicate -rpath '/Users/aden/nwrfcsdk/lib' ignored
goos: darwin
goarch: arm64
pkg: github.com/eclipse-keypont/pkcs11-go/cryptoki
cpu: Apple M3 Pro
BenchmarkULongToBytes-12           	63683739	        18.52 ns/op	      16 B/op	       2 allocs/op
BenchmarkULongToBytes-12           	67568042	        17.74 ns/op	      16 B/op	       2 allocs/op
BenchmarkULongToBytes-12           	67934781	        17.39 ns/op	      16 B/op	       2 allocs/op
BenchmarkULongToBytes-12           	70180568	        17.54 ns/op	      16 B/op	       2 allocs/op
BenchmarkULongToBytes-12           	69620712	        17.54 ns/op	      16 B/op	       2 allocs/op
BenchmarkNewIntegerAttribute-12    	49253836	        23.91 ns/op	      48 B/op	       3 allocs/op
BenchmarkNewIntegerAttribute-12    	51210015	        23.85 ns/op	      48 B/op	       3 allocs/op
BenchmarkNewIntegerAttribute-12    	51010287	        23.75 ns/op	      48 B/op	       3 allocs/op
BenchmarkNewIntegerAttribute-12    	51036410	        23.88 ns/op	      48 B/op	       3 allocs/op
BenchmarkNewIntegerAttribute-12    	50622854	        23.81 ns/op	      48 B/op	       3 allocs/op
PASS
ok  	github.com/eclipse-keypont/pkcs11-go/cryptoki	12.393s
```

Candidate:

```text
# github.com/ad3n/pkcs11-go/cryptoki.test
ld: warning: duplicate -rpath '/Users/aden/nwrfcsdk/lib' ignored
goos: darwin
goarch: arm64
pkg: github.com/ad3n/pkcs11-go/cryptoki
cpu: Apple M3 Pro
BenchmarkULongToBytes-12           	197080471	         6.229 ns/op	       8 B/op	       1 allocs/op
BenchmarkULongToBytes-12           	200999974	         5.967 ns/op	       8 B/op	       1 allocs/op
BenchmarkULongToBytes-12           	200951644	         5.975 ns/op	       8 B/op	       1 allocs/op
BenchmarkULongToBytes-12           	201019741	         5.966 ns/op	       8 B/op	       1 allocs/op
BenchmarkULongToBytes-12           	200412348	         5.995 ns/op	       8 B/op	       1 allocs/op
BenchmarkNewIntegerAttribute-12    	92252106	        12.88 ns/op	      40 B/op	       2 allocs/op
BenchmarkNewIntegerAttribute-12    	93205092	        12.92 ns/op	      40 B/op	       2 allocs/op
BenchmarkNewIntegerAttribute-12    	95627686	        13.61 ns/op	      40 B/op	       2 allocs/op
BenchmarkNewIntegerAttribute-12    	92477530	        12.92 ns/op	      40 B/op	       2 allocs/op
BenchmarkNewIntegerAttribute-12    	92103710	        12.82 ns/op	      40 B/op	       2 allocs/op
PASS
ok  	github.com/ad3n/pkcs11-go/cryptoki	12.205s
```
