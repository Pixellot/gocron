package gocron

import (
    "testing"
    "context"
    "time"
    "sync"
)

type Counter struct {
    times int
    mtx sync.Mutex
}

func (c* Counter) Run(t time.Time) {
    c.mtx.Lock()
    defer c.mtx.Unlock()
    c.times += 1
}

func TestCronJobWithoutStartTime(t *testing.T) {
    job := &Counter{times: 0}
    interval := time.Millisecond * 10

    ctx, cancel := context.WithTimeout(context.Background(),
        time.Millisecond * 100)
    defer cancel()

    c := NewCron(job, ctx, interval)
    c.Start()

    got := job.times
    want := 10

    if got < want - 1 || got > want + 1 {
        t.Errorf("want %d, but got %d", want, got)
    }
}

func TestMultipleCronOnSameJob(t *testing.T) {
    job := &Counter{times: 0}
    interval := time.Millisecond * 10

    ctx, cancel := context.WithTimeout(context.Background(),
        time.Millisecond * 100)
    defer cancel()

    c1 := NewCron(job, ctx, interval)
    go func() {
        c1.Start()
    }()

    c2 := NewCron(job, ctx, interval)
    c2.Start()

    got := job.times
    want := 20

    if got < want - 1 || got > want + 1 {
        t.Errorf("want %d, but got %d", want, got)
    }
}

func TestCronJobStartTime(t *testing.T) {
    job := &Counter{times: 0}
    interval := time.Millisecond * 10

    ctx, cancel := context.WithTimeout(context.Background(),
        time.Millisecond * 100 + time.Millisecond * 100)
    defer cancel()

    start := time.Now().Add(time.Millisecond * 100)

    c := NewCron(job, ctx, interval, start)
    c.Start()

    got := job.times
    want := 10

    if got < want - 1 || got > want + 1 {
        t.Errorf("want %d, but got %d", want, got)
    }
}

func TestCronJobStartClock(t *testing.T) {
    job := &Counter{times: 0}
    interval := time.Millisecond * 10

    ctx, cancel := context.WithTimeout(context.Background(),
        time.Millisecond * 100 + time.Millisecond * 100)
    defer cancel()

    now := time.Now().Add(time.Millisecond * 100)
    start := ClockToTime(now.Hour(), now.Minute(), now.Second(), now.Nanosecond())

    c := NewCron(job, ctx, interval, start)
    c.Start()

    got := job.times
    want := 10

    if got < want - 1 || got > want + 1 {
        t.Errorf("want %d, but got %d", want, got)
    }
}
