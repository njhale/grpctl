package cmd

import (
	"fmt"

	"github.com/njhale/grpctl/internal"
	"github.com/spf13/cobra"
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
	// cmd.AddCommand(serviceSetCmd)
	// cmd.AddCommand(serviceSetCmd)
	// cmd.AddCommand(serviceSetCmd)

	return cmd
}

func serviceSetCmd(store serviceStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <name> [address|proto] <value>",
		Short: "",
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
				// TODO(njhale): Implement setting proto from file and stdin
			default:
				return fmt.Errorf("Unknown service setting %s", setting)
			}

			return store.Set(service)
		},
	}

	return cmd
}

var serviceGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(njhale): implement this!
		return nil
	},
}

var serviceListCmd = &cobra.Command{
	Use:   "list",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(njhale): implement this!
		return nil
	},
}

var serviceRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO(njhale): implement this!
		return nil
	},
}
