package gocron

import (
    "testing"
    "context"
    "time"
)

type Counter struct {
    times int
}

func (c* Counter) Run(t time.Time) {
    c.times += 1
}

func TestCronJob(t *testing.T) {
    job := &Counter{0}
    interval := time.Millisecond * 100

    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    c := NewCron(ctx, interval, job)
    c.Start()

    got := job.times
    want := 10

    if got != want {
        t.Errorf("want %d, but got %d", want, got)
    }
}

