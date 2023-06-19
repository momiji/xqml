package xqml

import (
	"bytes"
	"strings"
	"testing"
)

func Test_Parse(t *testing.T) {
	// empty
	testParse(t, "", "{}", "<root></root>", `{"root":null}`, true, false, nil, false)
	// root simple
	testParse(t, `<r>1</r>`, `{"r":1}`, "", "", true, false, nil, false)
	testParse(t, `<r>1</r>`, `{"r":1}`, "", "", true, true, nil, false)
	testParse(t, `<n:r>1</n:r>`, `{"n:r":1}`, "", "", true, true, nil, false)
	// root attributes
	testParse(t, `<r x="1">1</r>`, `{"r":{"#text":1,"@x":"1"}}`, "", "", true, false, nil, false)
	// level2 simple
	testParse(t, `<r><e>1</e></r>`, `{"r":{"e":1}}`, "", "", true, false, nil, false)
	// mixed #text
	testParse(t, `<r>y<e>1</e>x</r>`, `{"r":{"#text":"yx","e":1}}`, `<r>yx<e>1</e></r>`, "", true, false, nil, false)
	// bool
	testParse(t, `<r><e>true</e></r>`, `{"r":{"e":true}}`, "", "", true, false, nil, false)
	testParse(t, `<r><e>True</e></r>`, `{"r":{"e":true}}`, `<r><e>true</e></r>`, "", true, false, nil, false)
	testParse(t, `<r><e>TRUE</e></r>`, `{"r":{"e":true}}`, `<r><e>true</e></r>`, "", true, false, nil, false)
	testParse(t, `<r><e>false</e></r>`, `{"r":{"e":false}}`, "", "", true, false, nil, false)
	testParse(t, `<r><e>False</e></r>`, `{"r":{"e":false}}`, `<r><e>false</e></r>`, "", true, false, nil, false)
	testParse(t, `<r><e>FALSE</e></r>`, `{"r":{"e":false}}`, `<r><e>false</e></r>`, "", true, false, nil, false)
	// invalid bool
	testParse(t, `<r><e>TrUE</e></r>`, `{"r":{"e":"TrUE"}}`, "", "", true, false, nil, false)
	testParse(t, `<r><e>FaLSE</e></r>`, `{"r":{"e":"FaLSE"}}`, "", "", true, false, nil, false)
	// mixed bool
	testParse(t, `<r>true<e>1</e></r>`, `{"r":{"#text":true,"e":1}}`, "", "", true, false, nil, false)
	testParse(t, `<r>true<e>1</e>y</r>`, `{"r":{"#text":"truey","e":1}}`, `<r>truey<e>1</e></r>`, "", true, false, nil, false)
	// xml declaration
	testParse(t, `<?xml version="1.0"?><r>1</r>`, `{"r":1}`, `<r>1</r>`, "", true, false, nil, false)
	// array
	testParse(t, `<r><e>1</e><e>2</e></r>`, `{"r":{"e":[1,2]}}`, "", "", true, false, nil, false)
	testParse(t, `<r>y<e>1</e><e>2</e>x</r>`, `{"r":{"#text":"yx","e":[1,2]}}`, `<r>yx<e>1</e><e>2</e></r>`, "", true, false, nil, false)
	// force list
	testParse(t, `<r><e>1</e></r>`, `{"r":{"e":[1]}}`, "", "", true, false, []string{"e"}, false)
	testParse(t, `<r><e>1</e></r>`, `{"r":{"e":[1]}}`, "", "", true, false, []string{"r.e"}, false)
	testParse(t, `<r>y<e>1</e>x</r>`, `{"r":{"#text":"yx","e":[1]}}`, `<r>yx<e>1</e></r>`, "", true, false, []string{"r.e"}, false)
	// null
	testParse(t, `<r><e></e></r>`, `{"r":{"e":null}}`, "", "", true, false, nil, false)
	testParse(t, `<r><e></e><e>1</e></r>`, `{"r":{"e":[null,1]}}`, "", "", true, false, nil, false)
	testParse(t, `<r><e x="2"></e><e>1</e></r>`, `{"r":{"e":[{"@x":"2"},1]}}`, "", "", true, false, nil, false)
	testParse(t, `<r><e></e><e x="3">1</e></r>`, `{"r":{"e":[null,{"#text":1,"@x":"3"}]}}`, "", "", true, false, nil, false)
	testParse(t, `<r><e x="2"></e><e x="3">1</e></r>`, `{"r":{"e":[{"@x":"2"},{"#text":1,"@x":"3"}]}}`, "", "", true, false, nil, false)
	// cr
	testParse(t, `<r>1</r>\n`, `{"r":1}`, `<r>1</r>`, "", true, false, nil, false)
	testParse(t, `<r>1</r>\n   \n   `, `{"r":1}`, `<r>1</r>`, "", true, false, nil, false)
	// html
	testParse(t, `<r><e>1<br>2</e></r>`, `{"r":{"e":{"#text":"1 2","br":null}}}`, `<r><e>1 2<br></br></e></r>`, "", true, false, nil, true)
	testParse(t, `<r><e>1<br>2</e></r>`, `{"r":{"e":[{"#text":"1 2","br":null}]}}`, `<r><e>1 2<br></br></e></r>`, "", true, false, []string{"r.e"}, true)
	// float
	testParse(t, `<r>1.001</r>`, `{"r":1.001}`, "", "", true, false, nil, false)
}

