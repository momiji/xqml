package xqml

import (
	"bytes"
	"strings"
	"testing"
)

func Test_Parse(t *testing.T) {
	// empty
	testParse(t, "", "{}", "", true, false, nil, false)
	// root simple
	testParse(t, `<r>1</r>`, `{"r":1}`, "", true, false, nil, false)
	testParse(t, `<r>1</r>`, `{"r":1}`, "", true, true, nil, false)
	testParse(t, `<n:r>1</n:r>`, `{"n:r":1}`, "", true, true, nil, false)
	// root attributes
	testParse(t, `<r x="1">1</r>`, `{"r":{"#text":1,"@x":"1"}}`, "", true, false, nil, false)
	// level2 simple
	testParse(t, `<r><e>1</e></r>`, `{"r":{"e":1}}`, "", true, false, nil, false)
	// mixed #text
	testParse(t, `<r>y<e>1</e>x</r>`, `{"r":{"#text":"yx","e":1}}`, `<r>yx<e>1</e></r>`, true, false, nil, false)
	// bool
	testParse(t, `<r><e>true</e></r>`, `{"r":{"e":true}}`, "", true, false, nil, false)
	testParse(t, `<r><e>True</e></r>`, `{"r":{"e":true}}`, `<r><e>true</e></r>`, true, false, nil, false)
	testParse(t, `<r><e>false</e></r>`, `{"r":{"e":false}}`, "", true, false, nil, false)
	testParse(t, `<r><e>False</e></r>`, `{"r":{"e":false}}`, `<r><e>false</e></r>`, true, false, nil, false)
	// invalid bool
	testParse(t, `<r><e>TRUE</e></r>`, `{"r":{"e":"TRUE"}}`, "", true, false, nil, false)
	testParse(t, `<r><e>FALSE</e></r>`, `{"r":{"e":"FALSE"}}`, "", true, false, nil, false)
	// mixed bool
	testParse(t, `<r>true<e>1</e></r>`, `{"r":{"#text":true,"e":1}}`, "", true, false, nil, false)
	testParse(t, `<r>true<e>1</e>y</r>`, `{"r":{"#text":"truey","e":1}}`, `<r>truey<e>1</e></r>`, true, false, nil, false)
	// xml declaration
	testParse(t, `<?xml version="1.0"?><r>1</r>`, `{"r":1}`, `<r>1</r>`, true, false, nil, false)
	// array
	testParse(t, `<r><e>1</e><e>2</e></r>`, `{"r":{"e":[1,2]}}`, "", true, false, nil, false)
	testParse(t, `<r>y<e>1</e><e>2</e>x</r>`, `{"r":{"#text":"yx","e":[1,2]}}`, `<r>yx<e>1</e><e>2</e></r>`, true, false, nil, false)
	// force list
	testParse(t, `<r><e>1</e></r>`, `{"r":{"e":[1]}}`, "", true, false, []string{"e"}, false)
	testParse(t, `<r><e>1</e></r>`, `{"r":{"e":[1]}}`, "", true, false, []string{"r.e"}, false)
	testParse(t, `<r>y<e>1</e>x</r>`, `{"r":{"#text":"yx","e":[1]}}`, `<r>yx<e>1</e></r>`, true, false, []string{"r.e"}, false)
	// null
	testParse(t, `<r><e></e></r>`, `{"r":{"e":null}}`, "", true, false, nil, false)
	testParse(t, `<r><e></e><e>1</e></r>`, `{"r":{"e":[null,1]}}`, "", true, false, nil, false)
	testParse(t, `<r><e x="2"></e><e>1</e></r>`, `{"r":{"e":[{"@x":"2"},1]}}`, "", true, false, nil, false)
	testParse(t, `<r><e></e><e x="3">1</e></r>`, `{"r":{"e":[null,{"#text":1,"@x":"3"}]}}`, "", true, false, nil, false)
	testParse(t, `<r><e x="2"></e><e x="3">1</e></r>`, `{"r":{"e":[{"@x":"2"},{"#text":1,"@x":"3"}]}}`, "", true, false, nil, false)
	// cr
	testParse(t, `<r>1</r>\n`, `{"r":1}`, `<r>1</r>`, true, false, nil, false)
	testParse(t, `<r>1</r>\n   \n   `, `{"r":1}`, `<r>1</r>`, true, false, nil, false)
	// html
	testParse(t, `<r><e>1<br>2</e></r>`, `{"r":{"e":{"#text":"1 2","br":null}}}`, `<r><e>1 2<br></br></e></r>`, true, false, nil, true)
	testParse(t, `<r><e>1<br>2</e></r>`, `{"r":{"e":[{"#text":"1 2","br":null}]}}`, `<r><e>1 2<br></br></e></r>`, true, false, []string{"r.e"}, true)
}

func testParse(t *testing.T, src string, rjson string, rxml string, keepAttrs bool, keepNs bool, forceList []string, html bool) {
	x := NewXQML()
	x.SetReadAttributes(keepAttrs)
	x.SetReadNamespaces(keepNs)
	x.SetReadForceList(forceList...)
	x.SetReadHtml(html)
	//
	t.Logf("")
	t.Logf("xml => json: %s => %s\n", src, rjson)
	src = strings.ReplaceAll(src, "\\n", "\n")
	reader := strings.NewReader(src)
	json, err := x.ParseXml(reader, true)
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}
	src = strings.ReplaceAll(src, "\n", "\\n")
	res := Stringify(json)
	if rjson != res {
		t.Errorf("ERROR: received %s\n", res)
	}
	//
	if rxml == "" {
		rxml = src
	}
	t.Logf("json => xml: %s => %s\n", rjson, rxml)
	writer := new(bytes.Buffer)
	err = x.WriteXml(writer, json)
	if err != nil {
		t.Errorf("ERROR: %v", err)
	}
	res = writer.String()
	if rxml != res {
		t.Errorf("ERROR: received %s\n", res)
	}
}

func Test_Simple(t *testing.T) {
	src := `<r><e>1<br>2</e></r>`
	xq := NewXQML()
	xq.SetReadForceList("r.e")
	xq.SetReadHtml(true)
	// parse
	reader := strings.NewReader(src)
	json, err := xq.ParseXml(reader, true)
	if err != nil {
		t.Errorf("ERROR: %v\n", err)
	}
	t.Logf("%s", Stringify(json))
	// write
	writer := new(bytes.Buffer)
	err = xq.WriteXml(writer, json)
	if err != nil {
		t.Errorf("ERROR: %v\n", err)
	}
	t.Logf("%s", writer.String())
}
