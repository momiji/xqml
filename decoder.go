package xqml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

type Decoder struct {
	// Attributes allows to keep attributes. Default is true.
	Attributes bool
	// Namespaces allows to keep namespaces. Default is true.
	Namespaces bool
	// ForceList allows to force some elements to be parsed as slices, even when only one element is present.
	// Supports "r.x" paths notation and "x" element names. Multiple values can be passed as comma separated values, like "r.x,r.y,z".
	ForceList []string
	// Html allows HTML content, by auto-closing known HTML tags. Default is false.
	Html bool
	// Partials allow to call Decode() multiple times to return multiple XML files. When true, Decode() can be called until io.EOF is reached. Default is false.
	Partials bool
	// Cast allows to cast values to boolean/int/float. Default is true.
	Cast bool
	// Sep allows to set text separator between multiple CDATA. Default is " ".
	Sep         string
	decoder     *xml.Decoder
	forceList   map[string]bool
	done        bool
	initialized bool
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may read
// data from r beyond the XML values requested.
func NewDecoder(reader io.Reader) *Decoder {
	decoder := xml.NewDecoder(reader)
	decoder.Strict = false
	decoder.Entity = xml.HTMLEntity
	return &Decoder{
		Attributes:  true,
		Namespaces:  true,
		ForceList:   nil,
		Html:        false,
		Cast:        true,
		Sep:         " ",
		Partials:    false,
		decoder:     decoder,
		forceList:   nil,
		done:        false,
		initialized: false,
	}
}

// SetReadForceList allows to ensure some elements are parsed as slice, even when only one element is present.
// Use "x" for element name at any path, or "r.x" path. Also supports multiple commma separated paths at once, like "a.b,a.c,d".
func (x *Decoder) setForceList() {
	x.forceList = make(map[string]bool)
	for _, a := range x.ForceList {
		for _, b := range strings.Split(a, ",") {
			x.forceList[b] = true
		}
	}
}

// Decode reads the next XML-encoded value from its input
// and stores it in the value pointed to by v.
//
// See the documentation for Unmarshal for details about the
// conversion of XML into a Go value.
func (x *Decoder) Decode(v any) error {
	// check input value is a valid pointer
	switch v.(type) {
	case *any:
	case *map[string]any:
	default:
		return fmt.Errorf("invalid argument, must be a *map[string]any or *any")
	}
	// initialize
	if !x.initialized {
		if x.Html {
			x.decoder.AutoClose = xml.HTMLAutoClose
		}
		x.setForceList()
		x.initialized = true
	}
	// parse input
	root := map[string]any{}
	curr := elem{data: root, content: ContentObject}
	err := x.parse(&curr, nil)
	if err != nil {
		return err
	}
	// set return value
	switch v.(type) {
	case *any:
		*(v.(*any)) = root
	case *map[string]any:
		*(v.(*map[string]any)) = root
	}
	// return
	if x.Partials && len(root) == 0 {
		return io.EOF
	}
	return nil
}
