package internal

import (
	"context"
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type DialFunc func(address string) (*grpc.ClientConn, error)

type Parser interface {
	ParseFiles(filenames ...string) ([]*desc.FileDescriptor, error)
}

type ServiceDiscoveryOption func(*serviceDiscovery)

func WithBlockingDial(ctx context.Context, opts ...grpc.DialOption) ServiceDiscoveryOption {
	return func(s *serviceDiscovery) {
		s.dial = func(address string) (*grpc.ClientConn, error) {
			cc, err := grpc.DialContext(ctx, address, append(opts, grpc.FailOnNonTempDialError(true), grpc.WithBlock())...)
			if err != nil {
				return nil, err
			}

			return cc, nil
		}
	}
}

func WithParser(parser Parser) ServiceDiscoveryOption {
	return func(s *serviceDiscovery) {
		s.parser = parser
	}
}

func NewServiceDiscovery(opts ...ServiceDiscoveryOption) (*serviceDiscovery, error) {
	// Create default
	discovery := &serviceDiscovery{
		parser: protoparse.Parser{},
	}
	WithBlockingDial(context.Background(), grpc.WithInsecure())(discovery)

	// Customize
	for _, opt := range opts {
		opt(discovery)
	}

	return discovery, nil
}

type serviceDiscovery struct {
	parser Parser
	dial   DialFunc
}

func (s *serviceDiscovery) Services(server Server) ([]*desc.ServiceDescriptor, error) {
	// TODO(njhale): memoize ServiceDescriptors

	var (
		services []*desc.ServiceDescriptor
		err      error
	)

	switch proto, address := server.Proto, server.Address; {
	case proto != "":
		services, err = s.servicesFromProto(server.Proto)
		if err != nil {
			err = fmt.Errorf("proto parse error: %w", err)
		}
	case address != "":
		fmt.Printf("getting services from reflection for %s\n", server.Name)
		services, err = s.servicesFromReflection(server.Address)
		if err != nil {
			err = fmt.Errorf("server reflection error: %w", err)
		}
	default:
		err = fmt.Errorf("no server proto or address defined")
	}

	return services, err
}

func (s *serviceDiscovery) servicesFromProto(proto string) ([]*desc.ServiceDescriptor, error) {
	files, err := s.parser.ParseFiles(proto)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", proto, err)
	}

	var services []*desc.ServiceDescriptor
	resolved := map[string]struct{}{}
	for _, file := range files {
		for _, service := range file.GetServices() {
			name := service.GetFullyQualifiedName()
			if _, ok := resolved[name]; ok {
				// Already resolved, skip
				continue
			}

			services = append(services, service)
			resolved[name] = struct{}{}
		}
	}

	return services, nil
}

func (s *serviceDiscovery) servicesFromReflection(address string) ([]*desc.ServiceDescriptor, error) {
	fmt.Println("dialing...")
	cc, err := s.dial(address)
	if err != nil {
		return nil, err
	}
	defer cc.Close()
	fmt.Println("connected to server")

	ctx := context.TODO()
	client := grpcreflect.NewClient(ctx, reflectpb.NewServerReflectionClient(cc))
	defer client.Reset()

	serviceNames, err := client.ListServices()
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	var services []*desc.ServiceDescriptor
	resolved := map[string]struct{}{}
	for _, name := range serviceNames {
		if _, ok := resolved[name]; ok {
			// Already resolved, skip
			continue
		}

		service, err := client.ResolveService(name)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve descriptor for service %s: %w", name, err)
		}

		services = append(services, service)
		resolved[name] = struct{}{}
	}

	return services, nil
}
