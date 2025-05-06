// Package broker -----------------------------
// @file      : emqx.go
// @author    : jzechen
// @contact   : 593089672@qq.com
// @time      : 2025/4/17 13:49
// -------------------------------------------
package broker

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.oease.dev/goe/contracts"
	"go.uber.org/zap"
	"time"
)

type EMQX struct {
	client mqtt.Client
	logger *zap.Logger
}

func NewEMQX(c *EMQXConfig) (contracts.EMQX, error) {
	opts := mqtt.NewClientOptions().AddBroker(c.Addr).
		SetClientID(c.ID).
		SetUsername(c.Username).
		SetPassword(c.Password).
		SetDefaultPublishHandler(c.MessageHandler).
		SetOnConnectHandler(c.OnConnectHandler).
		SetConnectionLostHandler(c.ConnectionLostHandler).
		SetAutoReconnect(true).
		SetKeepAlive(60 * time.Second)
	if c.TLSConfig != nil && c.TLSConfig.TLS != nil {
		opts.SetTLSConfig(c.TLSConfig.TLS)
	}
	_client := mqtt.NewClient(opts)
	tk := _client.Connect()
	if tk.Wait() && tk.Error() != nil {
		return nil, tk.Error()
	}

	bk := &EMQX{
		client: _client,
		logger: zap.L().With(zap.String("module", "emqx")),
	}

	return bk, nil
}

func (b *EMQX) Subscribe(module string, qos byte, callback func(mqtt.Client, mqtt.Message)) error {
	if !b.client.IsConnected() {
		if token := b.client.Connect(); token.Wait() && token.Error() != nil {
			return fmt.Errorf("connect to emqx failed with %w", token.Error())
		}
	}

	token := b.client.Subscribe(module, qos, callback)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("subscribe module %s failed with %w", module, token.Error())
	}
	b.logger.Debug("Subscribed to topic", zap.String("topic", module))
	return nil
}

func (b *EMQX) Unsubscribe(module string) error {
	token := b.client.Unsubscribe(module)
	token.Wait()
	if token.Error() != nil {
		return fmt.Errorf("unSubscribe module %s failed with %w", module, token.Error())
	}
	b.logger.Debug("UnSubscribe topic", zap.String("topics", module))
	return nil
}

func (b *EMQX) Close() {
	// close connection
	b.client.Disconnect(250)
	b.logger.Debug("exit emqx broker successfully")
}

func (b *EMQX) Publish(topic string, qos byte, retained bool, payload any) error {
	token := b.client.Publish(topic, qos, retained, payload)
	token.Wait()
	if token.Error() != nil {
		b.logger.Error("MQTT publish failed", zap.Any("payload", payload), zap.Error(token.Error()))
		return token.Error()
	}
	return nil
}
