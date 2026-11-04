package judgeclient

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"testing"
	"time"

	judgev1 "github.com/alonespig/zeorcode/backend/internal/judgeproto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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

func TestPoolUsesConfiguredTLSAuthority(t *testing.T) {
	t.Parallel()

	const serverName = "judge-node.example.test"
	certificate, caPEM := newTestServerCertificate(t, serverName)
	caFile := t.TempDir() + "/ca.pem"
	if err := os.WriteFile(caFile, caPEM, 0o600); err != nil {
		t.Fatalf("write CA file: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	server := grpc.NewServer(grpc.Creds(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
	})))
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
	pool, err := New(ctx, Config{
		Addresses:  []string{listener.Addr().String()},
		Token:      "test-token",
		CAFile:     caFile,
		ServerName: serverName,
	})
	if err != nil {
		t.Fatalf("create TLS pool: %v", err)
	}
	defer func() { _ = pool.Close() }()

	if _, err := pool.Judge(ctx, &judgev1.JudgeRequest{}, nil); err != nil {
		t.Fatalf("judge over TLS authority override: %v", err)
	}
}

func newTestServerCertificate(t *testing.T, serverName string) (tls.Certificate, []byte) {
	t.Helper()

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate CA key: %v", err)
	}
	now := time.Now()
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "judge test CA"},
		NotBefore:             now.Add(-time.Minute),
		NotAfter:              now.Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create CA certificate: %v", err)
	}

	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate server key: %v", err)
	}
	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: serverName},
		DNSNames:     []string{serverName},
		NotBefore:    now.Add(-time.Minute),
		NotAfter:     now.Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("create server certificate: %v", err)
	}

	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER})
	serverPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})
	certificate, err := tls.X509KeyPair(serverPEM, keyPEM)
	if err != nil {
		t.Fatalf("parse server certificate: %v", err)
	}
	return certificate, caPEM
}
