package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/njhale/grpctl/internal"
)

func rootCmd() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "grpctl",
		Short: "",
		Long:  ``,
		// Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	root.SetOut(os.Stdout)

	configFS, err := internal.DefaultConfigFS()
	if err != nil {
		return nil, err
	}
	store := internal.NewServerFileStore(configFS)

	root.AddCommand(configCmd(store))

	// Add generated server commands
	servers, err := store.List()
	if err != nil {
		return nil, err
	}

	discovery, err := internal.NewServiceDiscovery(internal.WithBlockingDial(context.TODO(), grpc.WithInsecure()))
	if err != nil {
		return nil, err
	}

	var sub string
	if len(os.Args) > 1 {
		sub = os.Args[1]
	}
	var cmd *cobra.Command
	for _, server := range servers {
		if server.Name == sub {
			cmd, err = serverCmd(discovery, &server)
		} else {
			cmd, err = serverCmd(nil, &server)
		}
		if err != nil {
			return nil, err
		}

		root.AddCommand(cmd)
	}

	return root, nil
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
