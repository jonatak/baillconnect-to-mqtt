package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/jonatak/baillconnect-to-mqtt/internal/application"
	"github.com/jonatak/baillconnect-to-mqtt/internal/bailup"
	"github.com/jonatak/baillconnect-to-mqtt/internal/config"
	"github.com/jonatak/baillconnect-to-mqtt/internal/mqtt"
)

const startupTimeout = 30 * time.Second

func NewHVACService(ctx context.Context, cfg config.Config) (*application.HVACService, error) {
	if cfg.Baillconnect.Email == "" || cfg.Baillconnect.Password == "" || cfg.Baillconnect.Regulation == "" {
		return nil, ErrInit
	}

	gateway := bailup.NewGateway(cfg.Baillconnect.Email, cfg.Baillconnect.Password, cfg.Baillconnect.Regulation)
	ctx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()

	err := gateway.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("connect HVAC gateway: %w", err)
	}
	return application.NewHVACService(gateway), nil
}

func NewMQTTServer(
	ctx context.Context,
	system *application.HVACService,
	cfg config.Config,
) (*mqtt.Processor, error) {

	ctx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()

	state, err := system.CurrentState(ctx)
	if err != nil {
		return nil, fmt.Errorf("load initial HVAC state: %w", err)
	}

	if cfg.MQTT.Host == "" || cfg.MQTT.Username == "" || cfg.MQTT.Password == "" || cfg.MQTT.TopicPrefix == "" || cfg.MQTT.ClientID == "" || cfg.MQTT.Port <= 0 {
		return nil, ErrMqttInit
	}

	params := mqtt.HandlerParams{
		Host:     cfg.MQTT.Host,
		Username: cfg.MQTT.Username,
		Password: cfg.MQTT.Password,
		Port:     cfg.MQTT.Port,
		ClientID: cfg.MQTT.ClientID,
		Prefix:   cfg.MQTT.TopicPrefix,
	}

	handler, err := mqtt.NewMQTTHandler(params, state)

	if err != nil {
		return nil, err
	}

	return mqtt.NewProcessor(handler, system, time.Duration(cfg.PollInterval)*time.Second), nil
}
