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

func (x *Xqml) write(value any, parent string) error {
	switch value.(type) {
	case map[string]any:
		return x.writeMap(value.(map[string]any), parent)
	case []any:
		if parent == "" {
			return x.write(map[string]any{x.root: map[string]any{x.element: value}}, "")
		} else {
			v := value.([]any)
			return x.writeSlice(&v, parent)
		}
	default:
		if parent == "" {
			return x.write(map[string]any{x.root: value}, "")
		} else {
			return x.writeValue(value, parent)
		}
	}
}

func (x *Xqml) writeMap(value map[string]any, parent string) error {
	var err error
	var attrs []*tag
	var elems []*tag
	var text any
	//var text any
	for k, v := range value {
		sk := string(k)
		if strings.HasPrefix(sk, "@") {
			attrs = append(attrs, &tag{sk[1:], v})
		} else if sk == "#text" {
			text = v
		} else {
			elems = append(elems, &tag{sk, v})
		}
	}
	// remove root unexpected values
	if parent == "" {
		attrs = nil
	}
	// sorts tags and elems
	sort.Slice(attrs, func(i, j int) bool {
		return strings.Compare(attrs[i].name, attrs[j].name) > 0
	})
	sort.Slice(elems, func(i, j int) bool {
		return strings.Compare(elems[i].name, elems[j].name) > 0
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
			cdata := xml.CharData([]byte(fmt.Sprintf("%v", text)))
			err = x.encoder.EncodeToken(cdata)
			if err != nil {
				return err
			}
		}
	}
	// content
	for _, e := range elems {
		err = x.write(e.value, e.name)
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
func (x *Xqml) writeSlice(value *[]any, parent string) error {
	for _, a := range *value {
		err := x.write(a, parent)
		if err != nil {
			return err
		}
	}
	return nil
}

func (x *Xqml) writeValue(value any, parent string) error {
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

func (x *Xqml) writeText(value any) error {
	if value == nil {
		return nil
	}
	cdata := xml.CharData([]byte(fmt.Sprintf("%v", value)))
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
