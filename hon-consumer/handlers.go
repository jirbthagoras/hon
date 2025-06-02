package main

import (
	"context"
	"log/slog"
	"sync"

	"github.com/jirbthagoras/hon/shared"
)

type ConsumerHandler struct {
	AMQP    *shared.AMQP
	Service *ConsumerService
	WG      *sync.WaitGroup
}

func NewConsumerHandler(amqp *shared.AMQP, service *ConsumerService, wg *sync.WaitGroup) *ConsumerHandler {
	return &ConsumerHandler{
		AMQP:    amqp,
		Service: service,
		WG:      wg,
	}
}

func (h *ConsumerHandler) BundleConsumer() []func() {
	return []func(){h.handleGoal, h.handleDeadline}
}

func (h *ConsumerHandler) handleGoal() {
	// Generate Agent
	agent, err := shared.NewAgent(h.AMQP, context.Background())
	if err != nil {
		slog.Error("Error when creating agent")
		panic(err)
	}

	// Generate Consumer
	goalConsumer, err := agent.NewConsumer("goal_queue", "goal-consumer")
	if err != nil {
		slog.Error("Error when creating consumer")
		panic(err)
	}

	// Make the consumer listens
	for message := range goalConsumer {
		err = h.Service.SendGoalEmail(&message)
		if err != nil {
			slog.Error(err.Error())
		}
	}

	defer h.WG.Done()
}

func (h *ConsumerHandler) handleDeadline() {
	// Generate Agent
	agent, err := shared.NewAgent(h.AMQP, context.Background())
	if err != nil {
		slog.Error("Error when creating agent: Deadline")
		panic(err)
	}

	// Generate Consumer
	goalConsumer, err := agent.NewConsumer("deadline_queue", "deadline-consumer")
	if err != nil {
		slog.Error("Error when creating consumer: Deadline")
		panic(err)
	}

	// Make the consumer listens
	for message := range goalConsumer {
		err = h.Service.SendDeadlineEmail(&message)
		if err != nil {
			slog.Error(err.Error())
		}
	}

	defer h.WG.Done()
}
