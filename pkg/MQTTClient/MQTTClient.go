package MQTTClient

import (
	"crypto/rand"
	"fmt"
	"os"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Config struct {
	MessageCount *int
	MessageSize  *int
	Interval     *int
	TargetTopic  *string
	Username     *string
	Password     *string
	Host         *string
	Port         *int
}

type Client struct {
	ID             int
	Config         Config
	Connection     mqtt.Client
	Updates        chan int
	ConnectionDone chan struct{}
}

func (c *Client) Connect() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", *c.Config.Host, *c.Config.Port))
	opts.SetClientID(fmt.Sprintf("mqtt-load-generator-%d", c.ID))
	opts.SetUsername(*c.Config.Username)
	opts.SetPassword(*c.Config.Password)
	// We use a closure so we can have access to the scope if required
	opts.OnConnect = func(client mqtt.Client) {
		c.ConnectionDone <- struct{}{}
	}
	opts.OnConnectionLost = func(client mqtt.Client, err error) {
		optionsReader := client.OptionsReader()
		fmt.Printf("Connection lost for client '%s' message: %v\n", optionsReader.ClientID(), err.Error())
	}

	mqttClient := mqtt.NewClient(opts)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Error establishing MQTT connection:", token.Error().Error())
		os.Exit(1)
	}

	c.Connection = mqttClient
}

func (c Client) Start(wg *sync.WaitGroup) {
	defer wg.Done()
	payload := make([]byte, *c.Config.MessageSize)
	rand.Read(payload)

	for i := 0; i < *c.Config.MessageCount; i++ {
		token := c.Connection.Publish(*c.Config.TargetTopic, 1, false, payload)
		token.Wait()
		time.Sleep(time.Duration(*c.Config.Interval) * time.Millisecond)
		c.Updates <- 1
	}
	c.Connection.Disconnect(1)
}
