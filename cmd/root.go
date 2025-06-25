package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AyCarlito/go-mod-visualization/pkg/logger"
	"github.com/AyCarlito/go-mod-visualization/pkg/visualizer"
)

func init() {
	rootCmd.PersistentFlags().StringVar(&inputFilePath, "input", "", "Path to input file. Reads from stdin if unset.")
	rootCmd.PersistentFlags().StringVar(&outputFilePath, "output", "", "Path to output file. Writes to stdout if unset.")
	rootCmd.PersistentFlags().StringVar(&format, "format", "dot", "Output format. Must be one of 'dot' or 'html'.")
}

// CLI Flags
var (
	inputFilePath  string
	outputFilePath string
	format         string
)

var rootCmd = &cobra.Command{
	Use:           "go-mod-visualization",
	Short:         "Visualize go module dependencies.",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Build logger.
		log, err := logger.NewZapConfig().Build()
		if err != nil {
			panic(fmt.Errorf("failed to build zap logger: %v", err))
		}
		cmd.SetContext(logger.ContextWithLogger(cmd.Context(), log))

		return visualizer.NewVisualizer(cmd.Context(), inputFilePath, outputFilePath, format).Visualize()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		// By default, cobra prints the error and usage string on every error.
		// We only desire this behaviour in the case where command line parsing fails e.g. unknown command or flag.
		// Cobra does not provide a mechanism for achieving this fine grain control, so we implement our own.
		if strings.Contains(err.Error(), "command") || strings.Contains(err.Error(), "flag") {
			// Parsing errors are printed along with the usage string.
			fmt.Println(err.Error())
			fmt.Println(rootCmd.UsageString())
		} else {
			// Other errors logged, no usage string displayed.
			log := logger.LoggerFromContext(rootCmd.Context())
			log.Error(err.Error())
		}
		os.Exit(1)
	}
}
