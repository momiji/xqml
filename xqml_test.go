package xqml

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func Test_Parse(t *testing.T) {
	// empty
	testParse(t, "", "{}", "", false, nil)
	// root simple
	testParse(t, `<r>1</r>`, `{"r":1}`, "", false, nil)
	testParse(t, `<r>1</r>`, `{"r":1}`, "", true, nil)
	testParse(t, `<n:r>1</n:r>`, `{"n:r":1}`, "", true, nil)
	// root attributes
	testParse(t, `<r x="1">1</r>`, `{"r":{"#text":1,"@x":"1"}}`, "", false, nil)
	// level2 simple
	testParse(t, `<r><e>1</e></r>`, `{"r":{"e":1}}`, "", false, nil)
	// mixed #text
	testParse(t, `<r>y<e>1</e>x</r>`, `{"r":{"#text":"yx","e":1}}`, `<r>yx<e>1</e></r>`, false, nil)
	// bool
	testParse(t, `<r><e>true</e></r>`, `{"r":{"e":true}}`, "", false, nil)
	testParse(t, `<r><e>True</e></r>`, `{"r":{"e":true}}`, `<r><e>true</e></r>`, false, nil)
	testParse(t, `<r><e>false</e></r>`, `{"r":{"e":false}}`, "", false, nil)
	testParse(t, `<r><e>False</e></r>`, `{"r":{"e":false}}`, `<r><e>false</e></r>`, false, nil)
	// invalid bool
	testParse(t, `<r><e>TRUE</e></r>`, `{"r":{"e":"TRUE"}}`, "", false, nil)
	testParse(t, `<r><e>FALSE</e></r>`, `{"r":{"e":"FALSE"}}`, "", false, nil)
	// mixed bool
	testParse(t, `<r>true<e>1</e></r>`, `{"r":{"#text":true,"e":1}}`, "", false, nil)
	testParse(t, `<r>true<e>1</e>y</r>`, `{"r":{"#text":"truey","e":1}}`, `<r>truey<e>1</e></r>`, false, nil)
	// xml declaration
	testParse(t, `<?xml version="1.0"?><r>1</r>`, `{"r":1}`, `<r>1</r>`, false, nil)
	// array
	testParse(t, `<r><e>1</e><e>2</e></r>`, `{"r":{"e":[1,2]}}`, "", false, nil)
	testParse(t, `<r>y<e>1</e><e>2</e>x</r>`, `{"r":{"#text":"yx","e":[1,2]}}`, `<r>yx<e>1</e><e>2</e></r>`, false, nil)
	// force list
	testParse(t, `<r><e>1</e></r>`, `{"r":{"e":[1]}}`, "", false, []string{"e"})
	testParse(t, `<r><e>1</e></r>`, `{"r":{"e":[1]}}`, "", false, []string{"r.e"})
	// null
	testParse(t, `<r><e></e></r>`, `{"r":{"e":null}}`, "", false, nil)
	testParse(t, `<r><e></e><e>1</e></r>`, `{"r":{"e":[null,1]}}`, "", false, nil)
	testParse(t, `<r><e x="2"></e><e>1</e></r>`, `{"r":{"e":[{"@x":"2"},1]}}`, "", false, nil)
	testParse(t, `<r><e></e><e x="3">1</e></r>`, `{"r":{"e":[null,{"#text":1,"@x":"3"}]}}`, "", false, nil)
	testParse(t, `<r><e x="2"></e><e x="3">1</e></r>`, `{"r":{"e":[{"@x":"2"},{"#text":1,"@x":"3"}]}}`, "", false, nil)
	// cr
	testParse(t, `<r>1</r>\n`, `{"r":1}`, `<r>1</r>`, false, nil)
	testParse(t, `<r>1</r>\n   \n   `, `{"r":1}`, `<r>1</r>`, false, nil)
}

func testParse(t *testing.T, src string, rjson string, rxml string, keepNs bool, forceList []string) {
	x := NewXQML()
	x.Namespace(keepNs)
	x.ForceList(forceList...)
	//
	fmt.Printf("XML => JSON: %s => %s\n", src, rjson)
	src = strings.ReplaceAll(src, "\\n", "\n")
	reader := strings.NewReader(src)
	json, err := x.ParseXml(reader, true)
	if err != nil {
		t.Errorf("%v", err)
	}
	src = strings.ReplaceAll(src, "\n", "\\n")
	res := Stringify(json)
	if rjson != res {
		t.Errorf("received %s\n", res)
	}
	//
	if rxml == "" {
		rxml = src
	}
	fmt.Printf("JSON => XML: %s => %s\n", rjson, rxml)
	writer := new(bytes.Buffer)
	err = x.WriteXml(writer, json)
	if err != nil {
		t.Errorf("%v", err)
	}
	res = writer.String()
	if rxml != res {
		t.Errorf("received %s\n", res)
	}
}
