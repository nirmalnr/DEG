package degledgerrecorder

import (
	"context"
	"sync"

	"github.com/beckn-one/beckn-onix/pkg/log"
	"github.com/beckn-one/beckn-onix/pkg/model"
)

// DEGLedgerRecorder is a Step plugin that records trade data to the DEG Ledger
// after on_confirm calls.
type DEGLedgerRecorder struct {
	config *Config
	client *LedgerClient

	// wg tracks in-flight async requests for graceful shutdown
	wg sync.WaitGroup
}

// New creates a new DEGLedgerRecorder instance.
func New(cfg map[string]string) (*DEGLedgerRecorder, error) {
	config, err := ParseConfig(cfg)
	if err != nil {
		return nil, err
	}

	client := NewLedgerClient(config.LedgerHost, config.AsyncTimeout, config.RetryCount)

	return &DEGLedgerRecorder{
		config: config,
		client: client,
	}, nil
}

// Run implements the Step interface. It processes the request and records
// on_confirm events to the DEG Ledger asynchronously.
func (r *DEGLedgerRecorder) Run(ctx *model.StepContext) error {
	// Skip if plugin is disabled
	if !r.config.Enabled {
		log.Debug(ctx, "DEGLedgerRecorder: plugin disabled, skipping")
		return nil
	}

	// Extract the action from the request
	action := ExtractAction(ctx.Request.URL.Path, ctx.Body)

	// Only process on_confirm actions
	if action != "on_confirm" {
		log.Debugf(ctx, "DEGLedgerRecorder: action '%s' is not on_confirm, skipping", action)
		return nil
	}

	log.Infof(ctx, "DEGLedgerRecorder: processing on_confirm")

	// Parse the on_confirm payload
	payload, err := ParseOnConfirm(ctx.Body)
	if err != nil {
		// Log error but don't fail the main flow
		log.Warnf(ctx, "DEGLedgerRecorder: failed to parse on_confirm payload: %v", err)
		return nil
	}

	// Map to ledger records (one per order item)
	records := MapToLedgerRecords(payload, r.config.Role)

	if len(records) == 0 {
		log.Warnf(ctx, "DEGLedgerRecorder: no order items found in on_confirm, skipping ledger recording")
		return nil
	}

	log.Infof(ctx, "DEGLedgerRecorder: mapped %d ledger records from on_confirm (transaction_id=%s)",
		len(records), payload.Context.TransactionID)

	// Send records to ledger asynchronously (fire-and-forget)
	r.sendRecordsAsync(ctx, records, payload.Context.TransactionID)

	return nil
}

// sendRecordsAsync sends ledger records in the background without blocking the main flow.
func (r *DEGLedgerRecorder) sendRecordsAsync(parentCtx *model.StepContext, records []LedgerPutRequest, transactionID string) {
	for _, record := range records {
		r.wg.Add(1)
		go func(rec LedgerPutRequest) {
			defer r.wg.Done()

			// Create a new context with timeout for the async operation
			ctx, cancel := context.WithTimeout(context.Background(), r.config.AsyncTimeout)
			defer cancel()

			resp, err := r.client.PutRecord(ctx, rec)
			if err != nil {
				log.Errorf(parentCtx, err,
					"DEGLedgerRecorder: failed to record to ledger (transaction_id=%s, order_item_id=%s): %v",
					rec.TransactionID, rec.OrderItemID, err)
				return
			}

			log.Infof(parentCtx,
				"DEGLedgerRecorder: successfully recorded to ledger (transaction_id=%s, order_item_id=%s, record_id=%s)",
				rec.TransactionID, rec.OrderItemID, resp.RecordID)
		}(record)
	}
}

// Close gracefully shuts down the recorder, waiting for in-flight requests.
func (r *DEGLedgerRecorder) Close() {
	// Wait for all in-flight requests to complete
	r.wg.Wait()

	// Close the HTTP client
	if r.client != nil {
		r.client.Close()
	}
}
