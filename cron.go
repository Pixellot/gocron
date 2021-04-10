package gocron

import (
    "time"
    "context"
)

type Clock struct {
    Hour, Minutes, Seconds, Nanoseconds int
}

type Job interface {
    Run(t time.Time)
}

type Cron struct {
    ctx context.Context
    interval time.Duration

    job Job
}

func NewCron(ctx context.Context, interval time.Duration, job Job) *Cron {
    return &Cron{ctx: ctx, interval: interval, job: job}
}

func (c *Cron) Start() {
    for t := range cronFromNow(c.ctx, c.interval) {
        c.job.Run(t)
    }
}

func cronFromNow(ctx context.Context, interval time.Duration) <-chan time.Time {

    stream := make(chan time.Time, 1)

    go func() {
        tick(ctx, interval, stream)
    }()

    return stream
}

func cronFromDate(ctx context.Context, interval time.Duration, start time.Time) <-chan time.Time {

    now := time.Now()
    if start.Sub(now) < 0 {
        start = now
    }

    stream := make(chan time.Time, 1)

    go func() {

        delay := start.Sub(time.Now())
        if delay < 0 {
            delay = 0
        }
        t := <-time.After(delay)
        stream <- t

        tick(ctx, interval, stream)
    }()

    return stream
}

func cronFromClock(ctx context.Context, interval time.Duration, start Clock) <-chan time.Time {

    stream := make(chan time.Time, 1)

    go func() {

        t := <-time.After(sync(time.Now(), start))
        stream <- t

        tick(ctx, interval, stream)
    }()

    return stream
}

func tick(ctx context.Context, interval time.Duration, stream chan time.Time) {

    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        select {
        case t := <-ticker.C:
            stream <- t
        case <-ctx.Done():
            close(stream)
            return
        }
    }
}

func sync(t time.Time, c Clock) time.Duration {

    req := time.Date(
        t.Year(), t.Month(), t.Day(),
        c.Hour, c.Minutes, c.Seconds, c.Nanoseconds,
        t.Location())

    if diff := req.Sub(t); diff < 0 {
        req = req.AddDate(0, 0, 1)
    }

    return req.Sub(t)
}
