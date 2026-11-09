// Package profilegraph produces the offline curator-profile-graph/1 contract.
package profilegraph

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
)

// SchemaVersion identifies the closed public wire contract.
const SchemaVersion = "curator-profile-graph/1"

// Range is a half-open byte interval.
type Range struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// Content measures source bytes only when they were actually read.
type Content struct {
	Status string  `json:"status"`
	Bytes  *int    `json:"bytes"`
	Lines  *int    `json:"lines"`
	SHA256 *string `json:"sha256"`
}

// Revision separates commit and state identities from unknown provenance.
type Revision struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// Source records logical package provenance without local absolute paths.
type Source struct {
	Locator         string   `json:"locator"`
	Directory       string   `json:"directory"`
	Version         *string  `json:"version"`
	Revision        Revision `json:"revision"`
	Path            *string  `json:"path"`
	AuthoringSource *string  `json:"authoring_source"`
	Ownership       string   `json:"ownership"`
}

// Requirement preserves the declared selector rather than its resolved pin.
type Requirement struct {
	Form  string `json:"form"`
	Value string `json:"value"`
}

// Composition identifies a root or overlay declaration and effective order.
type Composition struct {
	Requirement Requirement `json:"requirement"`
	Weight      int64       `json:"weight"`
	Order       int         `json:"order"`
	Declaration string      `json:"declaration"`
}

// Load distinguishes rendering evidence from unknown runtime injection.
type Load struct {
	Environment string  `json:"environment"`
	Class       string  `json:"class"`
	State       string  `json:"state"`
	Basis       string  `json:"basis"`
	Condition   *string `json:"condition"`
}

// Emission measures an inclusive node projection in one emitted document.
type Emission struct {
	Document string  `json:"document"`
	Ranges   []Range `json:"ranges"`
	Bytes    int     `json:"bytes"`
	Lines    int     `json:"lines"`
	SHA256   string  `json:"sha256"`
}

// Heading identifies a Markdown section's level and decoded title.
type Heading struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
}

// Skill records inventory origin and known materialization evidence.
type Skill struct {
	Origin          string `json:"origin"`
	Materialization string `json:"materialization"`
}

// Node is an identity-bearing graph entity with source and emission metrics.
type Node struct {
	ID           string       `json:"id"`
	Identity     []string     `json:"identity"`
	Kind         string       `json:"kind"`
	Label        string       `json:"label"`
	Source       *Source      `json:"source"`
	Layer        *string      `json:"layer"`
	Content      Content      `json:"content"`
	Emissions    []Emission   `json:"emissions"`
	Loads        []Load       `json:"loads"`
	Composition  *Composition `json:"composition,omitempty"`
	Heading      *Heading     `json:"heading,omitempty"`
	SourceRanges []Range      `json:"source_ranges,omitempty"`
	Skill        *Skill       `json:"skill,omitempty"`
}

// Edge records a directed relationship and its declaration order.
type Edge struct {
	From        string       `json:"from"`
	To          string       `json:"to"`
	Kind        string       `json:"kind"`
	Order       int          `json:"order"`
	Requirement *Requirement `json:"requirement,omitempty"`
}

// Input records a consumed file or canonical restricted projection.
type Input struct {
	Root   string `json:"root"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int    `json:"bytes"`
}

// Coverage assigns a disjoint document byte interval to one leaf node.
type Coverage struct {
	Node  string `json:"node"`
	Range Range  `json:"range"`
}

// Document describes actual emitted bytes and their additive coverage ledger.
type Document struct {
	Node        string     `json:"node"`
	Environment string     `json:"environment"`
	Root        string     `json:"root"`
	Path        string     `json:"path"`
	Form        string     `json:"form"`
	Origin      string     `json:"origin"`
	Bytes       int        `json:"bytes"`
	Lines       int        `json:"lines"`
	SHA256      string     `json:"sha256"`
	Coverage    []Coverage `json:"coverage"`
}

// Diagnostic records an explicit uncertainty without claiming runtime facts.
type Diagnostic struct {
	Class   string  `json:"class"`
	Node    *string `json:"node"`
	Message string  `json:"message"`
}

// Graph is the complete atomic curator-profile-graph/1 result.
type Graph struct {
	SchemaVersion string `json:"schema_version"`
	Generator     struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"generator"`
	InputsDigest string `json:"inputs_digest"`
	Scope        struct {
		Profile      string   `json:"profile"`
		Mode         string   `json:"mode"`
		Environments []string `json:"environments"`
	} `json:"scope"`
	Inputs      []Input      `json:"inputs"`
	Nodes       []*Node      `json:"nodes"`
	Edges       []Edge       `json:"edges"`
	Documents   []Document   `json:"documents"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

func hash(raw []byte) string { return fmt.Sprintf("%x", sha256.Sum256(raw)) }
func lines(raw []byte) int {
	n := bytes.Count(raw, []byte{'\n'})
	if len(raw) > 0 && raw[len(raw)-1] != '\n' {
		n++
	}
	return n
}
func ptr[T any](v T) *T { return &v }
func measured(raw []byte) Content {
	return Content{Status: "text", Bytes: ptr(len(raw)), Lines: ptr(lines(raw)), SHA256: ptr(hash(raw))}
}
func tupleID(tuple []string) string {
	var b bytes.Buffer
	for _, s := range tuple {
		b.WriteString(strconv.Itoa(len(s)))
		b.WriteByte(':')
		b.WriteString(s)
	}
	return "n:" + hash(b.Bytes())
}

// Limits are injectable for bounded controls. Production always uses V1Limits.
type Limits struct{ Depth, Nodes, Edges, Entries, Documents, Environments, Output, Input, File int }

// V1Limits are the minimum production budgets promised by the v1 contract.
var V1Limits = Limits{128, 20000, 50000, 50000, 256, 256, 8 << 20, 32 << 20, 4 << 20}

// Error exposes only a named fatal class, never sensitive input values.
type Error struct{ Class string }

func (e *Error) Error() string   { return e.Class + ": graph input refused" }
func refusal(class string) error { return &Error{class} }
