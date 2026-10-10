// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"os"
	"testing"

	"github.com/ctx42/memfs/pkg/memfs"
	"github.com/ctx42/riff/internal/test"
	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
	"github.com/ctx42/testkit/pkg/testkit"
)

func Test_RIFF_New(t *testing.T) {
	// --- When ---
	rif := New(WithLoadData())

	// --- Then ---
	assert.Equal(t, IDRIFF, rif.ID())
	assert.Equal(t, uint32(0), rif.Size())
	assert.Equal(t, uint32(0), rif.Type())
	assert.False(t, rif.Multi())

	assert.True(t, rif.IsRegistered(IDfmt))
	assert.True(t, rif.IsRegistered(IDdata))
	assert.False(t, rif.IsRegistered(0))
}

func Test_RIFF_Bare(t *testing.T) {
	// --- When ---
	reg := NewRegistry(RAWCMake(WithLoadData()))
	rif := Bare(reg)

	// --- Then ---
	assert.Equal(t, IDRIFF, rif.ID())
	assert.Equal(t, uint32(0), rif.Size())
	assert.Equal(t, uint32(0), rif.Type())
	assert.False(t, rif.Multi())

	assert.False(t, rif.IsRegistered(IDfmt))
	assert.False(t, rif.IsRegistered(IDdata))
}

func Test_RIFF_Bare_NilRegistry(t *testing.T) {
	// --- When ---
	rif := Bare(nil)

	// --- Then ---
	assert.Equal(t, IDRIFF, rif.ID())
	assert.Equal(t, uint32(0), rif.Size())
	assert.Equal(t, uint32(0), rif.Type())
	assert.False(t, rif.Multi())

	assert.False(t, rif.IsRegistered(IDfmt))
	assert.False(t, rif.IsRegistered(IDdata))
}

func Test_RIFF_ReadFrom_SmokeTest(t *testing.T) {
	rif := New()

	tt := []struct {
		pth    string
		typ    uint32
		size   int64
		chunks int
	}{
		{"testdata/11k16bitpcm.wav", TypeWAVE, 304578, 2},
		{"testdata/11k8bitpcm.wav", TypeWAVE, 152312, 2},
		{"testdata/11kadpcm.wav", TypeWAVE, 77252, 3},
		{"testdata/11kgsm.wav", TypeWAVE, 31000, 3},
		{"testdata/11kulaw.wav", TypeWAVE, 152326, 3},
		{"testdata/32bit.wav", TypeWAVE, 352844, 2},
		{"testdata/8bit.wav", TypeWAVE, 88244, 2},
		{"testdata/8k16bitpcm.wav", TypeWAVE, 221026, 2},
		{"testdata/8k8bitpcm.wav", TypeWAVE, 110532, 2},
		{"testdata/8kadpcm.wav", TypeWAVE, 56072, 3},
		{"testdata/8kcelp.wav", TypeWAVE, 8350, 3},
		{"testdata/8kgsm.wav", TypeWAVE, 22550, 3},
		{"testdata/8kmp316.wav", TypeWAVE, 27386, 3},
		{"testdata/8kmp38.wav", TypeWAVE, 13584, 3},
		{"testdata/8ksbc12.wav", TypeWAVE, 19658, 3},
		{"testdata/8ktruespeech.wav", TypeWAVE, 14842, 3},
		{"testdata/8kulaw.wav", TypeWAVE, 110546, 3},
		{"testdata/bass.wav", TypeWAVE, 143786, 2},
		{"testdata/bwf.wav", TypeWAVE, 27074, 14},
		{"testdata/dirty-kick-24b441k.wav", TypeWAVE, 132136, 3},
		{"testdata/flloop.wav", TypeWAVE, 434838, 7},
		{"testdata/junkKick.wav", TypeWAVE, 83084, 9},
		{"testdata/kick-16b441k.wav", TypeWAVE, 31692, 3},
		{"testdata/kick.wav", TypeWAVE, 9012, 2},
		{"testdata/listChunkInHeader.wav", TypeWAVE, 104196, 4},
		{"testdata/listinfo.wav", TypeWAVE, 176720, 4},
		{"testdata/misaligned-chunk.wav", TypeWAVE, 3441572, 8},
		{"testdata/padded24b.wav", TypeWAVE, 24908, 5},
		{"testdata/sample16bit.wav", TypeWAVE, 882160, 3},
		{"testdata/sample.avi", TypeAVI, 230264, 4}, // --
		{"testdata/sample.rmi", TypeRMID, 29640, 4}, // --
		{"testdata/sample.wav", TypeWAVE, 54002, 2},
	}

	for _, tc := range tt {
		t.Run(tc.pth, func(t *testing.T) {
			// --- Given ---
			fil := must.Value(os.Open(tc.pth))

			// --- When ---
			n, err := rif.ReadFrom(fil)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.size, n)
			assert.Equal(t, uint32(tc.size-8), rif.Size())
			assert.Equal(t, tc.typ, rif.Type())
			assert.Len(t, tc.chunks, rif.Chunks())
		})
	}
}

