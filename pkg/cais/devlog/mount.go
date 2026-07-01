package devlog

import "io"

var defaultBuf *Buffer

// SetDefault makes the buffer available to MirrorDefault.
func SetDefault(buf *Buffer) {
	defaultBuf = buf
}

// Default returns the active development log buffer.
func Default() *Buffer {
	return defaultBuf
}

// Prepare initializes the default buffer for development environments.
func Prepare(env string) *Buffer {
	if !Enabled(env) {
		SetDefault(nil)
		return nil
	}
	buf := NewBuffer(1000)
	SetDefault(buf)
	return buf
}

// MirrorDefault duplicates writes to the default buffer when it becomes available.
func MirrorDefault(dst io.Writer) io.Writer {
	return &dynamicMirror{dst: dst}
}
