package judgeclient

import (
	"context"
	"net"
	"testing"
	"time"

	judgev1 "github.com/alonespig/zeorcode/backend/internal/judgeproto"

	"google.golang.org/grpc"
)

type fakeJudgeNode struct {
	judgev1.UnimplementedJudgeNodeServiceServer
}

func (fakeJudgeNode) Status(context.Context, *judgev1.StatusRequest) (*judgev1.StatusResponse, error) {
	return &judgev1.StatusResponse{NodeId: "test-node", MaxConcurrency: 2}, nil
}

func (fakeJudgeNode) Judge(_ *judgev1.JudgeRequest, stream judgev1.JudgeNodeService_JudgeServer) error {
	if err := stream.Send(&judgev1.JudgeEvent{
		Type:       judgev1.JudgeEventType_JUDGE_EVENT_TYPE_CASE_RESULT,
		CaseResult: &judgev1.CaseResult{CaseId: 1, Status: 1},
	}); err != nil {
		return err
	}
	return stream.Send(&judgev1.JudgeEvent{
		Type:        judgev1.JudgeEventType_JUDGE_EVENT_TYPE_FINISHED,
		FinalResult: &judgev1.FinalResult{Status: 1},
	})
}

func TestPoolJudgeStreamsFinalResult(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := grpc.NewServer()
	judgev1.RegisterJudgeNodeServiceServer(server, fakeJudgeNode{})
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
		<-serveDone
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := New(ctx, Config{Addresses: []string{listener.Addr().String()}, Token: "test-token", Insecure: true})
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	defer func() { _ = pool.Close() }()

	var caseEvents int
	final, err := pool.Judge(ctx, &judgev1.JudgeRequest{}, func(event *judgev1.JudgeEvent) error {
		if event.Type == judgev1.JudgeEventType_JUDGE_EVENT_TYPE_CASE_RESULT {
			caseEvents++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("judge: %v", err)
	}
	if caseEvents != 1 || final.Status != 1 {
		t.Fatalf("unexpected stream result: cases=%d final=%+v", caseEvents, final)
	}
}
