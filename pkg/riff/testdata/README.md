# Test Files

## Generated

These files are written by [generate](generate/main.go), part of this project
and released under its [MIT License](../../../LICENSE.md). Each one models the
chunk layout of a file encoded with a Windows Audio Compression Manager codec,
with silent sample data. Regenerate them from the `pkg/riff` directory:

```sh
go run ./testdata/generate testdata
```

- 11k16bitpcm.wav
- 11k8bitpcm.wav
- 11kadpcm.wav
- 11kgsm.wav
- 11kulaw.wav
- 8k16bitpcm.wav
- 8k8bitpcm.wav
- 8kadpcm.wav
- 8kcelp.wav
- 8kgsm.wav
- 8kmp316.wav
- 8kmp38.wav
- 8ksbc12.wav
- 8ktruespeech.wav
- 8kulaw.wav
- sample16bit.wav

## go-audio

These files come from the go-audio projects and are distributed under the
[Apache License 2.0](licenses/Apache-2.0.txt).

### [go-audio/wav](https://github.com/go-audio/wav)

Copyright 2016 Matt Aimonetti.

- 32bit.wav
- 8bit.wav
- bass.wav
- bloop.aif
- bwf.wav
- dirty-kick-24b441k.wav
- flloop.wav - modified: the smpl `SamplerDataCnt` field at offset 433208 is
  changed from 0 to 24, the size of the sample loop list, to test a field
  that encoders fill in either way
- kick-16b441k.wav
- kick.wav
- listChunkInHeader.wav - modified: the pad byte at offset 85, after the
  `IENG` INFO sub-chunk, is changed from 1 to 0, as the RIFF specification
  requires
- listinfo.wav
- misaligned-chunk.wav
- padded24b.wav
- sample.avi

### [go-audio/riff](https://github.com/go-audio/riff)

Copyright 2019 Matt Aimonetti.

- junkKick.wav - modified: the fmt extra format bytes size at offset 746 is
  changed from 0 to 22, the number of extra bytes the chunk holds
- sample.rmi
- sample.wav
