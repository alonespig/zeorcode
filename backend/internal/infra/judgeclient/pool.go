package judgeclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	judgev1 "github.com/alonespig/zeorcode/backend/api/judge/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Config struct {
	Addresses  []string
	Token      string
	Insecure   bool
	CAFile     string
	ServerName string
}

type node struct {
	address  string
	conn     *grpc.ClientConn
	client   judgev1.JudgeNodeServiceClient
	healthy  atomic.Bool
	inflight atomic.Int32
	reported atomic.Int32
	max      atomic.Int32
}

type Pool struct {
	nodes  []*node
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func New(ctx context.Context, cfg Config) (*Pool, error) {
	if strings.TrimSpace(cfg.Token) == "" {
		return nil, errors.New("judge node token is required")
	}
	transport, err := transportCredentials(cfg)
	if err != nil {
		return nil, err
	}
	poolCtx, cancel := context.WithCancel(ctx)
	p := &Pool{cancel: cancel}
	for _, rawAddress := range cfg.Addresses {
		address := strings.TrimSpace(rawAddress)
		if address == "" {
			continue
		}
		conn, err := grpc.NewClient(
			address,
			grpc.WithTransportCredentials(transport),
			grpc.WithPerRPCCredentials(bearerCredentials{token: cfg.Token, allowInsecure: cfg.Insecure}),
		)
		if err != nil {
			p.Close()
			return nil, fmt.Errorf("create judge node client %s: %w", address, err)
		}
		n := &node{address: address, conn: conn, client: judgev1.NewJudgeNodeServiceClient(conn)}
		n.healthy.Store(true)
		n.max.Store(1)
		p.nodes = append(p.nodes, n)
	}
	if len(p.nodes) == 0 {
		cancel()
		return nil, errors.New("no judge node addresses configured")
	}
	p.wg.Add(1)
	go p.healthLoop(poolCtx)
	return p, nil
}

func (p *Pool) Close() error {
	if p == nil {
		return nil
	}
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
	var closeErr error
	for _, n := range p.nodes {
		closeErr = errors.Join(closeErr, n.conn.Close())
	}
	return closeErr
}

func (p *Pool) Judge(ctx context.Context, req *judgev1.JudgeRequest, onEvent func(*judgev1.JudgeEvent) error) (*judgev1.FinalResult, error) {
	first := p.pick(nil)
	if first == nil {
		return nil, errors.New("no healthy judge node available")
	}
	final, err := p.judgeOnNode(ctx, first, req, onEvent)
	if err == nil || !retryable(err) {
		return final, err
	}
	first.healthy.Store(false)
	second := p.pick(first)
	if second == nil {
		return nil, err
	}
	return p.judgeOnNode(ctx, second, req, onEvent)
}

func (p *Pool) judgeOnNode(ctx context.Context, n *node, req *judgev1.JudgeRequest, onEvent func(*judgev1.JudgeEvent) error) (*judgev1.FinalResult, error) {
	n.inflight.Add(1)
	defer n.inflight.Add(-1)
	stream, err := n.client.Judge(ctx, req)
	if err != nil {
		return nil, err
	}
	for {
		event, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil, errors.New("judge node stream ended without final result")
		}
		if err != nil {
			return nil, err
		}
		if onEvent != nil {
			if err := onEvent(event); err != nil {
				return nil, err
			}
		}
		if event.Type == judgev1.JudgeEventType_JUDGE_EVENT_TYPE_FINISHED {
			if event.FinalResult == nil {
				return nil, errors.New("judge node returned empty final result")
			}
			return event.FinalResult, nil
		}
		if event.Type == judgev1.JudgeEventType_JUDGE_EVENT_TYPE_FAILED {
			return nil, errors.New("judge node reported failure")
		}
	}
}

func (p *Pool) pick(exclude *node) *node {
	var selected *node
	var selectedLoad int64
	for _, candidate := range p.nodes {
		if candidate == exclude || !candidate.healthy.Load() {
			continue
		}
		max := candidate.max.Load()
		if max <= 0 {
			max = 1
		}
		running := candidate.inflight.Load()
		if reported := candidate.reported.Load(); reported > running {
			running = reported
		}
		load := int64(running) * 1_000_000 / int64(max)
		if selected == nil || load < selectedLoad {
			selected, selectedLoad = candidate, load
		}
	}
	return selected
}

func (p *Pool) healthLoop(ctx context.Context) {
	defer p.wg.Done()
	p.checkAll(ctx)
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.checkAll(ctx)
		}
	}
}

func (p *Pool) checkAll(ctx context.Context) {
	var wg sync.WaitGroup
	for _, candidate := range p.nodes {
		wg.Add(1)
		go func(n *node) {
			defer wg.Done()
			probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			result, err := n.client.Status(probeCtx, &judgev1.StatusRequest{})
			if err != nil {
				n.healthy.Store(false)
				return
			}
			n.healthy.Store(true)
			n.reported.Store(result.RunningTasks)
			n.max.Store(result.MaxConcurrency)
		}(candidate)
	}
	wg.Wait()
}

func retryable(err error) bool {
	switch status.Code(err) {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}

type bearerCredentials struct {
	token         string
	allowInsecure bool
}

func (c bearerCredentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + c.token}, nil
}

func (c bearerCredentials) RequireTransportSecurity() bool { return !c.allowInsecure }

func transportCredentials(cfg Config) (credentials.TransportCredentials, error) {
	if cfg.Insecure {
		return insecure.NewCredentials(), nil
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: strings.TrimSpace(cfg.ServerName)}
	if cfg.CAFile != "" {
		pemBytes, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("read judge node CA: %w", err)
		}
		pool, err := x509.SystemCertPool()
		if err != nil || pool == nil {
			pool = x509.NewCertPool()
		}
		if !pool.AppendCertsFromPEM(pemBytes) {
			return nil, errors.New("judge node CA file contains no certificates")
		}
		tlsConfig.RootCAs = pool
	}
	return credentials.NewTLS(tlsConfig), nil
}
