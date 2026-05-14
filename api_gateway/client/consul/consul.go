package consul

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

const (
	UserServiceName        = "user-service"
	VideoServiceName       = "video-service"
	SocialServiceName      = "social-service"
	InteractionServiceName = "interaction-service"

	ConsulDefaultAddr = "localhost:8500"

	DefaultUserAddr        = "localhost:50051"
	DefaultVideoAddr       = "localhost:50052"
	DefaultSocialAddr      = "localhost:50053"
	DefaultInteractionAddr = "localhost:50054"
)

type DiscoverClient struct {
	client *api.Client
}

func NewDiscoverClient(address string) (*DiscoverClient, error) {
	config := api.DefaultConfig()
	config.Address = address
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建 Consul 发现客户端失败: %w", err)
	}
	return &DiscoverClient{client: client}, nil
}

func (d *DiscoverClient) Ping() error {
	_, err := d.client.Agent().Self()
	return err
}

func (d *DiscoverClient) DiscoverService(serviceName string) ([]string, error) {
	services, _, err := d.client.Health().Service(serviceName, "", true, nil)
	if err != nil {
		return nil, fmt.Errorf("发现服务 [%s] 失败: %w", serviceName, err)
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("未发现健康的服务实例: %s", serviceName)
	}

	var addresses []string
	for _, entry := range services {
		addr := fmt.Sprintf("%s:%d", entry.Service.Address, entry.Service.Port)
		addresses = append(addresses, addr)
	}
	return addresses, nil
}

func (d *DiscoverClient) DiscoverOneService(serviceName string) (string, error) {
	addrs, err := d.DiscoverService(serviceName)
	if err != nil {
		return "", err
	}
	return addrs[0], nil
}

func DefaultAddr(serviceName string) string {
	switch serviceName {
	case UserServiceName:
		return DefaultUserAddr
	case VideoServiceName:
		return DefaultVideoAddr
	case SocialServiceName:
		return DefaultSocialAddr
	case InteractionServiceName:
		return DefaultInteractionAddr
	default:
		return "localhost:50051"
	}
}
