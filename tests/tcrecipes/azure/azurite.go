package azure

import (
	"context"
	"fmt"
	"net"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// Default ports used by Azurite
	BlobPort  = "10000/tcp"
	QueuePort = "10001/tcp"
	TablePort = "10002/tcp"
)

// AzuriteContainer represents the Azurite container type used in the module
type AzuriteContainer struct {
	testcontainers.Container
	Settings options
}

func (c *AzuriteContainer) ServiceURL(ctx context.Context, srv Service) (string, error) {
	hostname, err := c.Host(ctx)
	if err != nil {
		return "", err
	}

	var port nat.Port
	switch srv {
	case BlobService:
		port = BlobPort
	case QueueService:
		port = QueuePort
	case TableService:
		port = TablePort
	default:
		return "", fmt.Errorf("unknown service: %s", srv)
	}

	mappedPort, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s", net.JoinHostPort(hostname, mappedPort.Port())), nil
}

func (c *AzuriteContainer) MustServiceURL(ctx context.Context, srv Service) string {
	url, err := c.ServiceURL(ctx, srv)
	if err != nil {
		panic(err)
	}

	return url
}

// Run creates an instance of the Azurite container type
func RunAzurite(ctx context.Context, img string, opts ...testcontainers.ContainerCustomizer) (*AzuriteContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{BlobPort, QueuePort, TablePort},
		Env:          map[string]string{},
		Entrypoint:   []string{"azurite"},
		Cmd:          []string{},
	}

	genericContainerReq := testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	}

	// 1. Gather all config options (defaults and then apply provided options)
	settings := defaultOptions()
	for _, opt := range opts {
		if err := opt.Customize(&genericContainerReq); err != nil {
			return nil, err
		}
	}

	// 2. evaluate the enabled services to apply the right wait strategy and Cmd options
	enabledServices := settings.EnabledServices
	if len(enabledServices) > 0 {
		waitingFor := make([]wait.Strategy, 0)
		for _, srv := range enabledServices {
			switch srv {
			case BlobService:
				genericContainerReq.Cmd = append(genericContainerReq.Cmd, "--blobHost", "0.0.0.0")
				waitingFor = append(waitingFor, wait.ForLog("Blob service is successfully listening"))
			case QueueService:
				genericContainerReq.Cmd = append(genericContainerReq.Cmd, "--queueHost", "0.0.0.0")
				waitingFor = append(waitingFor, wait.ForLog("Queue service is successfully listening"))
			case TableService:
				genericContainerReq.Cmd = append(genericContainerReq.Cmd, "--tableHost", "0.0.0.0")
				waitingFor = append(waitingFor, wait.ForLog("Table service is successfully listening"))
			}
		}

		if len(waitingFor) > 0 {
			genericContainerReq.WaitingFor = wait.ForAll(waitingFor...)
		}
	}

	container, err := testcontainers.GenericContainer(ctx, genericContainerReq)
	if err != nil {
		return nil, err
	}

	return &AzuriteContainer{Container: container, Settings: settings}, nil
}
