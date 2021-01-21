// Copyright (c) 2020 The golang.design Initiative Authors.
// All rights reserved.
//
// The code below is produced by Changkun Ou <hi@changkun.de>.

// +build ignore

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"strings"
	"text/template"
)

var (
	head = `// Code generated by go run gen.go; DO NOT EDIT.
package fields_test

import "testing"
`
	structTmpl = template.Must(template.New("ss").Parse(`
type {{.Name}} struct {
	{{.Properties}}
}

func (s {{.Name}}) addv(ss {{.Name}}) {{.Name}} {
	return {{.Name}}{
		{{.Addv}}
	}
}

func (s *{{.Name}}) addp(ss *{{.Name}}) *{{.Name}} {
	{{.Addp}}
	return s
}
`))
	benchHead = `func BenchmarkVec(b *testing.B) {`
	benchTail = `}`
	benchBody = template.Must(template.New("bench").Parse(`
	b.Run("addv-{{.Name}}", func(b *testing.B) {
		{{.InitV}}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				v1 = v1.addv(v2)
			} else {
				v2 = v2.addv(v1)
			}
		}
	})
	b.Run("addp-{{.Name}}", func(b *testing.B) {
		{{.InitP}}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if i%2 == 0 {
				v1 = v1.addp(v2)
			} else {
				v2 = v2.addp(v1)
			}
		}
	})
`))
)

type structFields struct {
	Name       string
	Properties string
	Addv       string
	Addp       string
}
type benchFields struct {
	Name  string
	InitV string
	InitP string
}

func main() {
	w := new(bytes.Buffer)
	w.WriteString(head)

	N := 10

	for i := 0; i < N; i++ {
		var (
			ps   = []string{}
			adv  = []string{}
			adpl = []string{}
			adpr = []string{}
		)
		for j := 0; j <= i; j++ {
			ps = append(ps, fmt.Sprintf("x%d\tfloat64", j))
			adv = append(adv, fmt.Sprintf("s.x%d + ss.x%d,", j, j))
			adpl = append(adpl, fmt.Sprintf("s.x%d", j))
			adpr = append(adpr, fmt.Sprintf("s.x%d + ss.x%d", j, j))
		}
		err := structTmpl.Execute(w, structFields{
			Name:       fmt.Sprintf("s%d", i),
			Properties: strings.Join(ps, "\n"),
			Addv:       strings.Join(adv, "\n"),
			Addp:       strings.Join(adpl, ",") + " = " + strings.Join(adpr, ","),
		})
		if err != nil {
			panic(err)
		}
	}

	w.WriteString(benchHead)
	for i := 0; i < N; i++ {
		nums1, nums2 := []string{}, []string{}
		for j := 0; j <= i; j++ {
			nums1 = append(nums1, fmt.Sprintf("%d", j))
			nums2 = append(nums2, fmt.Sprintf("%d", j+i))
		}
		numstr1 := strings.Join(nums1, ", ")
		numstr2 := strings.Join(nums2, ", ")

		err := benchBody.Execute(w, benchFields{
			Name: fmt.Sprintf("s%d", i),
			InitV: fmt.Sprintf(`v1 := s%d{%s}
v2 := s%d{%s}`, i, numstr1, i, numstr2),
			InitP: fmt.Sprintf(`v1 := &s%d{%s}
			v2 := &s%d{%s}`, i, numstr1, i, numstr2),
		})
		if err != nil {
			panic(err)
		}
	}
	w.WriteString(benchTail)

	out, err := format.Source(w.Bytes())
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile("impl_test.go", out, 0660); err != nil {
		panic(err)
	}
}