func Test_RIFF_CorrectingSize(t *testing.T) {
	// --- Given ---
	rif := New(WithLoadData())
	_, err := rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
	assert.NoError(t, err)

	// --- When ---
	dst := &bytes.Buffer{}
	n, err := rif.WriteTo(dst)

	// --- Then ---
	assert.NoError(t, err)

	assert.Equal(t, int64(27074), n)
	assert.Equal(t, uint32(27074), rif.Size()+8)
	assert.Equal(t, "ebdf4fefcbbb804b44cddbc50521ba29f833e05c", testkit.SHA1Reader(dst))
}

func Test_RIFF_WriteTo_SmokeTest(t *testing.T) {
	rif := New(WithLoadData())

	tt := []struct {
		pth  string
		size int64
		hash string
	}{
		{"testdata/11k16bitpcm.wav", 304578, "5f90c25e3993819fc8765e87cb262751548ce71c"},
		{"testdata/11k8bitpcm.wav", 152312, "96ae9bec8dfcd366918387e05b21b114fd3e54e4"},
		{"testdata/11kadpcm.wav", 77252, "3be2a0c1ffebae0636fc77f8ef8f701be8a7de9e"},
		{"testdata/11kgsm.wav", 31000, "019860d50c93ac790f0bcffaa23df418236da075"},
		{"testdata/11kulaw.wav", 152326, "8288b2123bfc102e97ed849c659b29df1e5519b7"},
		{"testdata/32bit.wav", 352844, "46bc96f0953df4f6f36053e65733f5e32d276167"},
		{"testdata/8bit.wav", 88244, "ef28c1a4fc75b0e4630499b1899bef244c837424"},
		{"testdata/8k16bitpcm.wav", 221026, "c853aa623e1a3b717fa452665f078b5b0dbf3139"},
		{"testdata/8k8bitpcm.wav", 110532, "118c7c2d55a6bbaf55543cd8e405be146f41666d"},
		{"testdata/8kadpcm.wav", 56072, "119f54e8cf5469d6eba0cb7a97a91ca2597faeec"},
		{"testdata/8kcelp.wav", 8350, "e534f24ce1177a9db1d538ddb5c523125854b7a5"},
		{"testdata/8kgsm.wav", 22550, "6b97716f68ec2d11d2fc5ef28642a2a78f39c376"},
		{"testdata/8kmp316.wav", 27386, "da10b9c9c4f71b7f661feebf023fdb22f492b437"},
		{"testdata/8kmp38.wav", 13584, "85da88b8fa13ea4bbee491389125729b40203682"},
		{"testdata/8ksbc12.wav", 19658, "35036ceba255622883b5b3c948e13eb1e7184a4e"},
		{"testdata/8ktruespeech.wav", 14842, "1be8b839cbf0135d6b330b5a5b3d63f65c1daae9"},
		{"testdata/8kulaw.wav", 110546, "5205e7e836c2f50da1261c32649532e4fee2cdaf"},
		{"testdata/bass.wav", 143786, "5c547df0f3d24da873d92e28ca3ecd22b83ad92b"},
		{"testdata/dirty-kick-24b441k.wav", 132136, "61b6cebfe073f34a81bc4352e69b5a044b07e188"},
		{"testdata/flloop.wav", 434838, "d12b046588af5475d846dd4a72f957cf03a33591"},
		{"testdata/junkKick.wav", 83084, "5974547313bcb8804618a3048f6902a429d62a4e"},
		{"testdata/kick-16b441k.wav", 31692, "1d7dbd0fe12ce2f8ec33ef5f90e271125a83ea94"},
		{"testdata/kick.wav", 9012, "cfa0812a881c3d3f2a2783b5b7a6d2ba62f7a1aa"},
		{"testdata/listChunkInHeader.wav", 104196, "9d722b63acf00c0031a5aaf7b3f73291321aafe3"},
		{"testdata/listinfo.wav", 176720, "8dd085ecf49c1ca781857dd1f26b16b636b7ab07"},
		{"testdata/misaligned-chunk.wav", 3441572, "dda435791002e772da033e3e7f09a3854c9e8d76"},
		{"testdata/padded24b.wav", 24908, "25b88e9afdc533be201b4393bc489cbfa61ec2e7"},
		{"testdata/sample.avi", 230264, "8f30db3104fafec017241b63fbba6588ee8cd5b4"},
		{"testdata/sample.rmi", 29640, "aa4cdab911e956439f8553772b56cd01993e9e12"},
	}

	for _, tc := range tt {
		t.Run(tc.pth, func(t *testing.T) {
			// --- Given ---
			_, err := rif.ReadFrom(must.Value(os.Open(tc.pth)))
			assert.NoError(t, err)

			// --- When ---
			buf := &bytes.Buffer{}
			n, err := rif.WriteTo(buf)

			// --- Then ---
			assert.NoError(t, err)

			// nolint: gocritic
			// fil := kit.CreateFile(t, filepath.Join("tmp", filepath.Base(tc.pth)))
			// content := kit.ReadAll(t, buf)
			// fil.Write(content)
			// buf = bytes.NewBuffer(content)

			assert.Equal(t, tc.size, n)
			size := int64(RealSize(rif.Size()) + 8)
			assert.Equal(t, tc.size, size)
			assert.Equal(t, tc.hash, testkit.SHA1Reader(buf))
		})
	}
}

