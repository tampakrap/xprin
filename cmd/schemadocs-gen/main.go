/*
Copyright 2026 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// schemadocs-gen fills the property tables of docs/testsuite-specification.md from the JSON schema written by schema-gen.
// A list of every property, with links to their rows, sits between the DOCS_START_INDEX and DOCS_END_INDEX markers. Each table sits between <!-- DOCS_START_<NAME> --> and <!-- DOCS_END_<NAME> --> markers; the rest of the file is hand-written.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
)

// section is one table of the document: a type of the schema, under the heading that has this anchor.
type section struct {
	name   string // Suffix of the markers: DOCS_START_<name>
	title  string // What the links to this section say
	anchor string // Anchor of its heading in the document
	def    string // Name of the type in $defs
}

// sections are the tables of the document, in order. Every type in $defs is either here or in inlined.
var sections = []section{
	{"ROOT", "root", "root", "TestSuiteSpec"},
	{"COMMON", "common", "common", "Common"},
	{"TESTS", "tests[*]", "tests", "TestCase"},
	{"INPUTS", "inputs", "inputs", "Inputs"},
	{"PATCHES", "patches", "patches", "Patches"},
	{"HOOKS", "hooks", "hooks", "Hooks"},
	{"ASSERTIONS", "assertions", "assertions", "Assertions"},
	{"ASSERTIONS_XPRIN", "assertions.xprin[*]", "assertionsxprin", "AssertionXprin"},
	{"ASSERTIONS_GOLDEN", "assertions.diff[*]` and `assertions.dyff[*]", "assertionsdiff-and-assertionsdyff", "AssertionGoldenFile"},
}

// inlined are the types whose properties are rows of the table of the property that uses them, e.g. `pre-test[*].run`.
var inlined = []string{"Hook"}

const (
	requiredKeyword = "x-xprin-required" // Set by a struct tag in internal/api, on the properties required only in the merged result
	indexName       = "INDEX"
	startMarker     = "<!-- DOCS_START_%s -->"
	endMarker       = "<!-- DOCS_END_%s -->"
)

// node is the part of a JSON schema node that the tables need.
type node struct {
	Ref                  string           `json:"$ref"`
	Type                 string           `json:"type"`
	Description          string           `json:"description"`
	Enum                 []any            `json:"enum"`
	MinItems             *int             `json:"minItems"`
	Items                *node            `json:"items"`
	AdditionalProperties json.RawMessage  `json:"additionalProperties"`
	Properties           json.RawMessage  `json:"properties"`
	Required             []string         `json:"required"`
	Conditional          string           `json:"x-xprin-required"`
	order                []string         // Keys of Properties, in the order of the schema
	props                map[string]*node // Properties, parsed
}

type schema struct {
	Defs map[string]*node `json:"$defs"`
}

func main() {
	schemaPath := flag.String("schema", "", "JSON schema written by schema-gen (required)")
	docPath := flag.String("doc", "", "Document to fill (required)")

	flag.Parse()

	if *schemaPath == "" || *docPath == "" {
		fmt.Fprintf(os.Stderr, "schemadocs-gen: -schema and -doc are required\n")
		os.Exit(1)
	}

	if err := run(*schemaPath, *docPath); err != nil {
		fmt.Fprintf(os.Stderr, "schemadocs-gen: %v\n", err)
		os.Exit(1)
	}
}

func run(schemaPath, docPath string) error {
	raw, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}

	doc, err := os.ReadFile(docPath)
	if err != nil {
		return fmt.Errorf("read doc: %w", err)
	}

	out, err := render(raw, string(doc))
	if err != nil {
		return err
	}

	if out == string(doc) {
		return nil
	}

	if err := os.WriteFile(docPath, []byte(out), 0o644); err != nil { //nolint:gosec // A document, not a secret.
		return fmt.Errorf("write doc: %w", err)
	}

	return nil
}

// render returns doc with the content between each pair of markers replaced by the table of its section.
func render(rawSchema []byte, doc string) (string, error) {
	var s schema
	if err := json.Unmarshal(rawSchema, &s); err != nil {
		return "", fmt.Errorf("parse schema: %w", err)
	}

	for name, def := range s.Defs {
		if err := def.parseProperties(); err != nil {
			return "", fmt.Errorf("parse %s: %w", name, err)
		}
	}

	if err := checkCovered(s.Defs); err != nil {
		return "", err
	}

	for _, sec := range sections {
		def, ok := s.Defs[sec.def]
		if !ok {
			return "", fmt.Errorf("section %s: no type %q in the schema", sec.name, sec.def)
		}

		var err error

		doc, err = fill(doc, sec.name, table(s.Defs, sec, def))
		if err != nil {
			return "", err
		}
	}

	root, ok := sectionOf("TestSuiteSpec")
	if !ok {
		return "", errors.New("no section for TestSuiteSpec")
	}

	return fill(doc, indexName, index(s.Defs, root, s.Defs[root.def]))
}

// sectionOf returns the section of a type.
func sectionOf(def string) (section, bool) {
	for _, sec := range sections {
		if sec.def == def {
			return sec, true
		}
	}

	return section{}, false
}

// index lists the path of every property, from the root of the file, each linking to its row (or to its section, when its type has one).
func index(defs map[string]*node, sec section, def *node) string {
	var b strings.Builder

	indexEntries(&b, defs, sec, def, "", "")

	return b.String()
}

// indexEntries writes the properties of def: rel is their path inside the section sec, shown the one in the file.
func indexEntries(b *strings.Builder, defs map[string]*node, sec section, def *node, rel, shown string) {
	for _, key := range def.order {
		p := def.props[key]
		path := shown + key
		link := sec.anchor + "." + rel + key
		// A property whose type (or the items of it) has a section links to the section
		ref := p.Ref
		if p.Items != nil {
			ref = p.Items.Ref
		}

		if own, ok := sectionOf(refName(ref)); ok {
			link = own.anchor
		}

		fmt.Fprintf(b, "- [`%s`](#%s)\n", path, link)

		name := refName(p.Ref)
		suffix := ""

		if p.Items != nil {
			name = refName(p.Items.Ref)
			suffix = "[*]"
		}

		child := defs[name]
		if child == nil || (p.Ref == "" && p.Items == nil) {
			continue
		}

		if inner, ok := sectionOf(name); ok {
			indexEntries(b, defs, inner, child, "", path+suffix+".")
		} else if slices.Contains(inlined, name) {
			indexEntries(b, defs, sec, child, rel+key+suffix+".", path+suffix+".")
		}
	}
}

// checkCovered fails if a type of the schema has no table, so a new type can't go undocumented.
func checkCovered(defs map[string]*node) error {
	covered := map[string]bool{}
	for _, sec := range sections {
		covered[sec.def] = true
	}

	for _, name := range inlined {
		covered[name] = true
	}

	var missing []string

	for name := range defs {
		if !covered[name] {
			missing = append(missing, name)
		}
	}

	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("types without a table, add them to sections or inlined: %s", strings.Join(missing, ", "))
	}

	return nil
}

// fill replaces what is between the markers of the section.
func fill(doc, name, content string) (string, error) {
	start, end := fmt.Sprintf(startMarker, name), fmt.Sprintf(endMarker, name)

	i := strings.Index(doc, start)
	j := strings.Index(doc, end)

	if i < 0 || j < i {
		return "", fmt.Errorf("the document has no %s ... %s markers (in this order)", start, end)
	}

	return doc[:i+len(start)] + "\n\n" + content + "\n" + doc[j:], nil
}

// parseProperties reads the properties of the node in the order of the schema, which is the one of the Go types.
func (n *node) parseProperties() error {
	if len(n.Properties) == 0 {
		return nil
	}

	dec := json.NewDecoder(bytes.NewReader(n.Properties))

	if _, err := dec.Token(); err != nil { // The opening brace
		return err
	}

	n.props = map[string]*node{}

	for dec.More() {
		key, err := dec.Token()
		if err != nil {
			return err
		}

		var child node
		if err := dec.Decode(&child); err != nil {
			return err
		}

		k, _ := key.(string)
		n.order = append(n.order, k)
		n.props[k] = &child
	}

	return nil
}

// table renders the rows of a type.
func table(defs map[string]*node, sec section, def *node) string {
	var b strings.Builder

	b.WriteString("| Property | Type | Required | Description |\n")
	b.WriteString("| :------- | :--- | :------: | :---------- |\n")

	rows(&b, defs, sec, def, "")

	return b.String()
}

func rows(b *strings.Builder, defs map[string]*node, sec section, def *node, prefix string) {
	for _, key := range def.order {
		p := def.props[key]

		fmt.Fprintf(b, "| <a id=\"%s.%s\"></a>`%s` | %s | %s | %s |\n", sec.anchor, prefix+key, prefix+key, typeOf(defs, p), required(def, key, p), describe(defs, p))

		// The properties of an inlined type are rows of this table
		if target := inlinedTarget(defs, p); target != nil {
			sub := prefix + key
			if p.Type == "array" {
				sub += "[*]"
			}

			rows(b, defs, sec, target, sub+".")
		}
	}
}

// inlinedTarget returns the type that the property holds, or the items of it, when its properties are inlined.
func inlinedTarget(defs map[string]*node, p *node) *node {
	ref := p.Ref
	if p.Items != nil {
		ref = p.Items.Ref
	}

	name := refName(ref)
	if !slices.Contains(inlined, name) {
		return nil
	}

	return defs[name]
}

func refName(ref string) string {
	return strings.TrimPrefix(ref, "#/$defs/")
}

func required(def *node, key string, p *node) string {
	switch {
	case p.Conditional == "conditional":
		return "conditional"
	case slices.Contains(def.Required, key):
		return "yes"
	default:
		return "no"
	}
}

// typeOf renders the type of a property, linking to the section of its type.
func typeOf(defs map[string]*node, p *node) string {
	switch {
	case p.Ref != "":
		return typeOfRef(p.Ref)
	case p.Type == "array":
		t := "list of " + typeOf(defs, p.Items)
		if p.MinItems != nil {
			t += fmt.Sprintf(" (at least %d)", *p.MinItems)
		}

		return t
	case p.Type == "object":
		var values node
		if json.Unmarshal(p.AdditionalProperties, &values) == nil && values.Type != "" {
			return "map of `" + values.Type + "`"
		}

		return "`object`"
	case p.Type == "":
		return "any"
	default:
		return "`" + p.Type + "`"
	}
}

func typeOfRef(ref string) string {
	name := refName(ref)

	for _, sec := range sections {
		if sec.def == name {
			return fmt.Sprintf("[`%s`](#%s)", sec.title, sec.anchor)
		}
	}

	return "`object`" // An inlined type: its properties are the rows that follow
}

// describe returns the description of a property, with the allowed values.
func describe(defs map[string]*node, p *node) string {
	d := p.Description
	if d == "" {
		if ref := defs[refName(p.Ref)]; ref != nil {
			d = ref.Description
		}
	}

	d = strings.TrimSuffix(strings.TrimSuffix(d, " (Optional)"), " (Required)")
	d = strings.ReplaceAll(d, "|", "\\|")

	if len(p.Enum) > 0 {
		var vals []string
		for _, v := range p.Enum {
			vals = append(vals, fmt.Sprintf("`%v`", v))
		}

		d += ". Allowed values: " + strings.Join(vals, ", ")
	}

	return d
}
