package xqml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const (
	ContentNone = iota
	ContentValue
	ContentObject
)

type elem struct {
	data    map[string]any
	name    string
	path    string
	content int
}

func (x *Decoder) parse(curr *elem, parent *elem) error {
	for {
		token, err := x.decoder.Token()
		// on error, check EOF
		if err != nil {
			if err == io.EOF {
				x.done = true
				return nil
			}
			return err
		}
		//
		switch token.(type) {
		case xml.StartElement:
			e := token.(xml.StartElement)
			if x.done {
				return fmt.Errorf("invalid XML element '%s' found for non-partial parse", e.Name.Local)
			}
			// create new element
			var data map[string]any
			name := newName(x.Namespaces, &e.Name)
			path := newPath(curr.path, name)
			item := &elem{data, name, path, ContentNone}
			// read attributes
			if x.Attributes {
				if len(e.Attr) > 0 {
					data = make(map[string]any)
					item.data = data
					item.content = ContentObject
				}
				for _, attr := range e.Attr {
					data["@"+newName(x.Namespaces, &attr.Name)] = attr.Value
				}
			}
			// upgrade parent if it is empty or a value
			x.upgradeValue(curr, parent)
			// set value
			x.addValue(curr, name, path, data)
			// recursive call
			err = x.parse(item, curr)
			if err != nil {
				return err
			}
			if curr.path == "" {
				if x.Partials {
					return nil
				} else {
					x.done = true
				}
			}
		case xml.EndElement:
			return nil
		case xml.CharData:
			cdata := string(token.(xml.CharData))
			cdata = strings.Trim(cdata, " \n\r\t")
			if cdata != "" {
				if x.done {
					return fmt.Errorf("invalid XML chardata '%s' found for non-partial parse", cdata)
				}
				value := any(cdata)
				if x.Cast {
					value = castValue(cdata)
				}
				x.setText(curr, parent, value)
			}
		}
	}
}

func (x *Decoder) getValue(item *elem, name string) any {
	// if value is already set...
	if data, isMap := item.data[name]; isMap {
		// if value is a slice, return last item
		if slice, isSlice := data.([]any); isSlice {
			return slice[len(slice)-1]
		}
		// return value
		return data
	}
	return nil
}

func (x *Decoder) setValue(item *elem, name string, path string, value any) {
	// if value is already set...
	if data, isMap := item.data[name]; isMap {
		// if value is a slice, set last item
		if slice, isSlice := data.([]any); isSlice {
			slice[len(slice)-1] = value
			return
		}
	}
	// set value or slice if forced
	if x.forceList[name] == true || x.forceList[path] == true {
		item.data[name] = []any{value}
	} else {
		item.data[name] = value
	}
}

func (x *Decoder) addValue(item *elem, name string, path string, value any) {
	// if value is already set => transform to slice or append to slice
	if data, isMap := item.data[name]; isMap {
		// if value is a slice
		if slice, isSlice := data.([]any); isSlice {
			slice = append(slice, value)
			item.data[name] = slice
			return
		}
		// transform to a slice
		item.data[name] = []any{data, value}
		return
	}
	// set value or slice if forced
	if x.forceList[name] == true || x.forceList[path] == true {
		item.data[name] = []any{value}
	} else {
		item.data[name] = value
	}
}

func (x *Decoder) setText(curr *elem, parent *elem, value any) {
	switch curr.content {
	case ContentNone:
		x.setValue(parent, curr.name, curr.path, value)
		curr.content = ContentValue
	case ContentValue:
		text := x.getValue(parent, curr.name)
		value = fmt.Sprintf("%v%s%v", text, x.Sep, value)
		x.setValue(parent, curr.name, curr.path, value)
	case ContentObject:
		if text, ok := curr.data["#text"]; ok {
			value = fmt.Sprintf("%v%s%v", text, x.Sep, value)
		}
		curr.data["#text"] = value
	}
}

func (x *Decoder) upgradeValue(curr *elem, parent *elem) {
	switch curr.content {
	case ContentNone:
		curr.data = make(map[string]any)
		curr.content = ContentObject
		x.setValue(parent, curr.name, curr.path, curr.data)
	case ContentValue:
		text := x.getValue(parent, curr.name)
		curr.data = map[string]any{"#text": text}
		curr.content = ContentObject
		x.setValue(parent, curr.name, curr.path, curr.data)
	}
}

func newName(keepNs bool, name *xml.Name) string {
	if keepNs && name.Space != "" {
		return name.Space + ":" + name.Local
	}
	return name.Local
}

func newPath(path string, name string) string {
	if path == "" {
		return name
	}
	return path + "." + name
}

func castValue(s string) any {
	if f, err := strconv.ParseInt(s, 10, 64); err == nil {
		return f
	}
	if f, err := strconv.ParseUint(s, 10, 64); err == nil {
		return f
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	if len(s) == 4 || len(s) == 5 {
		switch s {
		case "true", "True", "TRUE":
			return true
		case "false", "False", "FALSE":
			return false
		}
	}
	return s
}
