package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaStatsItem struct {
	LinkKey string    `json:"linkkey"`
	NowTime time.Time `json:"nowtime"`
}

func SendStats(resultKey, topic string) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   topic,
	})
	defer writer.Close()

	statEvent := KafkaStatsItem{
		LinkKey: resultKey,
		NowTime: time.Now(),
	}

	jsonOrder, _ := json.Marshal(statEvent)
	message := kafka.Message{
		Key:   []byte(statEvent.LinkKey),
		Value: jsonOrder,
	}
	err := writer.WriteMessages(context.Background(), message)
	if err != nil {
		panic(err)
	}

}
