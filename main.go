package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/agronskiy/telegram-page-checker-bot/internal/config"
	"github.com/agronskiy/telegram-page-checker-bot/internal/pipeline"

	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
)

var cfg config.Config

func doTick(
	singleUrl *config.SingleURL,
	htmlIds *config.ElementIds,
	t time.Time,
	lastReports map[uint32]time.Time,
	ctx context.Context,
) {
	log := log.WithField("name", singleUrl.Name)

	if t.UTC().Hour() < cfg.AllowedRequestsMinHour || t.UTC().Hour() >= cfg.AllowedRequestsMaxHour {
		log.WithField("user", singleUrl.Name).Println("Skipping request outside of hours")
		return
	}

	needSayNegative := false
	// We still send even a negative message as health check. Sending it in the morning.
	_, ok := lastReports[singleUrl.GetHash()]
	isWithinRange := t.UTC().Hour() >= cfg.HealthCheckMinHour && t.UTC().Hour() < cfg.HealthCheckMaxHour
	canSendHealthCheck := (!ok ||
		(t.Sub(lastReports[singleUrl.GetHash()]).Hours() >
			float64(cfg.HealthCheckMaxHour-cfg.HealthCheckMinHour)))
	if isWithinRange && canSendHealthCheck {
		needSayNegative = true
		lastReports[singleUrl.GetHash()] = t
		log.Info("Will send healthcheck regardless of status")
	}

	result := pipeline.RunWholePipeline(singleUrl, htmlIds, ctx)
	if err := sayResult(singleUrl, result, needSayNegative); err != nil {
		log.WithField("user", singleUrl.Name).Println("error in sending reply:", err)
	}
}

func runPeriodicChecks() {

	// We create browser
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	chromedp.Run(ctx) // This will explicitly allocate browser, kept alive

	var lastReports = make(map[uint32]time.Time)

	for _, singleUrl := range cfg.Urls {
		if singleUrl.Enabled {
			doTick(singleUrl, &cfg.HtmlElems, time.Now(), lastReports, ctx)
		}
	}

	minInterval := cfg.IntervalMinute
	maxInterval := cfg.IntervalMinute + 10
	nextCheck := time.NewTicker(time.Duration(
		rand.Intn(maxInterval-minInterval)+minInterval) * time.Minute)
	for {
		t := <-nextCheck.C
		for _, singleUrl := range cfg.Urls {
			if singleUrl.Enabled {
				doTick(singleUrl, &cfg.HtmlElems, t, lastReports, ctx)
			}
		}
		nextCheck.Stop()
		nextCheck = time.NewTicker(time.Duration(
			rand.Intn(maxInterval-minInterval)+minInterval) * time.Minute)
	}
}

func main() {
	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		PadLevelText:    true,
	})

	config.ReadConfig(&cfg)
	rand.Seed(time.Now().UnixNano())

	runPeriodicChecks()
}
