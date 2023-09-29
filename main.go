package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/dimishpatriot/is-live/internal/app"
)

func main() {
	var a *app.App

	sizeF := flag.String("s", "medium", "length of sites list short|medium")
	timeoutF := flag.Duration("t", time.Second*60, "time of execution")
	freqF := flag.Duration("f", time.Second*10, "time between rounds")
	limitF := flag.Int("l", 20, "max request by time")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeoutF)
	defer cancel()

	switch *sizeF {
	case "short":
		a = app.New(ctx, "./data/site-list-short.txt", *freqF, *limitF)
	case "medium":
		a = app.New(ctx, "./data/site-list-medium.txt", *freqF, *limitF)
	default:
		log.Fatalf("unexpected size flag: %s", *sizeF)
	}

	log.Fatal(a.Run())
}
