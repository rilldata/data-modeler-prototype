package docs

import (
	"fmt"

	"github.com/rilldata/rill/cli/pkg/browser"
	"github.com/spf13/cobra"
)

var docsUrl = "https://docs.rilldata.com"

// docsCmd represents the docs command
func DocsCmd() *cobra.Command {
	var docsCmd = &cobra.Command{
		Use:   "docs",
		Short: "Open docs.rilldata.com",
		Run: func(cmd *cobra.Command, args []string) {
			err := browser.Open(docsUrl)
			if err != nil {
				fmt.Printf("Could not open browser. Copy this URL into your browser: %s\n", docsUrl)
			}
		},
	}
	return docsCmd
}
