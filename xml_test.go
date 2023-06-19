package xqml

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func Test_XmlAuto(t *testing.T) {
	f, err := os.Create("xml_result.txt")
	if err != nil {
		t.Fatal(err)
	}
	done := map[string]bool{}
	xgen(0, 1, func(e string) {
		if _, isDone := done[e]; isDone {
			return
		}
		done[e] = true
		_, err := f.WriteString(fmt.Sprintf("%s\n", e))
		if err != nil {
			t.Fatal(err)
		}

		r, err := decode(e, true, true, nil, true)
		if err != nil {
			t.Fatal(err)
		}
		rr := Stringify(r)
		_, err = f.WriteString(fmt.Sprintf("=> %s\n", rr))
		if err != nil {
			t.Fatal(err)
		}
		ww, err := encode(r)
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.WriteString(fmt.Sprintf("=> %s\n", ww))
		if err != nil {
			t.Fatal(err)
		}
		//
		//ee1 := ee
		//ee1 = regexp.MustCompile("\\[([0-9]+)]").ReplaceAllString(ee1, "$1")
		//ee1 = regexp.MustCompile("\\[({[^{]+})]").ReplaceAllString(ee1, "$1")
		//ee1 = strings.ReplaceAll(ee1, `,"a":[]`, "")
		//ee1 = strings.ReplaceAll(ee1, `"a":[],`, "")
		//ee1 = strings.ReplaceAll(ee1, `{"a":[]}`, "null")
		//ee1 = strings.ReplaceAll(ee1, `,"b":[]`, "")
		//ee1 = strings.ReplaceAll(ee1, `"b":[],`, "")
		//ee1 = strings.ReplaceAll(ee1, `{"b":[]}`, "null")
		//ee1 = strings.ReplaceAll(ee1, `{"#text":"text"}`, `"text"`)
		//ee2 := fmt.Sprintf(`{"root":%s}`, ee1)
		//ee3 := fmt.Sprintf(`{"root":{"element":%s}}`, ee1)
		//if ee1 != rr && ee2 != rr && ee3 != rr {
		//	t.Errorf("\n%v\n%v\n%v\n", ee, ww, rr)
		//	_, err = f.WriteString(fmt.Sprintf("FAIL: %s\n", rr))
		//	if err != nil {
		//		t.Fatal(err)
		//	}
		//}
		_, err = f.WriteString("\n")
		if err != nil {
			t.Fatal(err)
		}
	})
}

func xgen(lvl int, max int, recv func(string)) {
	if lvl > 1 {
		recv("")
		recv("1")
		return
	}
	var fn func(int, []string, string)
	fn = func(i int, e []string, c string) {
		if i == len(e) {
			recv(c)
			return
		}
		f := e[i]
		split := strings.Split(f, ",")
		if len(split) == 1 {
			fn(i+1, e, c+f)
		} else {
			xgen(lvl+1, 3, func(s string) {
				fn(i+1, e, c+split[0]+s+split[1])
			})
		}
	}
	for _, e := range comb(max,
		"", "1", "<a>,</a>", "<b>,</b>", `<a s="s">,</a>`, `<b s="s">,</b>`,
		"2", "<a>,</a>", "<b>,</b>", `<a t="t">,</a>`, `<b t="t">,</b>`,
	) {
		fn(0, e, "")
	}
}
