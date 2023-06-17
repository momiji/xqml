package xqml

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func Test_JsonAuto(t *testing.T) {
	f, err := os.Create("json_result.txt")
	if err != nil {
		t.Fatal(err)
	}
	cgen(0, true, func(e any) {
		ee := Stringify(e)
		_, err := f.WriteString(fmt.Sprintf("%s\n", ee))
		if err != nil {
			t.Fatal(err)
		}
		x := newXqml(true, true, nil, true)
		w := new(bytes.Buffer)
		err = x.WriteXml(w, e)
		if err != nil {
			t.Fatal(err)
		}
		ww := w.String()
		_, err = f.WriteString(fmt.Sprintf("=> %s\n", ww))
		if err != nil {
			t.Fatal(err)
		}
		x = newXqml(true, true, nil, true)
		r, err := x.ParseXml(strings.NewReader(ww), true)
		if err != nil {
			t.Fatal(err)
		}
		rr := Stringify(r)
		_, err = f.WriteString(fmt.Sprintf("=> %s\n", rr))
		if err != nil {
			t.Fatal(err)
		}
		//
		ee1 := ee
		if ee1 == "[]" {
			ee1 = "null"
		}
		ee1 = regexp.MustCompile("\\[([0-9]+)]").ReplaceAllString(ee1, "$1")
		ee1 = regexp.MustCompile("\\[({[^{]+})]").ReplaceAllString(ee1, "$1")
		ee1 = strings.ReplaceAll(ee1, `,"a":[]`, "")
		ee1 = strings.ReplaceAll(ee1, `"a":[],`, "")
		ee1 = strings.ReplaceAll(ee1, `{"a":[]}`, "null")
		ee1 = strings.ReplaceAll(ee1, `,"b":[]`, "")
		ee1 = strings.ReplaceAll(ee1, `"b":[],`, "")
		ee1 = strings.ReplaceAll(ee1, `{"b":[]}`, "null")
		ee1 = strings.ReplaceAll(ee1, `{"#text":"text"}`, `"text"`)
		ee2 := fmt.Sprintf(`{"root":%s}`, ee1)
		ee3 := fmt.Sprintf(`{"root":{"element":%s}}`, ee1)
		if ee1 != rr && ee2 != rr && ee3 != rr {
			t.Errorf("\n%v\n%v\n%v\n", ee, ww, rr)
			_, err = f.WriteString(fmt.Sprintf("FAIL: %s\n", rr))
			if err != nil {
				t.Fatal(err)
			}
		}
		_, err = f.WriteString("\n")
		if err != nil {
			t.Fatal(err)
		}
	})
}

func cgen(lvl int, withA bool, recv func(any)) {
	// send numbers 1 and 2
	recv(1)
	recv(2)
	if lvl > 1 {
		return
	}
	// send arrays
	if withA {
		// send array of 0 items
		recv([]any{})
		// send array of 1 item
		cgen(lvl+1, false, func(v any) {
			recv([]any{v})
		})
		// send array of 2 items
		cgen(lvl+1, false, func(v any) {
			cgen(lvl+1, false, func(w any) {
				if !reflect.DeepEqual(v, w) {
					recv([]any{v, w})
				}
			})
		})
	}
	// send map of #text @t @u a b
	for _, e := range comb(99, "#text", "@t", "@u", "a", "b") {
		m := make(map[string]any)
		x := ""
		y := ""
		for _, f := range e {
			if strings.HasPrefix(f, "#") || strings.HasPrefix(f, "@") {
				m[f] = f[1:]
			} else {
				if x == "" {
					x = f
				} else {
					y = f
				}
			}
		}
		if y != "" {
			cgen(lvl+1, true, func(v any) {
				cgen(lvl+1, true, func(w any) {
					if !reflect.DeepEqual(v, w) {
						m[x] = v
						m[y] = w
						recv(m)
					}
				})
			})
		} else if x != "" {
			cgen(lvl+1, true, func(v any) {
				m[x] = v
				recv(m)
			})
		} else {
			recv(m)
		}
	}
}

func comb(max int, a ...string) [][]string {
	res := make([][]string, 0)
	done := map[string]bool{}
	for i := 1; i < 1<<len(a); i++ {
		r := make([]string, 0)
		for j := 0; j < len(a); j++ {
			if i&(1<<j) != 0 {
				r = append(r, a[j])
			}
		}
		if len(r) <= max {
			d := strings.Join(r, " ")
			if _, isDone := done[d]; !isDone {
				res = append(res, r)
				done[d] = true
			}
		}
	}
	return res
}

func clone(v any) any {
	mapvalue, isMap := v.(map[string]any)
	if isMap {
		return cloneMap(mapvalue)
	}
	slicevalue, isSlice := v.([]interface{})
	if isSlice {
		return cloneSlice(slicevalue)
	}
	return v
}

func cloneMap(m map[string]any) map[string]any {
	result := map[string]interface{}{}

	for k, v := range m {
		// Handle maps
		mapvalue, isMap := v.(map[string]any)
		if isMap {
			result[k] = cloneMap(mapvalue)
			continue
		}

		// Handle slices
		slicevalue, isSlice := v.([]interface{})
		if isSlice {
			result[k] = cloneSlice(slicevalue)
			continue
		}

		result[k] = v
	}

	return result
}

func cloneSlice(a []any) []any {
	result := []any{}

	for _, v := range a {
		// Handle maps
		mapvalue, isMap := v.(map[string]any)
		if isMap {
			result = append(result, cloneMap(mapvalue))
			continue
		}

		// Handle slices
		slicevalue, isSlice := v.([]any)
		if isSlice {
			result = append(result, cloneSlice(slicevalue))
			continue
		}

		result = append(result, v)
	}

	return result
}
