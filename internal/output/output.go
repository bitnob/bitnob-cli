package output

import (
	"encoding/json"
	"fmt"
	"io"
)

type Printer struct {
	Stdout io.Writer
	Stderr io.Writer
}

func New(stdout io.Writer, stderr io.Writer) Printer {
	return Printer{
		Stdout: stdout,
		Stderr: stderr,
	}
}

func (p Printer) Println(message string) error {
	_, err := fmt.Fprintln(p.Stdout, message)
	return err
}

func (p Printer) PrintJSON(v any) error {
	encoder := json.NewEncoder(p.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func (p Printer) Errorln(message string) error {
	_, err := fmt.Fprintln(p.Stderr, message)
	return err
}
