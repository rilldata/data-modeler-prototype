package printer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
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

func (p *Printer) PrintlnSuccess(str string) {
	p.Println(BoldGreen(str))
}

func (p *Printer) PrintlnWarn(str string) {
	p.Println(BoldYellow(str))
}

func (p *Printer) PrintlnError(str string) {
	p.Println(BoldRed(str))
}

func (p *Printer) PrintlnInfo(str string) {
	p.Println(BoldWhite(str))
}

func (p *Printer) PrintBold(str string) {
	p.Print(Bold(str))
}

// BoldGreen returns a string formatted with green and bold.
func BoldGreen(msg interface{}) string {
	return color.New(color.FgGreen).Add(color.Bold).Sprint(msg)
}

// BoldYellow returns a string formatted with yellow and bold.
func BoldYellow(msg interface{}) string {
	return color.New(color.FgYellow).Add(color.Bold).Sprint(msg)
}

// BoldRed returns a string formatted with red and bold.
func BoldRed(msg interface{}) string {
	return color.New(color.FgRed).Add(color.Bold).Sprint(msg)
}

// BoldWhite returns a string formatted with white and bold.
func BoldWhite(msg interface{}) string {
	return color.New(color.FgWhite).Add(color.Bold).Sprint(msg)
}

// Bold returns a string formatted with bold.
func Bold(msg interface{}) string {
	return color.New(color.Bold).Sprint(msg)
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
