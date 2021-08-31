package cmd

import (
	"fmt"
	"os"

	"github.com/njhale/grpctl/internal"
	"github.com/spf13/cobra"
)

func rootCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "grpctl",
		Short: "",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Printf("in root with args: %v\n", args)
			return nil
		},
	}

	cmd.SetOut(os.Stdout)

	configFS, err := internal.DefaultConfigFS()
	if err != nil {
		return nil, err
	}
	store := internal.NewServiceFileStore(configFS)

	cmd.AddCommand(serviceCmd(store))

	// Add generated service commands
	services, err := store.List()
	if err != nil {
		return nil, err
	}
	addGeneratedCommands(cmd, services)

	return cmd, nil
}

func addGeneratedCommands(root *cobra.Command, services []internal.Service) {
	for _, service := range services {
		// TODO(njhale): Generate opts, args, and logic for dynamic service commands
		root.AddCommand(&cobra.Command{
			Use:   service.Name,
			Short: fmt.Sprintf("call %s at %s", service.Name, service.Address),
			Long:  ``,
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.Printf("in %s with args: %v\n", service.Name, args)
				return nil
			},
		})
	}
}

func Execute() {
	cmd, err := rootCmd()
	if err == nil {
		err = cmd.Execute()
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
