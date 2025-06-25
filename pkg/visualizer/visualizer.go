package visualizer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/AyCarlito/go-mod-visualization/pkg/logger"
)

// Visualizer can take Go module dependencies and generate graphical reprsentations of them.
type Visualizer struct {
	ctx            context.Context
	graph          *Graph
	inputFilePath  string
	outputFilePath string
}

// NewVisualizer returns a new *Visualizer.
func NewVisualizer(ctx context.Context, inputFilePath, outputFilePath, format string) *Visualizer {
	return &Visualizer{
		ctx:            ctx,
		graph:          newGraph(format),
		inputFilePath:  inputFilePath,
		outputFilePath: outputFilePath,
	}
}

// Visualize generates a graphviz graph from Go module depdencies.
func (v *Visualizer) Visualize() error {
	log := logger.LoggerFromContext(v.ctx)
	log.Info("Starting visualization")

	f := os.Stdin
	if v.inputFilePath != "" {
		var err error
		f, err = os.Open(v.inputFilePath)
		if err != nil {
			return fmt.Errorf("failed to open input file: %v", err)
		}
		defer f.Close()
	}

	log.Info("Building graph")
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		// Ignore empty lines.
		if line == "" {
			continue
		}

		// Each line should be composed of two parts:
		// - A module.
		// - A requirement.
		moduleAndRequirement := strings.Split(line, " ")
		if len(moduleAndRequirement) != 2 {
			return fmt.Errorf("line must be formatted as 'path@version requirement': %s", line)
		}

		module := moduleAndRequirement[0]
		requirement := moduleAndRequirement[1]

		// Add a node for the module and requirement.
		v.graph.AddNode(module)
		v.graph.AddNode(requirement)

		// Connect the module and requirement.
		v.graph.AddEdge(module, requirement)
	}

	// EOF errors are excluded here.
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read input: %v", err)

	}

	// Generate the string representation of the graph.
	log.Info("Converting graph")
	dot, err := v.graph.String()
	if err != nil {
		return fmt.Errorf("failed to generate string representation of graph: %v", err)
	}

	f = os.Stdout
	if v.outputFilePath != "" {
		var err error
		f, err = os.Create(v.outputFilePath)
		if err != nil {
			return fmt.Errorf("failed to create output file: %v", err)
		}
		defer f.Close()
	}

	log.Info("Writing output")
	_, err = f.WriteString(dot)
	if err != nil {
		return fmt.Errorf("failed to write output: %v", err)
	}

	return nil
}
