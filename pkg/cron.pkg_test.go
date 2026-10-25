package sdk_pkg

import (
	"testing"
	"time"
)

// NewCron returns the ICron interface which only exposes RegisterJob;
// Start/Shutdown live on the concrete scheduler, so tests assert through.
func newTestCron(t *testing.T) *cron {
	t.Helper()

	c, err := NewCron()
	if err != nil {
		t.Fatalf("NewCron: %v", err)
	}

	return c.(*cron)
}

func TestCronLifecycle(t *testing.T) {
	c := newTestCron(t)

	fired := make(chan struct{}, 4)

	// gocron CronJob here runs with seconds precision, so "* * * * * ?"
	// fires every second and keeps the test fast.
	job, err := c.RegisterJob("ticker", "* * * * * ?", func() {
		select {
		case fired <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatalf("RegisterJob: %v", err)
	}

	if job == nil {
		t.Fatal("RegisterJob returned nil job")
	}

	c.Start()

	select {
	case <-fired:
	case <-time.After(3 * time.Second):
		t.Fatal("job did not fire within 3s")
	}

	if err := c.Shutdown(); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

func TestCronRegisterJobValidation(t *testing.T) {
	c := newTestCron(t)

	defer func() { _ = c.Shutdown() }()

	cases := []struct {
		jobName string
		cron    string
		task    func()
		wantErr bool
	}{
		{"", "* * * * * ?", func() {}, true},
		{"name", "", func() {}, true},
		{"name", "* * * * * ?", nil, true},
		{"name", "definitely not a cron", func() {}, true},
		{"valid", "* * * * * ?", func() {}, false},
	}

	for _, tc := range cases {
		_, err := c.RegisterJob(tc.jobName, tc.cron, tc.task)
		if tc.wantErr && err == nil {
			t.Errorf("%q/%q/%v: expected error", tc.jobName, tc.cron, tc.task == nil)
		}

		if !tc.wantErr && err != nil {
			t.Errorf("%q/%q: unexpected error %v", tc.jobName, tc.cron, err)
		}
	}
}

func TestCronShutdownTwice(t *testing.T) {
	if testing.Short() {
		t.Skip("slow: gocron's second Shutdown blocks ~12s before timing out")
	}

	c := newTestCron(t)

	if _, err := c.RegisterJob("warmup", "* * * * * ?", func() {}); err != nil {
		t.Fatal(err)
	}

	c.Start()

	time.Sleep(1200 * time.Millisecond)

	if err := c.Shutdown(); err != nil {
		t.Fatalf("first shutdown: %v", err)
	}

	// gocron surfaces "timed out waiting for scheduler to stop" rather
	// than a clean ErrSchedulerStopped; either way the second call must
	// not report success.
	if err := c.Shutdown(); err == nil {
		t.Fatal("second shutdown must report the scheduler is stopped")
	}
}
