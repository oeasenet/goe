// Package broker -----------------------------
// @file      : config.go
// @author    : jzechen
// @contact   : 593089672@qq.com
// @time      : 2025/4/17 16:02
// -------------------------------------------
package broker

import (
	"crypto/tls"
	"crypto/x509"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"
	"os"
	"time"
)

// EMQXConfig emqx broker configuration struct
// Refer to the official EMQX documentation at https://docs.emqx.com/zh/emqx/latest/
// and go client documentation at https://www.emqx.com/en/blog/how-to-use-mqtt-in-golang
type EMQXConfig struct {
	// ID client id
	ID string `json:"id"`
	// Addr emqx server addr, eg: tcp://127.0.0.1:1883
	Addr string `json:"addr"`
	// Username
	Username string `json:"username"`
	// Password
	Password string `json:"password"`
	// MessageHandler is a callback type which can be set to be executed upon the arrival of messages published to topics to which the client is subscribed.
	MessageHandler mqtt.MessageHandler `json:"-"`
	// ConnectHandler OnConnectHandler is a callback that is called when the client state changes from unconnected/disconnected to connected.
	// Both at initial connection and on reconnection.
	OnConnectHandler mqtt.OnConnectHandler `json:"-"`
	// ConnectionLostHandler is a callback that is called when the client loses its connection to the broker.
	ConnectionLostHandler mqtt.ConnectionLostHandler `json:"-"`
	TLSConfig             *TLSConfig                 `json:"TLSConfig"`
}

func (cfg *EMQXConfig) Complete() {
	if cfg == nil {
		return
	}
	var loggerModule = zap.String("module", "emqx")

	var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
		zap.L().With(loggerModule).Debug("Received message from topic", zap.String("payload", string(msg.Payload())), zap.String("topic", msg.Topic()))
	}

	var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
		zap.L().With(loggerModule).Debug("Connected to emqx broker successfully")
	}

	var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
		zap.L().With(loggerModule).Debug("broker connect lost", zap.Error(err))
		// try to reconnect
		for !client.IsConnected() {
			zap.L().With(loggerModule).Debug("Reconnecting...")
			token := client.Connect()
			if token.Wait() && token.Error() != nil {
				zap.L().With(loggerModule).Debug("Reconnect failed. Retrying in 5 seconds...", zap.Error(token.Error()))
				time.Sleep(5 * time.Second)
			} else {
				zap.L().With(loggerModule).Debug("Reconnected successfully.")
				break
			}
		}
	}

	cfg.MessageHandler = messagePubHandler
	cfg.OnConnectHandler = connectHandler
	cfg.ConnectionLostHandler = connectLostHandler

	if cfg.TLSConfig != nil && cfg.TLSConfig.Enable {
		cfg.TLSConfig.TLS = cfg.TLSConfig.NewTlsConfig()
	}
}

type TLSConfig struct {
	Enable   bool        `yaml:"enable"`
	CA       string      `yaml:"ca"`
	CertFile string      `yaml:"certFile"`
	KeyFile  string      `yaml:"keyFile"`
	TLS      *tls.Config `yaml:"-"`
}

func (c *TLSConfig) NewTlsConfig() *tls.Config {
	if c == nil {
		return &tls.Config{}
	}

	ca, err := os.ReadFile(c.CA)
	if err != nil {
		zap.L().Fatal("Failed to read CA certificate", zap.Error(err))
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca)

	// Import client certificate/key pair
	clientKeyPair, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		zap.L().Fatal("Failed to LoadX509KeyPair", zap.Error(err))
	}
	return &tls.Config{
		RootCAs:            pool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{clientKeyPair},
	}
}
