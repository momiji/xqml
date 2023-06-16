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
	attributes bool
	namespaces bool
	forceList  map[string]bool
	html       bool
	indent     string
	root       string
	element    string
	encoder    *xml.Encoder
	decoder    *xml.Decoder
	partial    bool
	done       bool
}

func NewXQML() *Xqml {
	return &Xqml{
		attributes: true,
		namespaces: true,
		forceList:  nil,
		html:       false,
		indent:     "",
		root:       DefaultRootTag,
		element:    DefaultElementTag,
		done:       false,
	}
}

// SetReadAttributes allows to keep attributes.
// Default is to keep attributes, use SetReadAttributes(false) to remove them.
func (x *Xqml) SetReadAttributes(b bool) {
	x.attributes = b
}

// SetReadNamespaces allows to keep namespaces.
// Default is to keep namespaces, use SetReadNamespaces(false) to remove them.
func (x *Xqml) SetReadNamespaces(b bool) {
	x.namespaces = b
}

// SetReadForceList allows to ensure some elements are parsed as slice, even when only one element is present.
// Use "x" for element name at any path, or "r.x" path. Also supports multiple commma separated paths at once, like "a.b,a.c,d".
func (x *Xqml) SetReadForceList(s ...string) {
	if x.forceList == nil {
		x.forceList = make(map[string]bool)
	}
	for _, a := range s {
		for _, b := range strings.Split(a, ",") {
			x.forceList[b] = true
		}
	}
}

// SetReadHtml allows to manage HTML content, by auto-closing known HTML tags.
// Default is to not manage HTML content.
func (x *Xqml) SetReadHtml(b bool) {
	x.html = b
}

// SetWriteRoot allows to set default element tag name in <root>...</root>.
// It used only with WriteXml(value).
func (x *Xqml) SetWriteRoot(s string) {
	x.root = s
}

// SetWriteElement allows to set default element tag name in <root><element>...</element></root>.
// It used only with WriteXml([]value).
func (x *Xqml) SetWriteElement(s string) {
	x.element = s
}

// SetWriteIndent allows to set indent string. Use "" for compact write.
func (x *Xqml) SetWriteIndent(s string) {
	x.indent = s
}

// SetWriteEncoder allows to set a customer xml encoder.
func (x *Xqml) SetWriteEncoder(e *xml.Encoder) {
	x.encoder = e
}

// SetReadDecoder allows to set a customer xml decoder.
// Default decoder has Strict = true and Entity = xml.HTMLEntity.
func (x *Xqml) SetReadDecoder(e *xml.Decoder) {
	x.decoder = e
}

// SetReadPartials allow to call ParseXml multiple times and return multiple XML files
func (x *Xqml) SetReadPartials(b bool) {
	x.partial = b
}

func (x *Xqml) ParseXml(reader io.Reader, cast bool) (map[string]any, error) {
	if x.decoder == nil {
		decoder := xml.NewDecoder(reader)
		decoder.Strict = false
		decoder.Entity = xml.HTMLEntity
		if x.html {
			decoder.AutoClose = xml.HTMLAutoClose
		}
		x.decoder = decoder
	}
	root := make(map[string]any)
	curr := elem{data: root, name: "", path: ""}
	err := x.parse(&curr, nil, cast)
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
	err := x.write(value)
	if err != nil {
		return err
	}
	return x.encoder.Close()
}

func Stringify(v any) string {
	s, _ := json.Marshal(v)
	return string(s)
}

func ToJson(b []byte) (map[string]any, error) {
	var v map[string]any
	err := json.Unmarshal(b, &v)
	if err != nil {
		return nil, err
	}
	return v, nil
}
