package xqml

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"
)

const (
	DefaultRootTag    = "root"
	DefaultElementTag = "element"
)

type Xqml struct {
	namespace bool
	forceList map[string]bool
	indent    string
	root      string
	element   string
	encoder   *xml.Encoder
	decoder   *xml.Decoder
	partial   bool
	done      bool
}

func NewXQML() *Xqml {
	return &Xqml{
		namespace: true,
		forceList: nil,
		indent:    "",
		root:      DefaultRootTag,
		element:   DefaultElementTag,
		done:      false,
	}
}

// Namespace allows to keep namespaces.
// Default is to keep namespaces, use Namespace(false) to remove them.
func (x *Xqml) Namespace(b bool) {
	x.namespace = b
}

// ForceList allows to ensure some elements are parsed as slice, even when only one element is present.
func (x *Xqml) ForceList(s ...string) {
	if x.forceList == nil {
		x.forceList = make(map[string]bool)
	}
	for _, a := range s {
		for _, b := range strings.Split(a, ",") {
			x.forceList[b] = true
		}
	}
}

// Root allows to set default element tag name in <root>...</root>.
// It used only with WriteXml(value).
func (x *Xqml) Root(s string) {
	x.root = s
}

// Element allows to set default element tag name in <root><element>...</element></root>.
// It used only with WriteXml([]value).
func (x *Xqml) Element(s string) {
	x.element = s
}

// Indent allows to set indent string. Use "" for compact write.
func (x *Xqml) Indent(s string) {
	x.indent = s
}

// Encoder allows to set a customer xml encoder.
func (x *Xqml) Encoder(e *xml.Encoder) {
	x.encoder = e
}

// Decoder allows to set a customer xml decoder.
// Default decoder has Strict = true and Entity = xml.HTMLEntity.
func (x *Xqml) Decoder(e *xml.Decoder) {
	x.decoder = e
}

// Partial allow to call ParseXml multiple times and return multiple XML files
func (x *Xqml) Partial(b bool) {
	x.partial = b
}

func (x *Xqml) ParseXml(reader io.Reader, cast bool) (map[string]any, error) {
	if x.decoder == nil {
		decoder := xml.NewDecoder(reader)
		decoder.Strict = false
		//decoder.AutoClose = xml.HTMLAutoClose
		decoder.Entity = xml.HTMLEntity
		x.decoder = decoder
	}
	root := make(map[string]any)
	curr := elem{data: root, name: "", path: ""}
	parent := curr
	err := x.parse(&curr, &parent, cast)
	if err != nil {
		return nil, err
	}
	if x.partial && len(root) == 0 {
		return nil, io.EOF
	}
	return root, nil
}

func (x *Xqml) WriteXml(writer io.Writer, value any) error {
	if x.encoder == nil {
		encoder := xml.NewEncoder(writer)
		if x.indent != "" {
			encoder.Indent("", x.indent)
		}
		x.encoder = encoder
	}
	err := x.write(value, "")
	if err != nil {
		return err
	}
	return x.encoder.Close()
}

func Stringify(v any) string {
	s, _ := json.Marshal(v)
	return string(s)
}
