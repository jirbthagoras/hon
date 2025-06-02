package main

import (
	"log/slog"
	"sync"

	"github.com/jirbthagoras/hon/shared"
)

func main() {
	AMQP := shared.NewAMQPConnection()
	mailer := NewMailer()
	sql := shared.GetConnection()
	var wg sync.WaitGroup

	service := NewConsumerService(mailer, sql)
	handler := NewConsumerHandler(AMQP, service, &wg)

	handlers := handler.BundleConsumer()

	wg.Add(len(handlers))

	for _, handler := range handlers {
		go func(h func()) {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("Recovered from panic in consumer", "error", r)
				}
			}()
			h()
		}(handler)
	}

	wg.Wait()
}
