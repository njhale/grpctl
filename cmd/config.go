package cmd

import (
	"fmt"
	"io"

	"github.com/njhale/grpctl/internal"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

type serverStore interface {
	Set(internal.Server) error
	Get(string) (internal.Server, error)
	Remove(string) error
	List() ([]internal.Server, error)
}

func configCmd(store serverStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "",
		Long:  ``,
	}

	cmd.AddCommand(configSetCmd(store))
	cmd.AddCommand(configGetCmd(store))
	cmd.AddCommand(configListCmd(store))
	cmd.AddCommand(configRemoveCmd(store))

	return cmd
}

func configSetCmd(store serverStore) *cobra.Command {
	return &cobra.Command{
		Use:   "set <server> address|proto <value>",
		Short: "configure settings for interacting with a server",
		Long:  ``,
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(njhale): Use cobra callbacks for argument validation
			name := args[0]
			if name == "server" {
				return fmt.Errorf(`"server" is a reserved word and cannot be used as a server name`)
			}

			server, err := store.Get(name)
			if err != nil {
				return err
			}

			server.Name = name
			switch setting, value := args[1], args[2]; setting {
			case "address":
				server.Address = value
			case "proto":
				server.Proto = value
			default:
				return fmt.Errorf("unknown server setting %s", setting)
			}

			return store.Set(server)
		},
	}
}

func configGetCmd(store serverStore) *cobra.Command {
	return &cobra.Command{
		Use:   "get <server>",
		Short: "get configuration for a server",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			server, err := store.Get(args[0])
			if err != nil {
				return err
			}

			// TODO(njhale): Improve formatting
			printYaml(cmd.OutOrStdout(), server)

			return nil
		},
	}
}

func configListCmd(store serverStore) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list configurations for all known servers",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			servers, err := store.List()
			if err != nil {
				return err
			}

			// TODO(njhale): Improve formatting
			printYaml(cmd.OutOrStdout(), servers)

			return nil
		},
	}
}

func configRemoveCmd(store serverStore) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <server>",
		Short: "remove configuration for a server",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			err := store.Remove(name)
			if err != nil {
				return err
			}

			cmd.Printf("%s removed", name)

			return nil
		},
	}
}

func printYaml(out io.Writer, v interface{}) {
	// TODO(njhale): A nicer way to dynamically format and print to stdout.
	marshalled, err := yaml.Marshal(v)
	if err != nil {
		panic(fmt.Errorf("Failed to marshal message: %w", err))
	}

	_, err = out.Write(marshalled)
	if err != nil {
		panic(fmt.Errorf("Failed to write message: %w", err))
	}
}
