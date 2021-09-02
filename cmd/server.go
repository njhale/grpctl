package cmd

import (
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/njhale/grpctl/internal"
	"github.com/spf13/cobra"
)

type serviceDiscovery interface {
	Services(server internal.Server) ([]*desc.ServiceDescriptor, error)
}

func serverCmd(server internal.Server, discovery serviceDiscovery) (*cobra.Command, error) {
	// TODO(njhale): Finish me.
	root := &cobra.Command{
		Use:   server.Name,
		Short: fmt.Sprintf("call %s at %s", server.Name, server.Address),
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	fmt.Println("getting services")
	services, err := discovery.Services(server)
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

	fmt.Printf("found %d services\n", len(services))

	methods := map[string][]*desc.MethodDescriptor{}
	serviceCmds := map[string]*cobra.Command{}
	for _, service := range services {
		cmd, err := serviceCmd(service)
		if err != nil {
			return nil, err
		}
		serviceCmds[service.GetName()] = cmd

		sm := service.GetMethods()
		for _, m := range sm {
			name := m.GetName()
			methods[name] = append(methods[name], m)
		}
	}

	for name, dupes := range methods {
		unique := len(dupes) == 1
		for _, method := range dupes {
			serviceName := method.GetService().GetName()
			serviceCmd, ok := serviceCmds[serviceName]
			if !ok {
				return nil, fmt.Errorf("method %s missing parent service %s", name, serviceName)
			}

			cmd, err := methodCmd(method)
			if err != nil {
				return nil, err
			}

			serviceCmd.AddCommand(cmd)
			serviceCmd.Hidden = serviceCmd.Hidden && unique // Unhide a service anytime there's a collision with the methods of other services
			root.AddCommand(serviceCmd)

			if unique {
				// Add a top level command for the service
				root.AddCommand(cmd)
			}
		}
	}

	return root, nil
}

func serviceCmd(service *desc.ServiceDescriptor) (*cobra.Command, error) {
	// TODO(njhale): Finish me.
	name := internal.Commandize(service.GetName())
	root := &cobra.Command{
		Use:    name,
		Hidden: true, // Service commands are hidden by default
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return root, nil
}

func methodCmd(method *desc.MethodDescriptor) (*cobra.Command, error) {
	// TODO(njhale): Finish me.
	name := internal.Commandize(method.GetName())
	cmd := &cobra.Command{
		Use: name,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd, nil
}
