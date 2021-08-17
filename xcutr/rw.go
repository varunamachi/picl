package xcutr

import (
	"fmt"
	"io"
	"strings"
)

type nodeWriter struct {
	name  string
	inner io.Writer
}

func (cw *nodeWriter) Write(data []byte) (int, error) {
	// cw.inner.Write([]byte("[" + cw.name + "]  "))

	name := cw.name
	if len(name) > 10 {
		name = name[:10]
	}

	strData := string(data)
	lines := strings.Split(strData, "\n")
	for _, ln := range lines {
		_, err := fmt.Fprintf(cw.inner, "%12s | %s\n", name, ln)
		if err != nil {
			return 0, err
		}
	}
	return len(data), nil
}

func NewNodeWriter(name string, target io.Writer) io.Writer {
	return &nodeWriter{
		name:  name,
		inner: target,
	}
}
