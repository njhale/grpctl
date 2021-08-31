package cmd

import (
	"fmt"
	"io"

	"github.com/njhale/grpctl/internal"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

type serviceStore interface {
	Set(internal.Service) error
	Get(string) (internal.Service, error)
	Remove(string) error
	List() ([]internal.Service, error)
}

func serviceCmd(store serviceStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service",
		Short: "",
		Long:  ``,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(njhale): implement this!
			cmd.Printf("in parent with args: %v\n", args)
			return nil
		},
	}

	cmd.AddCommand(serviceSetCmd(store))
	cmd.AddCommand(serviceGetCmd(store))
	cmd.AddCommand(serviceListCmd(store))
	cmd.AddCommand(serviceRemoveCmd(store))

	return cmd
}

func serviceSetCmd(store serviceStore) *cobra.Command {
	return &cobra.Command{
		Use:   "set <name> [address|proto] <value>",
		Short: "configure settings for interacting with a service",
		Long:  ``,
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			service, err := store.Get(name)
			if err != nil {
				return err
			}

			// TODO(njhale): Use cobra callbacks for argument validation
			service.Name = name
			switch setting, value := args[1], args[2]; setting {
			case "address":
				service.Address = value
			case "proto":
				service.Proto = value
			default:
				return fmt.Errorf("Unknown service setting %s", setting)
			}

			return store.Set(service)
		},
	}
}

func serviceGetCmd(store serviceStore) *cobra.Command {
	return &cobra.Command{
		Use:   "get <name>",
		Short: "get configuration for a service",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service, err := store.Get(args[0])
			if err != nil {
				return err
			}

			// TODO(njhale): Improve formatting
			printYaml(cmd.OutOrStdout(), service)

			return nil
		},
	}
}

func serviceListCmd(store serviceStore) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "list configurations for all known services",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			services, err := store.List()
			if err != nil {
				return err
			}

			// TODO(njhale): Improve formatting
			printYaml(cmd.OutOrStdout(), services)

			return nil
		},
	}
}

func serviceRemoveCmd(store serviceStore) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "remove configuration for a service",
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
