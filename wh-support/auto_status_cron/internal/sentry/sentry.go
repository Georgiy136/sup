package sentry

import (
	"context"
	"fmt"
	"strconv"

	"gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/auto_status_cron/internal/services/models"
	"gitlab.wildberries.ru/wbwh/wh-core/gocore_sentry.git"

	sentrygo "github.com/getsentry/sentry-go"
	"github.com/valyala/fasthttp"
)

const (
	tagCategoryID = "category_id"
	tagStatusID   = "status_id"
	tagTicketID   = "ticket_id"

	transactionOperationCron     = "cron"
	transactionNameProcessBatch  = "ticket.process_batch"
	transactionNameProcessTicket = "ticket.process"

	spanOperationDBQuery       = "db.query"
	spanOperationRedisQuery    = "redis.query"
	spanOperationHTTPClient    = "http.client"
	spanOperationTicketHandler = "ticket.handler"
)

func StartTicketTransaction(ctx context.Context, ticket models.TicketCommonInfo) (*sentrygo.Span, context.Context) {
	ctx = initHubContext(ctx)
	tx := gocore_sentry.StartTransaction(ctx, transactionNameProcessTicket,
		sentrygo.WithTransactionSource(sentrygo.SourceTask),
		sentrygo.WithOpName(transactionOperationCron),
	)
	tx.Status = sentrygo.SpanStatusOK
	setTicketTags(tx, ticket)

	ctx = tx.Context()

	return tx, ctx
}

func StartBatchTransaction(ctx context.Context, tickets []models.TicketCommonInfo) (*sentrygo.Span, context.Context) {
	ctx = initHubContext(ctx)
	tx := gocore_sentry.StartTransaction(ctx, transactionNameProcessBatch,
		sentrygo.WithTransactionSource(sentrygo.SourceTask),
		sentrygo.WithOpName(transactionOperationCron),
	)
	tx.Status = sentrygo.SpanStatusOK
	ctx = tx.Context()

	if len(tickets) > 0 {
		tx.SetTag(tagCategoryID, strconv.FormatInt(tickets[0].CategoryID, 10))
	}

	return tx, ctx
}

func CaptureTicketException(ctx context.Context, ticket models.TicketCommonInfo, err error) {
	hub := sentrygo.GetHubFromContext(ctx)
	hub.WithScope(func(scope *sentrygo.Scope) {
		setTicketTags(scope, ticket)
		hub.CaptureException(err)
	})
}

func MarkTransactionFailed(tx *sentrygo.Span) {
	tx.Status = sentrygo.SpanStatusInternalError
}

func StartTicketHandlerSpan(ctx context.Context) (*sentrygo.Span, context.Context) {
	span := gocore_sentry.StartSpan(ctx, spanOperationTicketHandler)
	return span, span.Context()
}

func StartHTTPClientSpan(ctx context.Context, description string) *sentrygo.Span {
	span := gocore_sentry.StartSpan(ctx, spanOperationHTTPClient)
	span.Description = description

	return span
}

func StartDBQuerySpan(ctx context.Context, procedureName string) *sentrygo.Span {
	span := gocore_sentry.StartSpan(ctx, spanOperationDBQuery)
	span.Description = procedureName

	return span
}

func StartDBRedisSSpan(ctx context.Context, operation, key string) *sentrygo.Span {
	span := gocore_sentry.StartSpan(ctx, spanOperationRedisQuery)
	span.Description = operation + " " + key

	return span
}

func FinishTicketHandlerSpan(span *sentrygo.Span, err error) {
	if err != nil {
		span.Status = sentrygo.SpanStatusInternalError
		span.Description = err.Error()
	}
	span.Finish()
}

func FinishDBQuerySpan(span *sentrygo.Span, err error) {
	if err != nil {
		span.Status = sentrygo.SpanStatusInternalError
		span.Description = fmt.Sprintf("%s: %s", span.Description, err)
	}
	span.Finish()
}

func FinishHTTPClientSpan(span *sentrygo.Span, response *fasthttp.Response, err error) {
	if err != nil {
		span.Status = sentrygo.SpanStatusInternalError
		span.Description = fmt.Sprintf("%s: %s", span.Description, err)
	} else if response != nil {
		span.Status = sentrygo.HTTPtoSpanStatus(response.StatusCode())
	}
	span.Finish()
}

func initHubContext(ctx context.Context) context.Context {
	hub := sentrygo.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentrygo.CurrentHub().Clone()
	}

	return sentrygo.SetHubOnContext(ctx, hub)
}

type tagSetter interface {
	SetTag(name, value string)
}

func setTicketTags(target tagSetter, ticket models.TicketCommonInfo) {
	target.SetTag(tagTicketID, strconv.FormatInt(ticket.TicketID, 10))
	target.SetTag(tagStatusID, ticket.StatusID)
	target.SetTag(tagCategoryID, strconv.FormatInt(ticket.CategoryID, 10))
}