func Test_Compose(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		rif := New()
		_, _ = rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
		chs := rif.Chunks()

		// --- When ---
		nw := Compose(chs.Remove(IDfmt))

		// --- Then ---
		assert.Equal(t, 14, len(chs.IDs()))
		assert.Equal(t, 13, len(nw.Chunks()))
	})

	t.Run("size ok", func(t *testing.T) {
		// --- Given ---
		rif := New()
		_, _ = rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
		chs := rif.Chunks()

		// --- When ---
		nw := Compose(chs)

		// --- Then ---
		assert.Equal(t, rif.Size(), nw.Size())
		assert.Equal(t, len(chs.IDs()), len(nw.Chunks().IDs()))
	})
}

func Test_Modify(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		// --- Given ---
		rif := New()
		_, _ = rif.ReadFrom(must.Value(os.Open("testdata/bwf.wav")))
		s := rif.Size()
		chs := rif.Chunks()
		f := chs.First(IDfmt)

		// --- When ---
		rif.Modify(chs.Remove(IDfmt))

		// --- Then ---
		assert.Equal(t, s-(f.Size()+8), rif.Size())
		assert.Equal(t, 13, len(rif.Chunks().IDs()))
	})
}

func Benchmark_RIFFReuse(b *testing.B) {
	tt := []struct {
		pth string
	}{
		{"testdata/11k16bitpcm.wav"},
		{"testdata/11k8bitpcm.wav"},
		{"testdata/11kadpcm.wav"},
		{"testdata/11kgsm.wav"},
		{"testdata/11kulaw.wav"},
		{"testdata/32bit.wav"},
		{"testdata/8bit.wav"},
		{"testdata/8k16bitpcm.wav"},
		{"testdata/8k8bitpcm.wav"},
		{"testdata/8kadpcm.wav"},
		{"testdata/8kcelp.wav"},
		{"testdata/8kgsm.wav"},
		{"testdata/8kmp316.wav"},
		{"testdata/8kmp38.wav"},
		{"testdata/8ksbc12.wav"},
		{"testdata/8ktruespeech.wav"},
		{"testdata/8kulaw.wav"},
		{"testdata/bass.wav"},
		{"testdata/bwf.wav"},
		{"testdata/dirty-kick-24b441k.wav"},
		{"testdata/flloop.wav"},
		{"testdata/junkKick.wav"},
		{"testdata/kick-16b441k.wav"},
		{"testdata/kick.wav"},
		{"testdata/listChunkInHeader.wav"},
		{"testdata/listinfo.wav"},
		{"testdata/misaligned-chunk.wav"},
		{"testdata/padded24b.wav"},
		{"testdata/sample16bit.wav"},
		{"testdata/sample.avi"},
		{"testdata/sample.rmi"},
		{"testdata/sample.wav"},
	}

	rif := New(WithLoadData())

	for _, tc := range tt {
		b.Run(tc.pth, func(b *testing.B) {
			b.StopTimer()
			buf := must.Value(memfs.NewFile("file"))
			_, err := buf.ReadFrom(must.Value(os.Open(tc.pth)))
			if err != nil {
				b.Fatal(err)
			}

			b.StartTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				buf.SeekStart()
				if _, err := rif.ReadFrom(buf); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func Test_RIFF_ExtraBytesAtTheEnd(t *testing.T) {
	t.Run("respect declared size", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.ReadFrom(t, src, Uint32(IDRIFF))
		test.WriteUint32LE(t, src, 4)
		test.ReadFrom(t, src, Uint32(TypeWAVE))
		test.WriteUint32LE(t, src, 0x64646464)
		test.WriteUint32LE(t, src, 4)
		test.WriteUint32LE(t, src, 1000)

		rif := New(WithLoadData())

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(12), n)
		assert.Equal(t, uint32(4), rif.Size())
		assert.Len(t, 0, rif.Chunks())
	})

	t.Run("ignore declared size - extra valid chunk", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.ReadFrom(t, src, Uint32(IDRIFF))
		test.WriteUint32LE(t, src, 4)
		test.ReadFrom(t, src, Uint32(TypeWAVE))
		test.WriteUint32LE(t, src, 0x64646464)
		test.WriteUint32LE(t, src, 4)
		test.WriteUint32LE(t, src, 1000)

		rif := New(WithLoadData(), WithIgnoreSize())

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(24), n)
		assert.Equal(t, uint32(16), rif.Size()) // Size corrected.
		assert.Len(t, 1, rif.Chunks())
	})

	t.Run("ignore declared size - noise at the end", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.ReadFrom(t, src, Uint32(IDRIFF))
		test.WriteUint32LE(t, src, 4)
		test.ReadFrom(t, src, Uint32(TypeWAVE))
		test.WriteUint32LE(t, src, 0x64646464)
		test.WriteUint32LE(t, src, 8000)
		test.WriteUint32LE(t, src, 7000)

		rif := New(WithLoadData(), WithIgnoreSize())

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		want := "error reading RIFF offset 24: error decoding RAWC:dddd chunk: unexpected EOF"
		assert.ErrorContain(t, want, err)
		assert.Equal(t, int64(24), n)
		assert.Equal(t, uint32(4), rif.Size())
		assert.Len(t, 0, rif.Chunks())
	})
}

