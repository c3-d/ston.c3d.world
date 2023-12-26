package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/gabotechs/dep-tree/internal/config"
	"github.com/gabotechs/dep-tree/internal/js"
	"github.com/gabotechs/dep-tree/internal/language"
	"github.com/gabotechs/dep-tree/internal/python"
	"github.com/gabotechs/dep-tree/internal/rust"
	"github.com/gabotechs/dep-tree/internal/utils"
)

var configPath string
var jsFollowTsConfigPaths bool
var jsFollowWorkspaces bool
var pythonExcludeConditionalImports bool
var followReExports bool
var exclude []string

var root *cobra.Command

func NewRoot(args []string) *cobra.Command {
	if args == nil {
		args = os.Args[1:]
	}

	root = &cobra.Command{
		Use:          "dep-tree",
		Version:      "v0.15.10",
		Short:        "Visualize and check your project's dependency tree",
		SilenceUsage: true,
		Args:         cobra.ArbitraryArgs,
		Long: `
      ____         _ __       _
     |  _ \   ___ |  _ \    _| |_  _ __  ___   ___
     | | | | / _ \| |_) |  |_   _||  __|/ _ \ / _ \
     | |_| ||  __/| .__/     | |  | |  |  __/|  __/
     |____/  \__| |_|        | \__|_|   \___| \___|
`,
	}

	root.SetArgs(args)

	root.AddCommand(RenderCmd())
	root.AddCommand(CheckCmd())
	root.AddCommand(EntropyCmd())
	root.AddCommand(ConfigCmd())

	root.PersistentFlags().StringVarP(&configPath, "config", "c", "", "path to dep-tree's config file (default .dep-tree.yml)")
	// TODO: call this '--unwrap-exports'
	root.PersistentFlags().BoolVar(&followReExports, "follow-re-exports", true, "whether to follow re-exports or not while resolving imports between files")
	// TODO: call this '--js-tsconfig-paths'
	root.PersistentFlags().BoolVar(&jsFollowTsConfigPaths, "js-follow-ts-config-paths", false, "whether to follow the tsconfig.json paths while resolving imports or not (default false)")
	// TODO: call this '--js-workspaces'
	root.PersistentFlags().BoolVar(&jsFollowWorkspaces, "js-follow-workspaces", false, "whether to take the workspaces attribute in the root package.json into account or not (default false)")
	root.PersistentFlags().BoolVar(&pythonExcludeConditionalImports, "python-exclude-conditional-imports", false, "exclude conditional imports while calculating file dependencies, like imports wrapped inside if statements (default false)")
	root.PersistentFlags().StringArrayVar(&exclude, "exclude", nil, "Files that match this glob pattern will be ignored. You can provide an arbitrary number of --exclude flags")

	switch {
	case len(args) > 0 && utils.InArray(args[0], []string{"help", "completion", "-v", "--version", "-h", "--help"}):
		// do nothing.
	case len(args) == 0:
		// if not args where provided, default to help.
		root.SetArgs([]string{"help"})
	default:
		// if some args where provided, but it's none of the root commands,
		// choose a default command.
		cmd, _, err := root.Find(args)
		if err == nil && cmd.Use == root.Use && !errors.Is(cmd.Flags().Parse(args), pflag.ErrHelp) {
			root.SetArgs(append([]string{"render"}, args...))
		}
	}
	return root
}

func makeParserBuilder(entrypoint string, cfg *config.Config) (language.NodeParserBuilder, error) {
	switch {
	case utils.EndsWith(entrypoint, js.Extensions):
		return language.ParserBuilder(js.MakeJsLanguage, &cfg.Js, cfg), nil
	case utils.EndsWith(entrypoint, rust.Extensions):
		return language.ParserBuilder(rust.MakeRustLanguage, &cfg.Rust, cfg), nil
	case utils.EndsWith(entrypoint, python.Extensions):
		return language.ParserBuilder(python.MakePythonLanguage, &cfg.Python, cfg), nil
	default:
		return nil, fmt.Errorf("file \"%s\" not supported", entrypoint)
	}
}

func loadConfig() (*config.Config, error) {
	cfg, err := config.ParseConfig(configPath)
	if err != nil {
		return nil, err
	}
	if root.PersistentFlags().Changed("follow-re-exports") {
		cfg.FollowReExports = followReExports
	}
	if root.PersistentFlags().Changed("js-follow-ts-config-paths") {
		cfg.Js.FollowTsConfigPaths = jsFollowTsConfigPaths
	}
	if root.PersistentFlags().Changed("js-follow-workspaces") {
		cfg.Js.FollowWorkspaces = jsFollowWorkspaces
	}
	if root.PersistentFlags().Changed("python-exclude-conditional-imports") {
		cfg.Python.ExcludeConditionalImports = pythonExcludeConditionalImports
	}
	cfg.Exclude = append(cfg.Exclude, exclude...)
	// validate exclusion patterns.
	for _, exclusion := range cfg.Exclude {
		if _, err := utils.GlobstarMatch(exclusion, ""); err != nil {
			return nil, fmt.Errorf("exclude pattern '%s' is not correctly formatted", exclusion)
		}
	}
	return cfg, nil
}
