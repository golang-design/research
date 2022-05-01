// Copyright 2022 The golang.design Initiative.
// All rights reserved. Created by Changkun Ou <changkun.de>

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"gopkg.in/yaml.v3"
	"mvdan.cc/xurls/v2"
)

var md goldmark.Markdown

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
			extension.Table,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)
}

func usage() {
	fmt.Fprintf(os.Stderr, `pdfgen converts a golang.design research markdown file to a pdf.

usage: pdfgen content/posts/bench-time.md
`)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		usage()
		return
	}
	path := args[0]

	// Only deal with .md files
	if !strings.HasSuffix(path, ".md") {
		log.Fatalf("pdfgen: input file must be a markdown file.")
		return
	}

	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("pdfgen: failed to load the given markdown file.")
	}

	var buf bytes.Buffer
	context := parser.NewContext()
	if err := md.Convert(b, &buf, parser.WithContext(context)); err != nil {
		log.Fatal(err)
	}
	metaData := meta.Get(context)
	convertDate(metaData)

	authors := parseAuthor(b)
	metaData["author"] = authors

	abstrat := parseAbstract(b)
	// https://stackoverflow.com/questions/1919982/regex-smallest-possible-match-or-nongreedy-match
	re := regexp.MustCompile("\\[\\^(.*?)\\]")
	abstrat = re.ReplaceAllString(abstrat, "\\cite{$1}") // use citation key
	metaData["abstract"] = abstrat

	metaData["header-includes"] = `\usepackage{fancyhdr}
    \pagestyle{fancy}
	\fancyhead[LE,RO]{\rightmark}
    \fancyhead[RE,LO]{The golang.design Research}
    \fancyfoot{}
	\fancyfoot[C]{\thepage}`

	body := parseBody(b)
	body = re.ReplaceAllString(body, "\\cite{$1}") // use citation key

	head, err := yaml.Marshal(metaData)
	if err != nil {
		log.Fatalf("pdfgen: failed to construct metadata")
	}

	content := fmt.Sprintf(`---
%v
---
%v
`, string(head), body)

	// Prepare all content.

	ref := "content/posts/ref.tex"
	references := parseReferences(b)
	if err := os.WriteFile(ref, []byte(references), os.ModePerm); err != nil {
		log.Fatalf("pdfgen: cannot create reference file: %v", err)
	}
	defer os.Remove(ref)

	article := "content/posts/article.md"
	if err := os.WriteFile(article, []byte(content), os.ModePerm); err != nil {
		log.Fatalf("pdfgen: cannot create temporary file: %v", err)
	}
	defer os.Remove(article)

	// Generate pdf

	before, after, _ := strings.Cut(strings.TrimSuffix(path, ".md")+".pdf", "/posts")
	dst := before + after
	cmd := exec.Command("pandoc", article, ref,
		"-V", "linkcolor:blue",
		"--pdf-engine=xelatex",
		"-o", dst)
	log.Println(cmd.String())
	if b, err := cmd.CombinedOutput(); err != nil {
		log.Fatal(string(b))
	}
	return
}

func convertDate(metaData map[string]any) {
	dateRaw, ok := metaData["date"]
	if !ok {
		log.Fatalf("pdfgen: metadata missing date information.")
	}
	date, ok := dateRaw.(string)
	if !ok {
		log.Fatalf("pdfgen: metadata contains invalid date format.")
	}
	t, err := time.Parse("2006-01-02T15:04:05Z07:00", date)
	if err != nil {
		log.Fatalf("pdfgen: cannot parse date: %v", err)
	}
	metaData["date"] = t.Format("January 02, 2006")
}

type author struct {
	Name  string
	Email string
}

func (a author) String() string {
	return fmt.Sprintf("%v^[Email: %v]", a.Name, a.Email)
}

func parseAuthor(b []byte) []string {
	s := bufio.NewScanner(bytes.NewReader(b))
	authors := []string{}

	for s.Scan() {
		l := s.Text()
		if strings.HasPrefix(l, "Author(s): ") {
			authorsStr := strings.TrimPrefix(l, "Author(s): ")
			authorList := strings.Split(authorsStr, ", ")

			for _, a := range authorList {
				before, after, ok := strings.Cut(a, "](")
				if !ok {
					continue
				}
				name := strings.TrimPrefix(before, "[")
				email := strings.TrimPrefix(strings.TrimSuffix(after, ")"), "mailto:")
				email = strings.ReplaceAll(email, "[at]", "@")
				authors = append(authors, author{name, email}.String())
			}
		}
	}

	if len(authors) == 0 {
		log.Fatalf(`pdfgen: cannot find authors, make sure the markdown uses the correct convention:

Author(s): [FirstName LastName](mailto:email), [FirstName LastName](mailto:email)`)
	}

	return authors
}

func parseAbstract(b []byte) string {
	content := string(b)

	var ok bool
	_, content, ok = strings.Cut(content, "<!--abstract-->\n")
	if !ok {
		goto err
	}
	content, _, ok = strings.Cut(content, "\n<!--more-->")
	if !ok {
		goto err
	}

	return content

err:
	log.Fatal(`pdfgen: cannot find abstract, make sure the markdown uses the correct convention:

	<!--abstract-->
	abstract content goes here...
	<!--more-->
	`)
	return ""
}

func parseBody(b []byte) string {
	content := string(b)

	var ok bool
	_, content, ok = strings.Cut(content, "\n<!--more-->")
	if !ok {
		goto err
	}
	content, _, ok = strings.Cut(content, "## References")
	if !ok {
		goto err
	}
	return content

err:
	log.Fatal(`pdfgen: cannot find body, make sure the markdown uses the correct convention:

	<!--more-->

	content body...

	## References
	`)
	return ""
}

func parseReferences(b []byte) string {
	content := string(b)

	var ok bool
	_, content, ok = strings.Cut(content, "## References\n")
	if !ok {
		log.Fatal(`pdfgen: cannot find references, make sure the markdown uses the correct convention:

		## References

		[@ou2022bench]: Changkun Ou. 2020. Conduct Reliable Benchmarking in Go. TalkGo Meetup. Virtual Event. March 26. https://golang.design/s/gobench
		`)

	}

	content = strings.ReplaceAll(content, "[^", "\\bibitem{")
	content = strings.ReplaceAll(content, "]:", "}")

	rxStrict := xurls.Strict()
	urls := rxStrict.FindAllString(content, -1)
	for _, url := range urls {
		content = strings.ReplaceAll(content, url, "\\url{"+url+"}")
	}
	return "\\begin{thebibliography}{99}" + content + "\\end{thebibliography}"
}