func Test_RIFF_TooShortFile(t *testing.T) {
	t.Run("fewer bytes than declared", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.ReadFrom(t, src, Uint32(IDRIFF))
		test.WriteUint32LE(t, src, 20)
		test.ReadFrom(t, src, Uint32(TypeWAVE))
		test.WriteUint32LE(t, src, 0x64646464)
		test.WriteUint32LE(t, src, 4)
		test.WriteUint32LE(t, src, 1000)

		rif := New(WithLoadData())

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		want := "RIFF declared 20 bytes, decoder read 16 bytes: unexpected EOF"
		assert.ErrorContain(t, want, err)
		assert.Equal(t, int64(24), n)
		assert.Equal(t, uint32(20), rif.Size())
		assert.Len(t, 1, rif.Chunks())
	})

	t.Run("chunk name cut in half", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.ReadFrom(t, src, Uint32(IDRIFF))
		test.WriteUint32LE(t, src, 20)
		test.ReadFrom(t, src, Uint32(TypeWAVE))
		test.WriteUint16LE(t, src, 0x6464)

		rif := New(WithLoadData())

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		want := "error reading RIFF offset 14: unexpected EOF"
		assert.ErrorContain(t, want, err)
		assert.Equal(t, int64(14), n)
		assert.Equal(t, uint32(20), rif.Size())
		assert.Len(t, 0, rif.Chunks())
	})
}

