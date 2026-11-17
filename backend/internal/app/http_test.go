package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestHTTPShutdownDrainsRequestsAndWaitsForReminder(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	requestStarted := make(chan struct{})
	requestRelease := make(chan struct{})
	finishRequest := sync.OnceFunc(func() { close(requestRelease) })
	t.Cleanup(finishRequest)
	reminderCanceled := make(chan struct{})
	reminderRelease := make(chan struct{})
	finishReminder := sync.OnceFunc(func() { close(reminderRelease) })
	t.Cleanup(finishReminder)
	reminderExited := make(chan struct{})
	app := &HTTP{
		handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			close(requestStarted)
			<-requestRelease
			w.WriteHeader(http.StatusNoContent)
		}),
		runReminder: func(ctx context.Context) {
			defer close(reminderExited)
			<-ctx.Done()
			close(reminderCanceled)
			<-reminderRelease
		},
	}
	serveDone := make(chan error, 1)
	go func() { serveDone <- app.serve(ctx, listener) }()
	requestDone := make(chan error, 1)
	go func() {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("http://" + listener.Addr().String())
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusNoContent {
				err = fmt.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
			}
		}
		requestDone <- err
	}()
	select {
	case <-requestStarted:
	case err := <-requestDone:
		t.Fatalf("request ended before entering handler: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("HTTP request did not start")
	}
	cancel()
	select {
	case <-reminderCanceled:
	case <-time.After(5 * time.Second):
		t.Fatal("reminder context was not canceled")
	}
	finishRequest()
	if err := <-requestDone; err != nil {
		t.Fatalf("in-flight HTTP request did not finish normally: %v", err)
	}
	select {
	case err := <-serveDone:
		t.Fatalf("app exited before the reminder stopped: %v", err)
	default:
	}
	finishReminder()
	select {
	case err := <-serveDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("app did not exit")
	}
	select {
	case <-reminderExited:
	default:
		t.Fatal("app exited while reminder was still running")
	}
}

func TestHTTPDoesNotStartReminderWhenListenFails(t *testing.T) {
	started := make(chan struct{})
	app := &HTTP{runReminder: func(context.Context) { close(started) }}
	if err := app.Run(context.Background(), "127.0.0.1:-1"); err == nil {
		t.Fatal("Run accepted an invalid port")
	}
	select {
	case <-started:
		t.Fatal("reminder started even though HTTP could not listen")
	default:
	}
}
