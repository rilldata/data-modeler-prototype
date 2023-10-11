package printer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/gocarina/gocsv"
	"github.com/lensesio/tableprinter"
)

type Printer struct {
	humanOut    io.Writer
	resourceOut io.Writer
	format      *Format
}

type Format int

const (
	Human Format = iota
	JSON
	CSV
)

func NewFormatValue(val Format, p *Format) *Format {
	*p = val
	return p
}

func (f *Format) String() string {
	switch *f {
	case Human:
		return "human"
	case JSON:
		return "json"
	case CSV:
		return "csv"
	}

	return "unknown format"
}

func (f *Format) Set(s string) error {
	var v Format
	switch s {
	case "human":
		v = Human
	case "json":
		v = JSON
	case "csv":
		v = CSV
	default:
		return fmt.Errorf("failed to parse Format: %q. Valid values: %+v",
			s, []string{"human", "json", "csv"})
	}

	*f = v
	return nil
}

func (f *Format) Type() string {
	return "string"
}

func NewPrinter(format *Format) *Printer {
	return &Printer{
		format: format,
	}
}

// Format returns the format that was set for this printer
func (p *Printer) Format() Format { return *p.format }

// SetHumanOutput sets the output for human readable messages.
func (p *Printer) SetHumanOutput(out io.Writer) {
	p.humanOut = out
}

// SetResourceOutput sets the output for pringing resources via PrintResource.
func (p *Printer) SetResourceOutput(out io.Writer) {
	p.resourceOut = out
}

func (p *Printer) PrintResource(v interface{}) error {
	if p.format == nil {
		return errors.New("printer.Format is not set")
	}

	var out io.Writer = os.Stdout
	if p.resourceOut != nil {
		out = p.resourceOut
	}

	switch *p.format {
	case Human:
		var b strings.Builder
		tableprinter.Print(&b, v)
		fmt.Fprint(out, b.String())
		return nil
	case JSON:
		return p.PrintJSON(v)
	case CSV:
		type csvvaluer interface {
			MarshalCSVValue() interface{}
		}

		if c, ok := v.(csvvaluer); ok {
			v = c.MarshalCSVValue()
		}

		buf, err := gocsv.MarshalString(v)
		if err != nil {
			return fmt.Errorf("failed to marshal CSV: %w", err)
		}
		fmt.Fprintln(out, buf)
		return nil
	}

	return fmt.Errorf("unknown printer.Format: %T", *p.format)
}

func (p *Printer) PrintJSON(v interface{}) error {
	var out io.Writer = os.Stdout
	if p.resourceOut != nil {
		out = p.resourceOut
	}

	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintln(out, string(buf))
	return nil
}

func (p *Printer) Printf(format string, i ...interface{}) {
	fmt.Fprintf(p.out(), format, i...)
}

func (p *Printer) Println(i ...interface{}) {
	fmt.Fprintln(p.out(), i...)
}

func (p *Printer) Print(i ...interface{}) {
	fmt.Fprint(p.out(), i...)
}

// BoldGreen returns a string formatted with green and bold.
func BoldGreen(msg interface{}) string {
	return color.New(color.FgGreen).Add(color.Bold).Sprint(msg)
}

func (p *Printer) out() io.Writer {
	if p.humanOut != nil {
		return p.humanOut
	}

	if *p.format == Human {
		return color.Output
	}

	return io.Discard
}