func Test_RIFF_SizeLimit(t *testing.T) {
	t.Run("within limit", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.ReadFrom(t, src, Uint32(IDRIFF))
		test.WriteUint32LE(t, src, 16)
		test.ReadFrom(t, src, Uint32(TypeWAVE))
		test.WriteUint32LE(t, src, 0x64646464)
		test.WriteUint32LE(t, src, 4)
		test.WriteUint32LE(t, src, 1000)

		rif := New(WithLoadData(), WithSizeLimit(16))

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, int64(24), n)
		assert.Equal(t, uint32(16), rif.Size())
		assert.Len(t, 1, rif.Chunks())
	})

	t.Run("main chunk limit hit", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.ReadFrom(t, src, Uint32(IDRIFF))
		test.WriteUint32LE(t, src, 16)
		test.ReadFrom(t, src, Uint32(TypeWAVE))
		test.WriteUint32LE(t, src, 0x64646464)
		test.WriteUint32LE(t, src, 4)
		test.WriteUint32LE(t, src, 1000)

		rif := New(WithLoadData(), WithSizeLimit(15))

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		want := "error decoding RIFF chunk: size too large"
		assert.ErrorContain(t, want, err)
		assert.Equal(t, int64(4), n)
		assert.Equal(t, uint32(0), rif.Size())
		assert.Len(t, 0, rif.Chunks())
	})

	t.Run("subchunk limit hit", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.ReadFrom(t, src, Uint32(IDRIFF))
		test.WriteUint32LE(t, src, 16)
		test.ReadFrom(t, src, Uint32(TypeWAVE))
		test.WriteUint32LE(t, src, 0x64646464)
		test.WriteUint32LE(t, src, 4)
		test.WriteUint32LE(t, src, 1000)

		reg := NewRegistry(RAWCMake(WithLoadData(), WithSizeLimit(3)))
		rif := Bare(reg)

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		want := "error reading RIFF offset 20: error decoding RAWC:dddd chunk: size too large"
		assert.ErrorContain(t, want, err)
		assert.Equal(t, int64(20), n)
		assert.Equal(t, uint32(16), rif.Size())
		assert.Len(t, 0, rif.Chunks())
	})

	t.Run("ignore declared size - size limit hit", func(t *testing.T) {
		// --- Given ---
		src := bytes.NewBuffer(nil)
		test.ReadFrom(t, src, Uint32(IDRIFF))
		test.WriteUint32LE(t, src, 4)
		test.ReadFrom(t, src, Uint32(TypeWAVE))
		test.WriteUint32LE(t, src, 0x64646464)
		test.WriteUint32LE(t, src, 4)
		test.WriteUint32LE(t, src, 1000)

		rif := New(WithLoadData(), WithIgnoreSize(), WithSizeLimit(15))

		// --- When ---
		n, err := rif.ReadFrom(src)

		// --- Then ---
		want := "error reading RIFF offset 23: error decoding RAWC:dddd chunk: unexpected EOF"
		assert.ErrorContain(t, want, err)
		assert.Equal(t, int64(23), n)
		assert.Equal(t, uint32(4), rif.Size())
		assert.Len(t, 0, rif.Chunks())
	})
}
