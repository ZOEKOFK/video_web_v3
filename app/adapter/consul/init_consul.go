package consul

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
)

const (
	UserServiceName        = "user-service"
	VideoServiceName       = "video-service"
	SocialServiceName      = "social-service"
	InteractionServiceName = "interaction-service"

	ConsulDefaultAddr = "localhost:8500"
	ServiceHost       = "127.0.0.1"

	UserPort        = 50051
	VideoPort       = 50052
	SocialPort      = 50053
	InteractionPort = 50054
)

type ConsulClient struct {
	client           *api.Client
	registeredSvcIDs []string
	mu               sync.Mutex
}

func NewConsulClient(address string) (*ConsulClient, error) {
	config := api.DefaultConfig()
	config.Address = address
	client, err := api.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("创建 Consul 客户端失败: %w", err)
	}
	return &ConsulClient{client: client}, nil
}

func (c *ConsulClient) RegisterService(serviceName, host string, port int, tags []string) (string, error) {
	serviceID := fmt.Sprintf("%s-%s", serviceName, uuid.New().String()[:8])

	registration := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Tags:    tags,
		Port:    port,
		Address: host,
		Check: &api.AgentServiceCheck{
			GRPC:     fmt.Sprintf("%s:%d", host, port),
			Interval: "10s",
			Timeout:  "5s",
		},
	}

	if err := c.client.Agent().ServiceRegister(registration); err != nil {
		return "", fmt.Errorf("注册服务 [%s] 失败: %w", serviceName, err)
	}

	c.mu.Lock()
	c.registeredSvcIDs = append(c.registeredSvcIDs, serviceID)
	c.mu.Unlock()

	log.Printf("✅ 服务 [%s] 已注册到 Consul (ID: %s, %s:%d)", serviceName, serviceID, host, port)
	return serviceID, nil
}

func (c *ConsulClient) SafeRegister(serviceName string, host string, port int, tags []string) {
	if c == nil || c.client == nil {
		return
	}
	if _, err := c.RegisterService(serviceName, host, port, tags); err != nil {
		log.Printf("⚠️ 注册服务 [%s] 到 Consul 失败: %v", serviceName, err)
	}
}

func (c *ConsulClient) DeregisterService(serviceID string) error {
	if err := c.client.Agent().ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("注销服务 [%s] 失败: %w", serviceID, err)
	}
	log.Printf("✅ 服务 [%s] 已从 Consul 注销", serviceID)
	return nil
}

func (c *ConsulClient) Ping() error {
	_, err := c.client.Agent().Self()
	return err
}

func (c *ConsulClient) DeregisterAll() {
	if c == nil || c.client == nil {
		return
	}
	c.mu.Lock()
	ids := make([]string, len(c.registeredSvcIDs))
	copy(ids, c.registeredSvcIDs)
	c.mu.Unlock()

	for _, id := range ids {
		if err := c.DeregisterService(id); err != nil {
			log.Printf("注销服务 %s 失败: %v", id, err)
		}
	}
}

func (c *ConsulClient) SetupGracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("\n收到退出信号，正在注销所有服务...")
		c.DeregisterAll()
		os.Exit(0)
	}()
}
