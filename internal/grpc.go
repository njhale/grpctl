package internal

import (
	"context"
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
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
		parser:           protoparse.Parser{},
		services:         map[string][]*desc.ServiceDescriptor{},
		messageFactories: map[string]*dynamic.MessageFactory{},
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

	services         map[string][]*desc.ServiceDescriptor
	messageFactories map[string]*dynamic.MessageFactory
}

func (s *serviceDiscovery) Stub(server *Server) (*grpcdynamic.Stub, error) {
	address := server.Address
	if address == "" {
		return nil, fmt.Errorf("server has no address")
	}

	messageFactory, err := s.MessageFactory(server)
	if err != nil {
		return nil, fmt.Errorf("error getting message factory: %w", err)
	}

	stub := grpcdynamic.NewStubWithMessageFactory(
		&lazyChannel{
			dial:    s.dial,
			address: address,
		},
		messageFactory,
	)

	return &stub, nil
}

func (s *serviceDiscovery) MessageFactory(server *Server) (*dynamic.MessageFactory, error) {
	// If found, return memoized MessageFactory
	factory, ok := s.messageFactories[server.Name]
	if ok {
		return factory, nil
	}

	services, err := s.Services(server)
	if err != nil {
		return nil, err
	}

	if len(services) < 1 {
		return nil, fmt.Errorf("server has no services")
	}

	extensionRegistry := dynamic.NewExtensionRegistryWithDefaults()
	extensionRegistry.AddExtensionsFromFileRecursively(services[0].GetFile())

	factory = dynamic.NewMessageFactoryWithExtensionRegistry(extensionRegistry)

	// Memoize MessageFactory
	s.messageFactories[server.Name] = factory

	return factory, nil
}

func (s *serviceDiscovery) Services(server *Server) ([]*desc.ServiceDescriptor, error) {
	// If found, return memoized services
	services, ok := s.services[server.Name]
	if ok {
		return services, nil
	}

	var err error
	switch proto, address := server.Proto, server.Address; {
	case proto != "":
		services, err = s.servicesFromProto(server.Proto)
		if err == nil {
			break
		}

		err = fmt.Errorf("proto parse error: %w", err)
		fallthrough // Fall back on server reflection
	case address != "":
		services, err = s.servicesFromReflection(server.Address)
		if err != nil {
			err = fmt.Errorf("server reflection error: %w", err)
		}
	default:
		err = fmt.Errorf("no server proto or address defined")
	}

	// Memoize services
	s.services[server.Name] = services

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
	cc, err := s.dial(address)
	if err != nil {
		return nil, err
	}
	defer cc.Close()

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

type lazyChannel struct {
	address string
	dial    DialFunc
}

func (l *lazyChannel) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	cc, err := l.dial(l.address)
	if err != nil {
		return err
	}
	defer cc.Close()

	return cc.Invoke(ctx, method, args, reply, opts...)
}

func (l *lazyChannel) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	cc, err := l.dial(l.address)
	if err != nil {
		return nil, err
	}

	stream, err := cc.NewStream(ctx, desc, method, opts...)
	if err != nil {
		defer cc.Close()
		return nil, err
	}

	go func() {
		// Be sure to close the connection with the given context
		defer cc.Close()
		select {
		case <-ctx.Done():
		}
	}()

	return stream, nil

}
