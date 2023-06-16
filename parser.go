package xqml

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type elem struct {
	data    map[string]any
	name    string
	path    string
	isValue bool
	isEmpty bool
}

func (x *Xqml) parse(curr *elem, parent *elem, cast bool) error {
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
				return fmt.Errorf("Invalid XML element '%s' found for non-partial parse", e.Name.Local)
			}
			// create new element
			var data map[string]any
			name := newName(x.namespaces, &e.Name)
			path := newPath(curr.path, name)
			item := &elem{data, name, path, false, true}
			// read attributes
			if x.attributes {
				if len(e.Attr) > 0 {
					data = make(map[string]any)
					item.data = data
					item.isEmpty = false
				}
				for _, attr := range e.Attr {
					data["@"+newName(x.namespaces, &attr.Name)] = attr.Value
				}
			}
			// upgrade parent if it is empty or a value
			x.upgradeValue(curr, parent)
			// set value
			x.addValue(curr, name, path, data)
			// recursive call
			err = x.parse(item, curr, cast)
			if err != nil {
				return err
			}
			if curr.path == "" {
				if x.partial {
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
					return fmt.Errorf("Invalid XML chardata '%s' found for non-partial parse", cdata)
				}
				value := any(cdata)
				if cast {
					value = castValue(cdata)
				}
				x.setText(curr, parent, value)
			}
		}
	}
}

func (x *Xqml) setValue(item *elem, name string, path string, value any) {
	// if value already set => transform to slice or append to slice
	if data, ok := item.data[name]; ok {
		switch data.(type) {
		case []any:
			slice := data.([]any)
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

func (x *Xqml) addValue(item *elem, name string, path string, value any) {
	// if value already set => transform to slice or append to slice
	if data, ok := item.data[name]; ok {
		switch data.(type) {
		case []any:
			slice := data.([]any)
			slice = append(slice, value)
			item.data[name] = slice
			return
		default:
			item.data[name] = []any{data, value}
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

func (x *Xqml) setText(curr *elem, parent *elem, value any) {
	// if curr is empty, set value on parent
	if curr.isEmpty {
		x.setValue(parent, curr.name, curr.path, value)
		curr.isValue = true
		return
	}
	// set value
	if text, ok := curr.data["#text"]; ok {
		sep := ""
		if x.html {
			sep = " "
		}
		value = fmt.Sprintf("%v%s%v", text, sep, value)
	}
	curr.data["#text"] = value
}

func (x *Xqml) upgradeValue(curr *elem, parent *elem) {
	if curr.isEmpty {
		curr.data = make(map[string]any)
		curr.isEmpty = false
		if !curr.isValue {
			x.setValue(parent, curr.name, curr.path, curr.data)
		}
	}
	if curr.isValue {
		switch parent.data[curr.name].(type) {
		case []any:
			values := parent.data[curr.name].([]any)
			items := make([]any, len(values))
			var last map[string]any
			for i, v := range values {
				last = map[string]any{"#text": v}
				items[i] = last
			}
			curr.data = last
			curr.isValue = false
			parent.data[curr.name] = items
		default:
			curr.data["#text"] = parent.data[curr.name]
			curr.isValue = false
			parent.data[curr.name] = curr.data
		}
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
		case "true", "True", "false", "False":
			if b, err := strconv.ParseBool(s); err == nil {
				return b
			}
		}
	}
	return s
}
