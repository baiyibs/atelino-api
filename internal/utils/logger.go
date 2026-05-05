package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type ginWriter struct {
	out io.Writer
}

func (w *ginWriter) Write(p []byte) (n int, err error) {
	timestamp := time.Now().Format("2006/01/02 - 15:04:05")
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}

	const (
		colorReset = "\033[0m"
		colorCyan  = "\033[36m"
	)
	output := fmt.Sprintf("%s[Backend] %s | %s%s\n", colorCyan, timestamp, msg, colorReset)

	_, err = w.out.Write([]byte(output))
	return len(p), err
}

func SetupLog() {
	log.SetFlags(0)
	log.SetOutput(&ginWriter{out: os.Stdout})
}
