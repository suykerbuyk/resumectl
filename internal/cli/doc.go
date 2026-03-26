package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newDocCmd() *cobra.Command {
	var (
		manDir string
		mdDir  string
	)

	cmd := &cobra.Command{
		Use:    "doc",
		Short:  "Generate documentation (man pages or markdown)",
		Long:   "Generate man pages or markdown command reference. Used by the build system.",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := NewRootCmd()
			// Disable auto-generation header for cleaner output
			root.DisableAutoGenTag = true

			if manDir != "" {
				if err := os.MkdirAll(manDir, 0o755); err != nil {
					return fmt.Errorf("create man dir: %w", err)
				}
				header := &doc.GenManHeader{
					Title:   "RESUMECTL",
					Section: "1",
					Source:  "resumectl " + Version,
				}
				if err := doc.GenManTree(root, header, manDir); err != nil {
					return fmt.Errorf("generate man pages: %w", err)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Man pages generated in %s/\n", manDir)
			}

			if mdDir != "" {
				if err := os.MkdirAll(mdDir, 0o755); err != nil {
					return fmt.Errorf("create docs dir: %w", err)
				}
				if err := doc.GenMarkdownTree(root, mdDir); err != nil {
					return fmt.Errorf("generate markdown docs: %w", err)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Markdown docs generated in %s/\n", mdDir)
			}

			if manDir == "" && mdDir == "" {
				return fmt.Errorf("specify --man-dir and/or --md-dir")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&manDir, "man-dir", "", "output directory for man pages")
	cmd.Flags().StringVar(&mdDir, "md-dir", "", "output directory for markdown docs")

	return cmd
}
