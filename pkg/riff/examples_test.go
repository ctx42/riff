// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff_test

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/ctx42/riff/pkg/riff"
)

func Example() {
	src, err := os.Open("testdata/kick.wav")
	if err != nil {
		panic(err)
	}
	defer func() { _ = src.Close() }()

	// Decode chunks, reading only the metadata (the default).
	rif := riff.New()
	if _, err = rif.ReadFrom(src); err != nil {
		panic(err)
	}

	// The "fmt " chunk may appear only once in a RIFF file, decoding fails
	// otherwise.
	chs := rif.Chunks()
	if chs.Count(riff.IDfmt) == 0 {
		fmt.Println("no fmt chunk")
		return
	}
	ch := chs.First(riff.IDfmt).(*riff.ChunkFMT)

	extra, err := io.ReadAll(ch.Extra())
	if err != nil {
		panic(err)
	}

	fmt.Printf("Compression Code: %#04x\n", ch.CompCode)
	fmt.Printf("Channel Count: %d\n", ch.ChannelCnt)
	fmt.Printf("Sample Rate: %d\n", ch.SampleRate)
	fmt.Printf("Average Byte Rate: %d\n", ch.AvgByteRate)
	fmt.Printf("Block Align: %d\n", ch.BlockAlign)
	fmt.Printf("Bits Per Sample: %d\n", ch.BitsPerSample)
	fmt.Printf("Extra fmt Bytes: %v\n", extra)
	// Output:
	// Compression Code: 0x0001
	// Channel Count: 1
	// Sample Rate: 22050
	// Average Byte Rate: 44100
	// Block Align: 2
	// Bits Per Sample: 16
	// Extra fmt Bytes: []
}

func Example_rawChunk() {
	src, err := os.Open("testdata/kick.wav")
	if err != nil {
		panic(err)
	}
	defer func() { _ = src.Close() }()

	// Without registered decoders every chunk is decoded as ChunkRAWC, here
	// loading the chunk bodies.
	reg := riff.NewRegistry(riff.RAWCMake(riff.WithLoadData()))
	rif := riff.Bare(reg)
	if _, err = rif.ReadFrom(src); err != nil {
		panic(err)
	}

	ch := rif.Chunks().First(riff.IDfmt).(*riff.ChunkRAWC)
	body, err := io.ReadAll(ch.Body())
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", body)
	// Output:
	// [1 0 1 0 34 86 0 0 68 172 0 0 2 0 16 0]
}

func Example_dataChunk() {
	src, err := os.Open("testdata/kick.wav")
	if err != nil {
		panic(err)
	}
	defer func() { _ = src.Close() }()

	// Load the chunk data, not only the metadata.
	rif := riff.New(riff.WithLoadData())
	if _, err = rif.ReadFrom(src); err != nil {
		panic(err)
	}

	ch := rif.Chunks().First(riff.IDdata).(*riff.ChunkDATA)
	buf := make([]byte, 10)
	if _, err = io.ReadFull(ch.Data(), buf); err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", buf)
	// Output:
	// [76 0 75 0 77 0 73 0 74 0]
}

func ExampleNewRegistry() {
	src, err := os.Open("testdata/kick.wav")
	if err != nil {
		panic(err)
	}
	defer func() { _ = src.Close() }()

	// Decode chunks without a registered decoder with ChunkRAWC.
	reg := riff.NewRegistry(riff.RAWCMake())

	// Decode only "fmt " chunks with a dedicated decoder.
	reg.Register(riff.IDfmt, riff.FMTMake())

	rif := riff.Bare(reg)
	if _, err = rif.ReadFrom(src); err != nil {
		panic(err)
	}

	for _, ch := range rif.Chunks() {
		fmt.Printf("%s %T\n", riff.Uint32(ch.ID()), ch)
	}
	// Output:
	// fmt  *riff.ChunkFMT
	// data *riff.ChunkRAWC
}

func ExampleRIFF_ReadFrom_reuse() {
	// One instance can decode many files; each ReadFrom resets it first.
	rif := riff.New()

	for _, pth := range []string{"testdata/kick.wav", "testdata/bass.wav"} {
		src, err := os.Open(pth)
		if err != nil {
			panic(err)
		}

		if _, err = rif.ReadFrom(src); err != nil {
			panic(err)
		}
		_ = src.Close()

		fmt.Println(pth, len(rif.Chunks()))
	}
	// Output:
	// testdata/kick.wav 2
	// testdata/bass.wav 2
}

func Example_setTag() {
	src, err := os.Open("testdata/kick.wav")
	if err != nil {
		panic(err)
	}
	defer func() { _ = src.Close() }()

	rif := riff.New(riff.WithLoadData())
	if _, err = rif.ReadFrom(src); err != nil {
		panic(err)
	}

	title := riff.INFO(riff.LabINAM)
	if err = title.SetText([]byte("Kick")); err != nil {
		panic(err)
	}

	// Find the INFO list, or add one.
	chs := rif.Chunks()
	var info *riff.ChunkLIST
	for _, ch := range chs {
		if lst, ok := ch.(*riff.ChunkLIST); ok && lst.Type() == riff.IDINFO {
			info = lst
			break
		}
	}
	if info == nil {
		info = riff.LIST(nil)
		info.ListType = riff.IDINFO
		chs = append(chs, info)
	}
	info.Modify(append(info.Chunks().Remove(riff.LabINAM), title))
	rif.Modify(chs)

	dst := &bytes.Buffer{}
	if _, err = rif.WriteTo(dst); err != nil {
		panic(err)
	}

	// Decode the result again.
	if _, err = rif.ReadFrom(dst); err != nil {
		panic(err)
	}
	lst := rif.Chunks().First(riff.IDLIST).(*riff.ChunkLIST)
	tag := lst.Chunks().First(riff.LabINAM).(*riff.ChunkINFO)
	txt, err := io.ReadAll(tag.Text())
	if err != nil {
		panic(err)
	}
	fmt.Println(string(txt))
	// Output:
	// Kick
}

func ExampleRIFF_WriteTo() {
	src, err := os.Open("testdata/kick.wav")
	if err != nil {
		panic(err)
	}
	defer func() { _ = src.Close() }()

	// Load the chunk data so WriteTo can encode it back.
	rif := riff.New(riff.WithLoadData())
	if _, err = rif.ReadFrom(src); err != nil {
		panic(err)
	}

	// Edit a value.
	rif.Chunks().First(riff.IDfmt).(*riff.ChunkFMT).SampleRate = 44100

	dst := &bytes.Buffer{}
	if _, err = rif.WriteTo(dst); err != nil {
		panic(err)
	}

	// Decode the result again.
	if _, err = rif.ReadFrom(dst); err != nil {
		panic(err)
	}
	fmt.Println(rif.Chunks().First(riff.IDfmt).(*riff.ChunkFMT).SampleRate)
	// Output:
	// 44100
}
