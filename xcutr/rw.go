package xcutr

import (
	"fmt"
	"io"
	"strings"

	fc "github.com/fatih/color"
)

type nodeWriter struct {
	name      string
	inner     io.Writer
	colorFunc func(a ...interface{}) string
}

func (cw *nodeWriter) Write(data []byte) (int, error) {
	strData := string(data)
	lines := strings.Split(strData, "\n")
	for _, ln := range lines {
		if ln == "" || strings.Contains(ln, "[sudo] password for") {
			continue
		}
		_, err := fmt.Fprintf(cw.inner, "%s | %2s\n", cw.colorFunc(cw.name), ln)
		if err != nil {
			return 0, err
		}
	}
	return len(data), nil
}

func NewNodeWriter(
	name string, target io.Writer, color fc.Attribute) io.Writer {
	if len(name) < 10 {
		name = fmt.Sprintf("%10s", name)
	} else {
		name = fmt.Sprintf("%8s..", name[:8])
	}
	return &nodeWriter{
		name:      name,
		inner:     target,
		colorFunc: fc.New(color).SprintFunc(),
	}
}

// type sudoReader struct {
// 	password string
// 	done     bool
// }
