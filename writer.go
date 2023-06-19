package xqml

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

type tag struct {
	name  string
	value any
}

var emptyAttrs []xml.Attr

func (x *Encoder) write(value any) error {
	switch value.(type) {
	case map[string]any:
		// count number of elements
		m := value.(map[string]any)
		c := 0
		d := false
		key := ""
		for k := range m {
			if !strings.HasPrefix(k, "@") && k != "#text" {
				key = k
				c = c + 1
			} else {
				d = true
			}
		}
		if c == 0 {
			return x.writeAny(map[string]any{x.Root: value}, "")
		} else if c == 1 && d {
			return x.writeAny(map[string]any{x.Root: value}, "")
		} else if c == 1 {
			value2 := m[key]
			switch value2.(type) {
			case []any:
				return x.writeAny(map[string]any{x.Root: value}, "")
			default:
				return x.writeAny(value, "")
			}
		} else {
			return x.writeAny(map[string]any{x.Root: value}, "")
		}
	case []any:
		return x.writeAny(map[string]any{x.Root: map[string]any{x.Element: value}}, "")
	default:
		return x.writeAny(map[string]any{x.Root: value}, "")
	}

}

func (x *Encoder) writeAny(value any, parent string) error {
	switch value.(type) {
	case map[string]any:
		return x.writeMap(value.(map[string]any), parent)
	case []any:
		v := value.([]any)
		return x.writeSlice(&v, parent)
	default:
		return x.writeValue(value, parent)
	}
}

func (x *Encoder) writeMap(value map[string]any, parent string) error {
	var err error
	var attrs []*tag
	var elems []*tag
	var text any
	//var text any
	for k, v := range value {
		if strings.HasPrefix(k, "@") {
			attrs = append(attrs, &tag{k[1:], v})
		} else if k == "#text" {
			text = v
		} else {
			elems = append(elems, &tag{k, v})
		}
	}
	// remove root unexpected values
	if parent == "" {
		attrs = nil
	}
	// sorts tags and elems
	sort.Slice(attrs, func(i, j int) bool {
		return strings.Compare(attrs[i].name, attrs[j].name) < 0
	})
	sort.Slice(elems, func(i, j int) bool {
		return strings.Compare(elems[i].name, elems[j].name) < 0
	})
	// start
	name := xml.Name{Local: parent}
	if parent != "" {
		start := xml.StartElement{
			Name: name,
			Attr: *newAttrs(attrs),
		}
		err = x.encoder.EncodeToken(start)
		if err != nil {
			return err
		}
		if text != nil {
			cdata := xml.CharData(fmt.Sprintf("%v", text))
			err = x.encoder.EncodeToken(cdata)
			if err != nil {
				return err
			}
		}
	}
	// content
	for _, e := range elems {
		err = x.writeAny(e.value, e.name)
		if err != nil {
			return err
		}
	}
	// end
	if parent != "" {
		end := xml.EndElement{Name: name}
		err = x.encoder.EncodeToken(end)
		if err != nil {
			return err
		}
	}
	return nil
}
func (x *Encoder) writeSlice(value *[]any, parent string) error {
	for _, a := range *value {
		err := x.writeAny(a, parent)
		if err != nil {
			return err
		}
	}
	return nil
}

func (x *Encoder) writeValue(value any, parent string) error {
	name := xml.Name{Local: parent}
	start := xml.StartElement{
		Name: name,
		Attr: nil,
	}
	err := x.encoder.EncodeToken(start)
	if err != nil {
		return err
	}
	err = x.writeText(value)
	if err != nil {
		return err
	}
	end := xml.EndElement{
		Name: name,
	}
	err = x.encoder.EncodeToken(end)
	if err != nil {
		return err
	}
	return nil
}

func (x *Encoder) writeText(value any) error {
	if value == nil {
		return nil
	}
	cdata := xml.CharData(fmt.Sprintf("%v", value))
	return x.encoder.EncodeToken(cdata)
}

func newAttrs(attrs []*tag) *[]xml.Attr {
	if attrs == nil {
		return &emptyAttrs
	}
	res := make([]xml.Attr, len(attrs))
	for i, attr := range attrs {
		res[i] = xml.Attr{
			Name:  xml.Name{Local: attr.name},
			Value: attr.value.(string),
		}
	}
	return &res
}
