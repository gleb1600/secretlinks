package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

type StatsItem struct {
	LinkKey    string      `json:"linkkey"`
	CreateTime time.Time   `json:"createtime"`
	VisitTime  []time.Time `json:"visittime"`
}

type KafkaStatsItem struct {
	LinkKey string    `json:"linkkey"`
	NowTime time.Time `json:"nowtime"`
}

type StatsStorage struct {
	mu    sync.RWMutex
	items map[string]StatsItem
}

func NewStatsStorage() *StatsStorage {
	return &StatsStorage{
		items: make(map[string]StatsItem),
	}
}

func (s *StatsStorage) AddNewItem(linkKey string, createTime time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[linkKey] = StatsItem{
		LinkKey:    linkKey,
		CreateTime: createTime,
		VisitTime:  []time.Time{},
	}
}

func (s *StatsStorage) AppendVisitTime(linkKey string, visitTime time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if item, exists := s.items[linkKey]; exists {
		item.VisitTime = append(item.VisitTime, visitTime)
		s.items[linkKey] = item
	} else {
		s.items[linkKey] = StatsItem{
			LinkKey:    linkKey,
			CreateTime: visitTime,
			VisitTime:  []time.Time{},
		}
	}
}

func (s *StatsStorage) ShowStorage() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fmt.Println("Stats Storage:")
	for _, v := range s.items {
		if len(v.VisitTime) == 0 {
			fmt.Printf("ID: %s | Created: %s | Visits: %v\n",
				v.LinkKey,
				v.CreateTime.Format("2006-01-02 15:04:05"),
				len(v.VisitTime))
		} else {
			fmt.Printf("ID: %s | Created: %s | Visits: %v, last visit: %v\n",
				v.LinkKey,
				v.CreateTime.Format("2006-01-02 15:04:05"),
				len(v.VisitTime),
				v.VisitTime[len(v.VisitTime)-1].Format("2006-01-02 15:04:05"))
		}
	}
}

func (s *StatsStorage) ShowItem(linkKey string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if item, exists := s.items[linkKey]; exists {
		fmt.Printf("ID: %s | Created: %s | Visits: %d\n",
			item.LinkKey,
			item.CreateTime.Format("2006-01-02 15:04:05"),
			len(item.VisitTime))
	}
}

type KafkaReaderConfig struct {
	Brokers []string
	GroupID string
	Topic   string
	Storage *StatsStorage
}

func RunKafkaReader(ctx context.Context, wg *sync.WaitGroup, config KafkaReaderConfig) {
	defer wg.Done()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     config.Brokers,
		Topic:       config.Topic,
		GroupID:     config.GroupID,
		StartOffset: kafka.LastOffset,
	})
	defer reader.Close()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				fmt.Printf("Consumer error (topic %s): %v\n", config.Topic, err)
				continue
			}

			var stat KafkaStatsItem
			if err := json.Unmarshal(msg.Value, &stat); err != nil {
				fmt.Printf("JSON decode error (topic %s): %v\n", config.Topic, err)
				reader.CommitMessages(ctx, msg)
				continue
			}

			switch config.Topic {
			case "newlinks":
				config.Storage.AddNewItem(stat.LinkKey, stat.NowTime)
			case "updatelinks":
				config.Storage.AppendVisitTime(stat.LinkKey, stat.NowTime)
			}

			if err := reader.CommitMessages(ctx, msg); err != nil {
				fmt.Printf("Commit error (topic %s): %v\n", config.Topic, err)
			} else {
				config.Storage.ShowItem(stat.LinkKey)
			}
		}
	}
}

func main() {
	fmt.Println("Stats module activated")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaConfig := struct {
		Brokers []string
		GroupID string
	}{
		Brokers: []string{"localhost:9092"},
		GroupID: "links-group",
	}

	storage := NewStatsStorage()
	var wg sync.WaitGroup

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	topics := []string{"newlinks", "updatelinks"}
	wg.Add(len(topics))

	for _, topic := range topics {
		go RunKafkaReader(ctx, &wg, KafkaReaderConfig{
			Brokers: kafkaConfig.Brokers,
			GroupID: kafkaConfig.GroupID,
			Topic:   topic,
			Storage: storage,
		})
	}

	<-sigCh
	cancel()
	wg.Wait()

	fmt.Println("Final statistics:")
	storage.ShowStorage()
}
