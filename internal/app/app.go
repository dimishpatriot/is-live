package app

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type App struct {
	ctx           context.Context
	logger        log.Logger
	filename      string
	freq          time.Duration
	maxRequestsCh chan MaxRequests
	sitesList     []string
	statusListCh  chan Status
	tickerCh      *time.Ticker
}

type Status struct {
	counter Counter
	url     string
	time    time.Time
	status  int
	err     error
}

type MaxRequests struct{}

type Counter struct {
	round int
	idx   int
}

func New(ctx context.Context, filename string, freq time.Duration, maxReq int) *App {
	return &App{
		ctx:           ctx,
		logger:        *log.New(os.Stdout, "INFO: ", log.Lshortfile),
		filename:      filename,
		freq:          freq,
		maxRequestsCh: make(chan MaxRequests, maxReq),
		statusListCh:  make(chan Status),
	}
}

func (app *App) Run() error {
	sitesList, err := app.getSitesList()
	if err != nil {
		return err
	}
	app.sitesList = sitesList
	app.logger.Printf("got %d sites to check\n", len(sitesList))

	app.tickerCh = time.NewTicker(app.freq)
	defer app.tickerCh.Stop()

	round := 0
	startTimerCh := time.NewTimer(time.Millisecond)
	defer startTimerCh.Stop()

	go app.showResults()

	for {
		select {
		case <-app.ctx.Done():
			return fmt.Errorf("timeout: %w", app.ctx.Err())
		case <-startTimerCh.C:
			startTimerCh.Stop()
			app.startRound(round)
			round += 1
		case <-app.tickerCh.C:
			app.startRound(round)
			round += 1
		}
	}
}

func (app *App) getSitesList() ([]string, error) {
	list := []string{}

	f, err := os.Open(app.filename)
	if err != nil {
		return nil, fmt.Errorf("can't open file %s: %w", app.filename, err)
	}
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		list = append(list, fmt.Sprintf("http://%s", scanner.Text()))
	}

	return list, err
}

func (app *App) startRound(round int) {
	app.logger.Printf("--- round %d ---\n", round)
	for i, url := range app.sitesList {
		app.maxRequestsCh <- MaxRequests{}

		go func(i int, url string) {
			time := time.Now()

			req, _ := http.NewRequestWithContext(app.ctx, http.MethodGet, url, nil)
			<-app.maxRequestsCh

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				app.statusListCh <- Status{Counter{round: round, idx: i}, url, time, 0, err}
				return
			}
			app.statusListCh <- Status{Counter{round: round, idx: i}, url, time, res.StatusCode, nil}
		}(i, url)
	}
}

func (app *App) showResults() {
	for {
		select {
		case <-app.ctx.Done():
			return
		case s := <-app.statusListCh:
			if s.err != nil {
				app.logger.Printf("%d-%d\t[err]\t%s\t%s\n", s.counter.round, s.counter.idx, s.time, s.err)
			} else {
				app.logger.Printf("%d-%d\t[%d]\t%s\t%s\n", s.counter.round, s.counter.idx, s.status, s.time, s.url)
			}
		}
	}
}
