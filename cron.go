// Package gocron povides simple functionlity for scheduling cron jobs.
package gocron

import (
    "time"
    "context"
    "reflect"
)

// A Job is an interface for a scheduled job.
//
// Run performs the job while t is the job's execution time.
type Job interface {
    Run(t time.Time)
}

// A Cron represents a scheduled job.
type Cron struct {
    job Job

    ctx context.Context
    interval time.Duration

    start []time.Time
}

// NewCron returns a scheduled context enabled job with specified interval and an optional start time.
func NewCron(job Job, ctx context.Context, interval time.Duration, start ...time.Time) *Cron {
    c := &Cron{job: job, ctx: ctx, interval: interval}
    if len(start) > 0 {
        c.start = append(c.start, start[0])
    }
    return c
}

// Start performs a blocking call for starting the cron job.
func (c *Cron) Start() {
    for t := range cron(c.ctx, c.interval, c.start...) {
        c.job.Run(t)
    }
}

// ClockToTime returns Time instance from a specified clock.
// It is an auxiliary function when start time should not be dependent on date.
func ClockToTime(hour, minute, second, nanosecond int) time.Time {
    empty := time.Time{}
    location := time.Now().Location()
    return time.Date(empty.Year(), empty.Month(), empty.Day(), hour, minute, second, nanosecond, location)
}


func cron(ctx context.Context, interval time.Duration, start ...time.Time) <-chan time.Time {

    stream := make(chan time.Time, 1)

    go func() {

        delay := time.Second * 0
        if len(start) > 0 {
            delay = synch(time.Now(), start[0])
        }
        if proceed := wait(ctx, delay, stream); !proceed {
            return
        }

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

func wait(ctx context.Context, delay time.Duration, stream chan time.Time) bool {

    select {
    case t := <-time.After(delay):
        stream <- t
        return true
    case <-ctx.Done():
        close(stream)
        return false
    }
}

func synch(ref, wanted time.Time) time.Duration {
    if ref.IsZero() {
        return 0
    }

    ref = ref.In(wanted.Location())

    wantedDate := auxDate{}
    wantedDate.y, wantedDate.m, wantedDate.d = wanted.Date()
    emptyDate := auxDate{}
    emptyDate.y, emptyDate.m, emptyDate.d = time.Time{}.Date()

    if reflect.DeepEqual(wantedDate, emptyDate) {
        wanted = time.Date(
            ref.Year(), ref.Month(), ref.Day(),
            wanted.Hour(), wanted.Minute(), wanted.Second(), wanted.Nanosecond(),
            wanted.Location())

        if diff := wanted.Sub(ref); diff < 0 {
            wanted = wanted.AddDate(0, 0, 1)
        }

        return wanted.Sub(ref)
    }

    delay := wanted.Sub(ref)
    if delay < 0 {
        delay = 0
    }

    return delay
}

type auxDate struct {
    y, d int
    m time.Month
}

