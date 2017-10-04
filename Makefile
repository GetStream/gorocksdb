ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

REPO = github.com/GetStream/gorocksdb

BENCHMARKS ?= .

DATE := $(shell date '+%Y-%m-%dT%T')

BENCH_DIR = bench_results/$(DATE)

build: cargo
	go build

release: cargo-release
	go build -ldflags="-r $(ROOT_DIR)lib" main.go

cargo:
	cd lib/hello && cargo build
	cp lib/hello/target/debug/libhello.a lib/

cargo-release:
	cd lib/hello && cargo build --release
	cp lib/hello/target/release/libhello.a lib/

run: build
	./main

clean:
	rm -rf lib/libhello.a

bench: cargo-release
	mkdir -p $(BENCH_DIR)
	{ GOMAXPROCS=1 go test $(TEST_FLAGS) -tags "$(GO_BUILD_TAGS)" $(REPO) -run '^$$' -bench='$(BENCHMARKS)' -timeout=128h | tee $(BENCH_DIR)/full; } || true
	{ grep C/ $(BENCH_DIR)/full | sed 's=C/==' > $(BENCH_DIR)/c ; } || true
	{ grep Rust/ $(BENCH_DIR)/full | sed 's=Rust/==' > $(BENCH_DIR)/rust ; } || true
	@echo
	@echo Readable benchmark results
	@echo --------------------------
	@echo
	@echo Global
	@echo ------
	@benchstat $(BENCH_DIR)/full
	@echo
	@echo C vs Rust
	@echo ---------
	@cd $(BENCH_DIR) && benchstat c rust

bench-significant:
	make bench TEST_FLAGS='-count 5'

