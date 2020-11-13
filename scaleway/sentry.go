package scaleway

import (
	"fmt"
	"log"

	"github.com/getsentry/raven-go"
)

const (
	dsn = "https://02b954a634fc439dbb1ac16df9cd3da5@sentry.internal.scaleway.com/223"
)

// RecoverPanicAndSendReport will recover error if any, log them, and send them to sentry.
// It must be called with the defer built-in.
func RecoverPanicAndSendReport() {
	e := recover()
	if e == nil {
		return
	}

	sentryClient, err := newSentryClient()
	if err != nil {
		log.Printf("cannot create sentry client: %s", err)
	}

	err, isError := e.(error)
	if isError {
		logAndSentry(sentryClient, err)
	} else {
		logAndSentry(sentryClient, fmt.Errorf("unknownw error: %v", e))
	}

	panic(e)
}

func logAndSentry(sentryClient *raven.Client, err error) {
	log.Printf("%s", err)
	if sentryClient != nil {
		log.Printf("sending sentry report: %s", sentryClient.CaptureErrorAndWait(err, nil))
	}
}

// newSentryClient create a sentry client with build info tags.
func newSentryClient() (*raven.Client, error) {
	client, err := raven.New(dsn)
	if err != nil {
		return nil, err
	}

	tagsContext := map[string]string{
		"version": version,
	}

	client.SetTagsContext(tagsContext)

	return client, nil
}
