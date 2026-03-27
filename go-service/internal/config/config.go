package config

import (
	"os"
	"strconv"
)

type Config struct {
	KafkaBrokers     string
	KafkaInputTopic  string
	KafkaOutputTopic string
	KafkaGroupID     string
	MLServiceURL     string
	WorkerCount      int
	QueueSize        int
	DBDSN            string
}

func Load() Config {
	return Config{
		KafkaBrokers:     getEnv("KAFKA_BROKERS", "localhost:9092"),
		KafkaInputTopic:  getEnv("KAFKA_INPUT_TOPIC", "subsidy-tasks"),
		KafkaOutputTopic: getEnv("KAFKA_OUTPUT_TOPIC", "subsidy-results"),
		KafkaGroupID:     getEnv("KAFKA_GROUP_ID", "go-subsidy-group"),
		MLServiceURL:     getEnv("ML_SERVICE_URL", "http://localhost:8000"),
		WorkerCount:      getEnvInt("WORKER_COUNT", 5),
		QueueSize:        getEnvInt("QUEUE_SIZE", 50),
		DBDSN:            getEnv("DB_DSN", ""),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