func testParse(t *testing.T, src string, rjson string, rxml string, rjson2 string, keepAttrs bool, keepNs bool, forceList []string, html bool) {
	//
	t.Logf("")
	t.Logf("xml => json: %s => %s\n", src, rjson)
	src = strings.ReplaceAll(src, "\\n", "\n")
	reader := strings.NewReader(src)
	j, err := newXqml(keepAttrs, keepNs, forceList, html).ParseXml(reader, true)
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}
	src = strings.ReplaceAll(src, "\n", "\\n")
	res := Stringify(j)
	if rjson != res {
		t.Errorf("ERROR: received %s\n", res)
	}
	//
	if rxml == "" {
		rxml = src
	}
	t.Logf("json => xml: %s => %s\n", rjson, rxml)
	writer := new(bytes.Buffer)
	err = newXqml(keepAttrs, keepNs, forceList, html).WriteXml(writer, j)
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}
	res = writer.String()
	if rxml != res {
		t.Errorf("ERROR: received %s\n", res)
	}
	//
	t.Logf("xml => json: %s => %s\n", res, rjson)
	reader = strings.NewReader(res)
	j, err = newXqml(keepAttrs, keepNs, forceList, html).ParseXml(reader, true)
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}
	rjson2 = strings.ReplaceAll(rjson2, "\n", "\\n")
	if rjson2 == "" {
		rjson2 = rjson
	}
	res2 := Stringify(j)
	if rjson2 != res2 {
		t.Errorf("ERROR: received %s\n", res2)
	}
}

func Test_Write(t *testing.T) {
	testWrite(t, `{"r":{"a":[{"b":1},{"b":[2,3]}]}}`, `<r><a><b>1</b></a><a><b>2</b><b>3</b></a></r>`, `{"r":{"a":[{"b":1},{"b":[2,3]}]}}`, true, false, nil, true)
	// special case for first level being a slice, it encapsulates in root
	testWrite(t, `{"a":[{"b":1},{"b":[2,3]}]}`, `<root><a><b>1</b></a><a><b>2</b><b>3</b></a></root>`, `{"root":{"a":[{"b":1},{"b":[2,3]}]}}`, true, false, nil, true)
	//
	testWrite(t, `{"#text":"1"}`, `<root>1</root>`, `{"root":1}`, true, false, nil, true)
	testWrite(t, `{"a":1}`, `<a>1</a>`, ``, true, false, nil, true)
	testWrite(t, `{"a":[1]}`, `<root><a>1</a></root>`, `{"root":{"a":1}}`, true, false, nil, true)
	testWrite(t, `{"a":[1,2]}`, `<root><a>1</a><a>2</a></root>`, `{"root":{"a":[1,2]}}`, true, false, nil, true)
	testWrite(t, `{"a":1,"b":2}`, `<root><a>1</a><b>2</b></root>`, `{"root":{"a":1,"b":2}}`, true, false, nil, true)
	testWrite(t, `{"a":1,"a":2}`, `<a>2</a>`, `{"a":2}`, true, false, nil, true)
	testWrite(t, `{"#text":"1","a":2}`, `<root>1<a>2</a></root>`, `{"root":{"#text":1,"a":2}}`, true, false, nil, true)
	testWrite(t, `{"@u":"1","a":2}`, `<root u="1"><a>2</a></root>`, `{"root":{"@u":"1","a":2}}`, true, false, nil, true)
	testWrite(t, `{"#text":"1","a":2,"@u":"3"}`, `<root u="3">1<a>2</a></root>`, `{"root":{"#text":1,"@u":"3","a":2}}`, true, false, nil, true)
}
func testWrite(t *testing.T, src string, rxml string, rjson string, keepAttrs bool, keepNs bool, forceList []string, html bool) {
	//
	t.Logf("")
	t.Logf("json => xml: %s => %s\n", src, rxml)
	src = strings.ReplaceAll(src, "\\n", "\n")
	v, err := ToJson([]byte(src))
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}
	writer := new(bytes.Buffer)
	err = newXqml(keepAttrs, keepNs, forceList, html).WriteXml(writer, v)
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}
	res := writer.String()
	if res != rxml {
		t.Errorf("ERROR: received %s\n", res)
	}
	//
	if rjson == "" {
		rjson = src
	}
	t.Logf("xml => json: %s => %s\n", rxml, rjson)
	reader := strings.NewReader(rxml)
	j, err := newXqml(keepAttrs, keepNs, forceList, html).ParseXml(reader, true)
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}
	res = Stringify(j)
	if rjson != res {
		t.Errorf("ERROR: received %s\n", res)
	}
}

func Test_Simple(t *testing.T) {
	src := `<r><e>1<br>2</e></r>`
	// parse
	reader := strings.NewReader(src)
	json, err := newXqml(true, true, []string{"r.e"}, true).ParseXml(reader, true)
	if err != nil {
		t.Errorf("ERROR: %v\n", err)
	}
	t.Logf("%s", Stringify(json))
	// write
	writer := new(bytes.Buffer)
	err = newXqml(true, true, []string{"r.e"}, true).WriteXml(writer, json)
	if err != nil {
		t.Errorf("ERROR: %v\n", err)
	}
	t.Logf("%s", writer.String())
}

func newXqml(keepAttrs bool, keepNs bool, forceList []string, html bool) *Xqml {
	x := NewXQML()
	x.SetReadAttributes(keepAttrs)
	x.SetReadNamespaces(keepNs)
	x.SetReadForceList(forceList...)
	x.SetReadHtml(html)
	return x
}
