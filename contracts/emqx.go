// Package contracts -----------------------------
// @file      : emqx.go
// @author    : jzechen
// @contact   : 593089672@qq.com
// @time      : 2025/4/17 15:44
// -------------------------------------------
package contracts

import "github.com/eclipse/paho.mqtt.golang"

type EMQX interface {
	// Publish will publish a message with the specified QoS and content to the specified topic.
	// Returns a token to track delivery of the message to the broker.
	Publish(topic string, qos byte, retained bool, payload any) error
	// Subscribe starts a new subscription.
	// Provide a MessageHandler to be executed when a message is published on the topic provided, or nil for the default handler.
	Subscribe(module string, qos byte, callback func(mqtt.Client, mqtt.Message)) error
	// Unsubscribe will end the subscription from each of the topics provided.
	// Messages published to those topics from other clients will no longer be received.
	Unsubscribe(module string) error
	// Close will end the connection with the server,
	// but not before waiting 250 milliseconds to wait for existing work to be completed.
	Close()
}
