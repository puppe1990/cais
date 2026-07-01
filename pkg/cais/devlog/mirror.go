package devlog

import "io"

type mirror struct {
	dst io.Writer
	buf *Buffer
}

type dynamicMirror struct {
	dst io.Writer
}

func Mirror(dst io.Writer, buf *Buffer) io.Writer {
	return &mirror{dst: dst, buf: buf}
}

func (m *mirror) Write(p []byte) (int, error) {
	if m.buf != nil {
		_, _ = m.buf.Write(p)
	}
	if m.dst == nil {
		return len(p), nil
	}
	return m.dst.Write(p)
}

func (m *dynamicMirror) Write(p []byte) (int, error) {
	if defaultBuf != nil {
		_, _ = defaultBuf.Write(p)
	}
	if m.dst == nil {
		return len(p), nil
	}
	return m.dst.Write(p)
}
