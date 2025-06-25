package visualizer

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"golang.org/x/mod/semver"
)

//go:embed graph.dot.tmpl
var dotTemplate string

//go:embed graph.html.tmpl
var htmlTemplate string

// Edge is a connection beteeen two nodes.
type Edge struct {
	Src string
	Dst string
}

// Graph is a heirarchical representation of dependencies in a Go module.
type Graph struct {
	// Root is the module being represented.
	Root string

	// Edges are the edges in the graph.
	Edges []Edge

	// Selected are the dependencies included when the Go module is built.
	// It maps a module path to its latest version.
	Selected map[string]string

	// Unselected are the dependencies that are not included when the Go module is built.
	Unselected map[string]struct{}

	// format is the output format of the graph.
	format string
}

// newGraph returns a new *graph.
func newGraph(format string) *Graph {
	return &Graph{
		Selected:   make(map[string]string),
		Unselected: make(map[string]struct{}),
		format:     format,
	}
}

// AddEdge adds an Edge to the Graph.
func (g *Graph) AddEdge(src, dst string) {
	g.Edges = append(g.Edges, Edge{
		Src: src,
		Dst: dst,
	})
}

// AddNode adds a node to the Graph.
func (g *Graph) AddNode(node string) {
	// Extract the path version of the node.
	// Root node has no version.
	var path, version string
	if i := strings.Index(node, "@"); i < 0 {
		g.Root = node
		return
	} else {
		path = node[:i]
		version = node[i+1:]
	}

	// If haven't already encountered this path, this version is necessarily the latest version.
	if latestVersion, ok := g.Selected[path]; !ok {
		g.Selected[path] = version
	} else {
		// Already stored.
		if version == latestVersion {
			return
		}
		if semver.Compare(version, latestVersion) > 0 {
			// This version is now the latest version.
			// Store it, and add the previous latest version as an unselected dependency.
			g.Selected[path] = version
			g.Unselected[fmt.Sprintf("%s@%s", path, latestVersion)] = struct{}{}
		} else {
			// Add this version as an unselected dependency.
			g.Unselected[node] = struct{}{}
		}
	}
}

// String returns the string representation of the Graph.
func (g *Graph) String() (string, error) {
	var outputTemplate string
	switch g.format {
	case "dot":
		outputTemplate = dotTemplate
	case "html":
		outputTemplate = htmlTemplate
	default:
		return "", fmt.Errorf("invalid output format: %s", g.format)
	}

	t, err := template.New("graph").Funcs(template.FuncMap{"split": strings.Split}).Parse(outputTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %v", err)
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, g)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %v", err)
	}

	return tpl.String(), nil
}
