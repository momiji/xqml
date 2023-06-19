package xqml

import (
	"encoding/xml"
	"io"
)

const (
	DefaultRootTag    = "root"
	DefaultElementTag = "element"
)

type Encoder struct {
	// Indent allows to set output indentation. Default is "".
	Indent string
	// Root allows to set root element name. Default is "root".
	Root string
	// Element allows to set root.element element name. Default is "element".
	Element     string
	encoder     *xml.Encoder
	initialized bool
}

// NewEncoder returns a new encoder that writes to w.
// The Encoder should be closed after use to flush all data
// to w.
func NewEncoder(writer io.Writer) *Encoder {
	encoder := xml.NewEncoder(writer)
	return &Encoder{
		Indent:  "",
		Root:    DefaultRootTag,
		Element: DefaultElementTag,
		encoder: encoder,
	}
}

// Encode writes the XML encoding of v to the stream.
//
// See the documentation for Marshal for details about the conversion of Go
// values to XML.
func (x *Encoder) Encode(value any) error {
	// initialize
	if !x.initialized {
		x.encoder.Indent("", x.Indent)
		x.initialized = true
	}
	// write output
	err := x.write(value)
	if err != nil {
		return err
	}
	// return
	return x.encoder.Close()
}
