// Command enqueue-job publishes a test job to the platform RabbitMQ work queue.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ting-boundless/boundless/pkg/config"
	"github.com/ting-boundless/boundless/pkg/mq"
)

func main() {
	config.LoadEnvFile()

	jobType := flag.String("type", "ping", "job type")
	jobID := flag.String("id", "", "job id (random if empty)")
	flag.Parse()

	conn, err := mq.OpenFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer conn.Close()

	pub, err := mq.NewPublisher(conn, mq.DefaultWorkQueue())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer pub.Close()

	id := *jobID
	if id == "" {
		id, err = randomID()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	job := mq.Job{
		ID:   id,
		Type: *jobType,
		Time: time.Now().UTC(),
	}
	if err := pub.Publish(context.Background(), job); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	q := mq.DefaultWorkQueue()
	fmt.Printf("enqueued %s id=%s exchange=%s routing_key=%s queue=%s\n",
		job.Type, job.ID, q.Exchange, q.RoutingKey, q.Queue)
}

func randomID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}
