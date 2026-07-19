package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Handler processes a Kafka record.
//
// Returning nil means the record has been successfully handled
// and may be committed.
//
// Returning an error means the record was not successfully handled.
// The handler is responsible for any retry or dead-letter strategy.
type Handler func(context.Context, *kgo.Record) error

// Consumer consumes records from Kafka and delegates message processing to a
// user-provided Handler.
//
// The consumer is transport-focused and is responsible only for polling
// records, invoking the handler, and committing successfully processed
// records. It does not perform payload deserialization, retries, or
// dead-letter handling.
//
// Fatal transport errors, such as polling or commit failures, stop the
// consumer and are reported through the configured error channel. Message
// processing errors are considered non-fatal and do not stop the consumer.
type Consumer struct {
	client  *kgo.Client
	errChan chan error // Buffered channel used to report fatal consumer errors.
}

func NewConsumer(client *kgo.Client, errChan chan error) *Consumer {
	return &Consumer{
		client:  client,
		errChan: errChan,
	}
}

// Consume starts the consumer loop and blocks until the provided context is
// canceled or a fatal transport error occurs.
//
// Records are fetched in batches. Each record is passed to the provided
// Handler for processing. If the handler returns nil, the record is considered
// successfully processed and becomes eligible for committing. If the handler
// returns an error, the record is skipped and is not committed.
//
// The handler is responsible for deserializing the record payload and for
// implementing any retry or dead-letter strategy.
//
// Successfully processed records are committed in batches after all records in
// the fetched batch have been handled. Fatal polling or commit errors stop the
// consumer and are reported through the configured error channel.
func (c *Consumer) Consume(ctx context.Context, handler Handler) {
	for {
		fetches := c.client.PollFetches(ctx)

		if err := fetches.Err0(); err != nil {
			// avoid reporting normal context cancellation.
			if ctx.Err() != nil {
				return
			}

			c.errChan <- err
			return
		}

		iter := fetches.RecordIter()

		var records []*kgo.Record

		for !iter.Done() {
			record := iter.Next()

			if err := handler(ctx, record); err != nil {
				// TODO later after implementing logging log the error.
				// handler is responsible for retrying the record or sending dead-letter records.
				continue
			}

			records = append(records, record)
		}

		// do not mark empty batches for commiting.
		if len(records) == 0 {
			continue
		}

		c.client.MarkCommitRecords(records...)
	}
}
