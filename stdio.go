package simplemcp

import (
	"io"
	"os"
)

func newStdioReadWriteCloser() *stdioReadWriteCloser {
	return &stdioReadWriteCloser{
		in:  os.Stdin,
		out: os.Stdout,
	}
}

type stdioReadWriteCloser struct {
	in  io.Reader
	out io.Writer
}

func (s *stdioReadWriteCloser) Read(p []byte) (int, error) {
	return s.in.Read(p)
}

func (s *stdioReadWriteCloser) Write(p []byte) (int, error) {
	return s.out.Write(p)
}

func (s *stdioReadWriteCloser) Close() error {
	return nil
}
