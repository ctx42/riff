This file provides guidance to AI agents (and human contributors) when working
with code in this repository.

## Commands

```sh
# Run all tests with race detection (matches CI)
go test -race ./...

# Run a single test
go test -race -run TestName ./pkg/riff/...

# Lint (config is in tmp/.golangci.yml)
golangci-lint run --config tmp/.golangci.yml ./...
```

## Architecture

`github.com/ctx42/riff` decodes and encodes files in the Resource Interchange
File Format (RIFF). The whole library lives in `pkg/riff`, a single package.

- **`RIFF`** (`riff.go`) — the top-level "RIFF" chunk. `New` registers the
  built-in decoders; `Bare` takes a caller-built `Registry`. `ReadFrom` decodes
  sub-chunks in file order, `WriteTo` encodes them back, and `Reset` returns
  decoded chunks to the registry pool so one instance can decode many files.
- **`Chunk`** (`chunk.go`) — the interface every chunk decoder implements.
  Built-in decoders: `ChunkFMT` (`fmt `), `ChunkDATA` (`data`), `ChunkLIST`
  (`LIST`, holding `INFO` or `adtl` sub-chunks), `ChunkINFO`, `ChunkLABL`,
  `ChunkLTXT`, `ChunkSMPL` (`smpl`), and `ChunkRAWC`, the fallback for chunks
  with no registered decoder.
- **`Registry`** (`registry.go`) — maps chunk IDs to `Maker` functions and pools
  decoded chunks for reuse; unknown IDs go to its raw `IDMaker`.
- **`Opts` / `ChunkReader`** (`opts.go`) — decoding options (`WithLoadData`,
  `WithSizeLimit`, `WithIgnoreSize`) and the size-limited, byte-counting reader
  each chunk decodes from.

## Key design rules

- Without `WithLoadData` decoders read only metadata and skip chunk bodies;
  `WriteTo` on a `ChunkDATA`, `ChunkLIST`, or `ChunkRAWC` decoded that way
  returns `ErrSkipDataMode`.
- Chunk sizes exclude the ID, the size field, and the pad byte; odd-sized chunks
  are padded to an even length on read and write (`RealSize`).
- Integers are little endian except chunk IDs and types, which are read big
  endian so they compare against the `ID*` and `Type*` constants.
- Tests use `github.com/ctx42/testing` and `github.com/ctx42/testkit`; sample
  files live in `pkg/riff/testdata`.
- All SPDX license headers must be preserved on every source file.
