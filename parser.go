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
			data := make(map[string]any)
			name := newName(x.namespace, &e.Name)
			path := newPath(curr.path, name)
			item := &elem{data, name, path, false}
			// add attributes
			for _, attr := range e.Attr {
				data["@"+newName(x.namespace, &attr.Name)] = attr.Value
			}
			// upgrade parent if it is a value
			if curr.isValue {
				curr.data["#text"] = parent.data[curr.name]
				curr.isValue = false
				parent.data[curr.name] = curr.data
			}
			// set value
			x.setValue(curr, name, data)
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
			// if current is empty, downgrade parent
			if !curr.isValue && len(curr.data) == 0 {
				x.downgrade(curr, parent, nil)
			}
			return nil
		case xml.CharData:
			cdata := string(token.(xml.CharData))
			cdata = strings.Trim(cdata, " \n\r\t")
			if cdata != "" {
				if x.done {
					return fmt.Errorf("Invalid XML chardata '%s' found for non-partial parse", cdata)
				}
				if cast {
					x.setText(curr, parent, castValue(cdata))
				} else {
					x.setText(curr, parent, cdata)
				}
			}
		}
	}
}

func (x *Xqml) setValue(item *elem, name string, value any) {
	// if value already set => transform to slice or append to slice
	if data, ok := item.data[name]; ok {
		switch data.(type) {
		case []any:
			slice := data.([]any)
			item.data[name] = append(slice, value)
			return
		default:
			item.data[name] = []any{data, value}
			return
		}
	}
	// set value or slice if forced
	if x.forceList[name] == true || x.forceList[item.path] == true {
		item.data[name] = []any{value}
	} else {
		item.data[name] = value
	}
}

func (x *Xqml) setText(curr *elem, parent *elem, value any) {
	// if curr is empty => downgrade parent
	if len(curr.data) == 0 {
		x.downgrade(curr, parent, value)
		return
	}
	// set value
	if text, ok := curr.data["#text"]; ok {
		value = fmt.Sprintf("%v%v", text, value)
	}
	curr.data["#text"] = value
}

func (x *Xqml) downgrade(curr *elem, parent *elem, value any) {
	curr.isValue = true
	data := parent.data[curr.name]
	switch data.(type) {
	case []any:
		slice := data.([]any)
		slice[len(slice)-1] = value
	default:
		if x.forceList[curr.name] == true || x.forceList[curr.path] == true {
			parent.data[curr.name] = []any{value}
		} else {
			parent.data[curr.name] = value
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
