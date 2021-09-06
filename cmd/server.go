package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"github.com/njhale/grpctl/internal"
	"github.com/spf13/cobra"
)

type serviceDiscovery interface {
	Services(server *internal.Server) ([]*desc.ServiceDescriptor, error)
	Stub(server *internal.Server) (*grpcdynamic.Stub, error)
	MessageFactory(server *internal.Server) (*dynamic.MessageFactory, error)
}

func serverCmd(discovery serviceDiscovery, server *internal.Server) (*cobra.Command, error) {
	// TODO(njhale): Finish me.
	root := &cobra.Command{
		Use:   server.Name,
		Short: fmt.Sprintf("call %s at %s", server.Name, server.Address),
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	if discovery == nil {
		return root, nil
	}

	services, err := discovery.Services(server)
	if err != nil {
		return nil, fmt.Errorf("failed to discover services: %w", err)
	}

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

	stub, err := discovery.Stub(server)
	if err != nil {
		return nil, err
	}
	messageFactory, err := discovery.MessageFactory(server)
	if err != nil {
		return nil, err
	}

	for name, dupes := range methods {
		unique := len(dupes) == 1
		for _, method := range dupes {
			serviceName := method.GetService().GetName()
			serviceCmd, ok := serviceCmds[serviceName]
			if !ok {
				return nil, fmt.Errorf("method %s missing parent service %s", name, serviceName)
			}

			cmd, err := methodCmd(stub, messageFactory, method)
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
		Short:  service.GetSourceInfo().GetLeadingComments(),
		Hidden: true, // Service commands are hidden by default
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return root, nil
}

func methodCmd(stub *grpcdynamic.Stub, messageFactory *dynamic.MessageFactory, method *desc.MethodDescriptor) (*cobra.Command, error) {
	// TODO(njhale): Finish me.
	input := method.GetInputType()
	inFields := input.GetFields()

	name := internal.Commandize(method.GetName())
	use := []string{name}
	for _, field := range inFields {
		use = append(use, internal.Commandize(field.GetName()))
	}

	cmd := &cobra.Command{
		Use:   strings.Join(use, " "),
		Short: method.GetSourceInfo().GetLeadingComments(),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(njhale): support input types other than json
			in := cmd.InOrStdin()
			if len(args) > 0 {
				// Assume we've been given JSON literals as arugments
				in = strings.NewReader(strings.Join(args, " "))
			}

			decoder := json.NewDecoder(in)
			unmarshaler := jsonpb.Unmarshaler{}
			request := messageFactory.NewMessage(input)
			if len(inFields) > 0 {
				if err := unmarshaler.UnmarshalNext(decoder, request); err != nil {
					return err
				}
			}

			marshaler := jsonpb.Marshaler{}
			out := cmd.OutOrStdout()
			ctx := cmd.Context()
			switch {
			case method.IsClientStreaming() && method.IsServerStreaming(): // Bidirectional streaming
				cmd.Println("bidirectional streaming method")
			case method.IsClientStreaming(): // Client streaming
				cmd.Println("client streaming method")
			case method.IsServerStreaming(): // Server streaming
				cmd.Println("server streaming method")
			default: // Unary
				response, err := stub.InvokeRpc(ctx, method, request)
				if err == nil {
					err = marshaler.Marshal(out, response)
				}

				if err != nil {
					return fmt.Errorf("unary rpc failed: %w", err)
				}
				cmd.Println()
			}

			return nil
		},
	}

	return cmd, nil
}
