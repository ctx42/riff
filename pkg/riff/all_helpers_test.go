// SPDX-FileCopyrightText: (c) 2020 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package riff

import (
	"bytes"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/tester"
	"github.com/ctx42/testkit/pkg/iokit"
)

func Test_writeUint32LE(t *testing.T) {
	t.Run("write", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 1).Close()
		dst := &bytes.Buffer{}

		// --- When ---
		writeUint32LE(tspy, dst, 0x01020304)

		// --- Then ---
		assert.Equal(t, []byte{4, 3, 2, 1}, dst.Bytes())
	})

	t.Run("error - write", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t).ExpectFatal().IgnoreLogs().Close()

		// --- When ---
		assert.Panic(t, func() {
			writeUint32LE(tspy, iokit.ErrWriter(&bytes.Buffer{}, 0), 1)
		})
	})
}

func Test_writeUint32BE(t *testing.T) {
	t.Run("write", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 1).Close()
		dst := &bytes.Buffer{}

		// --- When ---
		writeUint32BE(tspy, dst, 0x01020304)

		// --- Then ---
		assert.Equal(t, []byte{1, 2, 3, 4}, dst.Bytes())
	})

	t.Run("error - write", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t).ExpectFatal().IgnoreLogs().Close()

		// --- When ---
		assert.Panic(t, func() {
			writeUint32BE(tspy, iokit.ErrWriter(&bytes.Buffer{}, 0), 1)
		})
	})
}

func Test_writeUint16LE(t *testing.T) {
	t.Run("write", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 1).Close()
		dst := &bytes.Buffer{}

		// --- When ---
		writeUint16LE(tspy, dst, 0x0102)

		// --- Then ---
		assert.Equal(t, []byte{2, 1}, dst.Bytes())
	})

	t.Run("error - write", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t).ExpectFatal().IgnoreLogs().Close()

		// --- When ---
		assert.Panic(t, func() {
			writeUint16LE(tspy, iokit.ErrWriter(&bytes.Buffer{}, 0), 1)
		})
	})
}

func Test_writeBytes(t *testing.T) {
	t.Run("write", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 1).Close()
		dst := &bytes.Buffer{}

		// --- When ---
		writeBytes(tspy, dst, []byte{1, 2})

		// --- Then ---
		assert.Equal(t, []byte{1, 2}, dst.Bytes())
	})

	t.Run("error - write", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t).ExpectFatal().IgnoreLogs().Close()

		// --- When ---
		assert.Panic(t, func() {
			writeBytes(tspy, iokit.ErrWriter(&bytes.Buffer{}, 0), []byte{1})
		})
	})
}

func Test_skip4B(t *testing.T) {
	t.Run("skip", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t, 1).Close()
		src := bytes.NewReader([]byte{1, 2, 3, 4, 5})

		// --- When ---
		skip4B(tspy, src)

		// --- Then ---
		assert.Equal(t, 1, src.Len())
	})

	t.Run("error - fewer than 4 bytes", func(t *testing.T) {
		// --- Given ---
		tspy := tester.New(t).ExpectFatal()
		tspy.ExpectLogEqual("unexpected EOF").Close()

		// --- When ---
		assert.Panic(t, func() {
			skip4B(tspy, bytes.NewReader([]byte{1, 2}))
		})
	})
}

func Test_isAllRead(t *testing.T) {
	t.Run("all read", func(t *testing.T) {
		// --- When ---
		have := isAllRead(bytes.NewReader(nil))

		// --- Then ---
		assert.True(t, have)
	})

	t.Run("bytes left", func(t *testing.T) {
		// --- When ---
		have := isAllRead(bytes.NewReader([]byte{1}))

		// --- Then ---
		assert.False(t, have)
	})
}
