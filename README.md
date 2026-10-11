[![Go](https://github.com/ctx42/riff/actions/workflows/go.yml/badge.svg)](https://github.com/ctx42/riff/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/ctx42/riff.svg)](https://pkg.go.dev/github.com/ctx42/riff/pkg/riff)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE.md)

# riff

RIFF file decoder and encoder for Go.

The `riff` package reads and writes files in the Resource Interchange File
Format (RIFF), the container behind WAVE, AVI, and RMID files. It decodes a
file into a list of chunks you can inspect, edit, and encode back, and lets
you register your own decoders for chunks it does not know.

## Features

- **Built-in chunk decoders**:
  - `fmt ` — audio format
  - `data` — audio samples
  - `LIST` with `INFO` (tags) or `adtl` (`labl` and `ltxt` labels)
  - `smpl` — sampler loops
- **Unknown chunks kept**: chunks without a decoder are read as raw chunks
  (`ChunkRAWC`); with `WithLoadData` a decode, edit, encode round trip
  preserves them.
- **Metadata-only mode**: by default chunk bodies such as audio samples are
  skipped, not loaded.
- **Custom decoders**: implement the `Chunk` interface and register it in a
  `Registry`.
- **Reusable instances**: one `RIFF` value decodes many files, reusing its
  chunk decoders.

## Prerequisites

Go 1.26 or later.

## Installation

```sh
go get github.com/ctx42/riff
```

## Usage

Import the package from `github.com/ctx42/riff/pkg/riff`. More runnable
examples are in [examples_test.go](pkg/riff/examples_test.go).

### Print the audio format

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ctx42/riff/pkg/riff"
)

func main() {
	src, err := os.Open("sample.wav")
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()

	// Decode chunks, reading only the metadata (the default).
	rif := riff.New()
	if _, err = rif.ReadFrom(src); err != nil {
		log.Fatal(err)
	}

	chi := rif.Chunks().First(riff.IDfmt)
	if chi == nil {
		log.Fatal("no fmt chunk")
	}
	ch := chi.(*riff.ChunkFMT)

	fmt.Printf("Compression Code: %#04x\n", ch.CompCode)
	fmt.Printf("Channel Count: %d\n", ch.ChannelCnt)
	fmt.Printf("Sample Rate: %d\n", ch.SampleRate)
	fmt.Printf("Average Byte Rate: %d\n", ch.AvgByteRate)
	fmt.Printf("Block Align: %d\n", ch.BlockAlign)
	fmt.Printf("Bits Per Sample: %d\n", ch.BitsPerSample)
}
```

### Edit and save a file

Decode with `WithLoadData` so the chunk bodies are loaded and can be written
back.

```go
package main

import (
	"log"
	"os"

	"github.com/ctx42/riff/pkg/riff"
)

func main() {
	src, err := os.Open("sample.wav")
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()

	rif := riff.New(riff.WithLoadData())
	if _, err = rif.ReadFrom(src); err != nil {
		log.Fatal(err)
	}

	chi := rif.Chunks().First(riff.IDfmt)
	if chi == nil {
		log.Fatal("no fmt chunk")
	}
	chi.(*riff.ChunkFMT).SampleRate = 44100

	dst, err := os.Create("edited.wav")
	if err != nil {
		log.Fatal(err)
	}
	if _, err = rif.WriteTo(dst); err != nil {
		log.Fatal(err)
	}
	if err = dst.Close(); err != nil {
		log.Fatal(err)
	}
}
```

### Decode only selected chunks

Build a `Registry` with the decoders you need and pass it to `Bare`. Chunks
without a registered decoder are read as `ChunkRAWC`.

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ctx42/riff/pkg/riff"
)

func main() {
	src, err := os.Open("sample.wav")
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()

	reg := riff.NewRegistry(riff.RAWCMake())
	reg.Register(riff.IDfmt, riff.FMTMake())

	rif := riff.Bare(reg)
	if _, err = rif.ReadFrom(src); err != nil {
		log.Fatal(err)
	}

	for _, ch := range rif.Chunks() {
		fmt.Printf("%s %T\n", riff.Uint32(ch.ID()), ch)
	}
}
```

### Options

Pass options to `New`, `Bare`, or a chunk maker such as `RAWCMake`.

| Option             | Effect                                                  |
|--------------------|---------------------------------------------------------|
| `WithLoadData()`   | Load chunk bodies; required to write a file back.       |
| `WithSizeLimit(n)` | Reject chunks larger than `n` bytes with `ErrTooLarge`. |
| `WithIgnoreSize()` | Read past declared chunk sizes; correct the RIFF size.  |

## FAQ

### Can I reuse a RIFF instance for multiple files?

Yes. `ReadFrom` resets the instance before decoding, so the chunks of the
previous file are no longer valid after the next call.

### Is the library safe for concurrent use?

No. Use one `RIFF` instance per goroutine or synchronize access yourself.

### How do I decode a chunk the package does not support?

Implement the [Chunk](pkg/riff/chunk.go) interface and register a `Maker` for
its ID, as in [Decode only selected chunks](#decode-only-selected-chunks).

### What happens to unsupported chunks?

They are decoded and encoded by `ChunkRAWC`. Decode with `WithLoadData` to
keep their bodies so a decode, edit, encode round trip preserves them.

### How strict is the decoder?

It follows the RIFF specification ([doc/riffmci.pdf](doc/riffmci.pdf)) and
rejects input that breaks it, such as a pad byte that is not zero. It
tolerates these deviations, which some encoders write:

- **Missing final pad byte**: odd-sized chunk data is followed by a zero pad
  byte. When the input ends right after the data of the last chunk, also
  inside a LIST, the missing pad byte is accepted. A pad byte missing
  anywhere else is an error.
- **Odd RIFF size**: the RIFF chunk holds whole, padded chunks, so its size is
  even. Encoders that leave the final pad byte out of it write an odd size,
  which is rounded up.

`WriteTo` always writes a conforming file, with every pad byte and an even
RIFF size.

## License

`riff` is released under the [MIT License](LICENSE.md).

The RIFF specification in [doc/riffmci.pdf](doc/riffmci.pdf), _Multimedia
Programming Interface and Data Specifications 1.0_ (August 1991), is a joint
publication of IBM Corporation and Microsoft Corporation, who hold its
copyright. It is included for reference and is not covered by the MIT
License.

Some test files in [pkg/riff/testdata](pkg/riff/testdata) come from the
go-audio projects and are distributed under the Apache License 2.0; their
[README](pkg/riff/testdata/README.md) lists their sources and changes.
