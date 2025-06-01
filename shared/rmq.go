package shared

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/rabbitmq/amqp091-go"
)

// Make connection as a struct cuz y not?
type AMQP struct {
	Connection *amqp091.Connection
}

func NewAMQPConnection() *AMQP {
	// init config, this fckin boiletplate pmo dude. Gonna find a way to refactor ts
	var config = NewConfig()
	// calls all the necessary vars
	username := config.GetString("RMQ_USERNAME")
	password := config.GetString("RMQ_PASSWORD")
	host := config.GetString("RMQ_HOST")
	port := config.GetString("RMQ_PORT")

	// craft a conn link
	connectionLink := fmt.Sprintf("amqp://%s:%s@%s:%s/", username, password, host, port)

	// Dial it with the conn link
	connection, err := amqp091.Dial(connectionLink)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	// returns a conn, don't forget to close with defer Connection.Close()
	return &AMQP{Connection: connection}
}

// Agent is a representation of channel, except this struct can reduce boilerplate code for Publishing and Consuming message
// Why not separate Producer and Consumer? Like one struct for consumer and one struct for producer?
// Remember guys, not everything needs a Struct.
// Furthermore, what we REALLY need to Publish and Consume is literally the same:
// Channel and Context that's IT!
// Wrap it on a struct, make a func to pub sub, and BOOYAH!
type Agent struct {
	Channel *amqp091.Channel
	Context context.Context
}

func NewAgent(AMQP *AMQP, context context.Context) (*Agent, error) {
	// cast a channel
	channel, err := AMQP.Connection.Channel()
	if err != nil {
		return nil, nil
	}
	// return an agent, embedding the newly created channel
	return &Agent{
		Channel: channel,
		Context: context,
	}, nil
}

// One agent can publish more than one message. So to optimize memory usage, just use the same Agent over and over
func (a *Agent) Publish(message amqp091.Publishing, exchange string, routingKey string) error {
	// publish with context
	err := a.Channel.PublishWithContext(
		a.Context,
		exchange,
		routingKey,
		false, false,
		message)
	if err != nil {
		slog.Error("Failed To Publish message")
		return err
	}
	return nil
}

// If you want to make more than one customer, please make another agent.
// Cuz the RMQ client will refuse the conn if there is more than one consumer in a channel.
func (a *Agent) NewConsumer(queue string, consumerName string) (<-chan amqp091.Delivery, error) {
	// Creates a new consumer
	consumer, err := a.Channel.ConsumeWithContext(
		a.Context, queue, consumerName,
		true, false, false, false, nil)

	if err != nil {
		return nil, err
	}

	return consumer, nil
}
