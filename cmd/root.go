package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/home-sol/kalki/cmd/execenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	// GitCommit is the git commit that was compiled. This will be filled in by the compiler.
	GitCommit string
	// GitLastTag is the last git tag that was compiled. This will be filled in by the compiler.
	GitLastTag string
	// GitExactTag is the exact git tag that was compiled. This will be filled in by the compiler.
	GitExactTag string
)

// NewRootCommand creates a new root command.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   execenv.RootCommandName,
		Short: "Kalki is a dns server",
		Long:  `Kalki is a dns server for home network. It provides a simple way to manage dns records.`,

		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			root := cmd.Root()
			root.Version = getVersion()
		},

		// For the root command, force the execution of the PreRun
		// even if we just display the help. This is to make sure that we check
		// the repository and give the user early feedback.
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				os.Exit(1)
			}
		},

		SilenceUsage:      true,
		DisableAutoGenTag: true,
	}

	env := execenv.NewEnv()

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cobra.yaml)")
	cmd.AddCommand(newVersionCommand(env))
	return cmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	cobra.OnInitialize(onInitialize)
	if err := NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

func onInitialize() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("kalki")

		viper.AddConfigPath("/etc/kalki/")

		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(fmt.Sprintf("%s/.kalki", home))
		viper.AddConfigPath(".")
	}

	viper.SetEnvPrefix("kalki")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func getVersion() string {
	if GitExactTag == "undefined" {
		GitExactTag = ""
	}

	if GitExactTag != "" {
		// we are exactly on a tag --> release version
		return GitLastTag
	}

	if GitLastTag != "" {
		// not exactly on a tag --> dev version
		return fmt.Sprintf("%s-dev-%.10s", GitLastTag, GitCommit)
	}

	// we don't have commit information, try golang build info
	if commit, dirty, err := getCommitAndDirty(); err == nil {
		if dirty {
			return fmt.Sprintf("dev-%.10s-dirty", commit)
		}
		return fmt.Sprintf("dev-%.10s", commit)
	}

	return "dev-unknown"
}

func getCommitAndDirty() (commit string, dirty bool, err error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", false, fmt.Errorf("unable to read build info")
	}

	var commitFound bool

	// get the commit and modified status
	// (that is the flag for repository dirty or not)
	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			commit = kv.Value
			commitFound = true
		case "vcs.modified":
			if kv.Value == "true" {
				dirty = true
			}
		}
	}

	if !commitFound {
		return "", false, fmt.Errorf("no commit found")
	}

	return commit, dirty, nil
}
