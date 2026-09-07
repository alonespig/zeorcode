package judge

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCompileReturnsOriginalCompilerOutput(t *testing.T) {
	const compilerOutput = "a.cc:3:5: error: expected ';' before 'return'"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/run" {
			t.Fatalf("request path = %q, want /run", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `[{"status":"Nonzero Exit Status","files":{"stdout":"","stderr":%q},"fileIds":{}}]`, compilerOutput)
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL)
	artifact, err := client.Compile("c++", "int main() { return 0 }")
	if err == nil {
		t.Fatal("Compile() error = nil, want compilation failure")
	}
	if artifact != nil {
		t.Fatalf("Compile() artifact = %#v, want nil", artifact)
	}
	if got := CompilationOutput(err); got != compilerOutput {
		t.Fatalf("CompilationOutput() = %q, want %q", got, compilerOutput)
	}
}

func TestCompilationOutputIgnoresInfrastructureErrors(t *testing.T) {
	if got := CompilationOutput(fmt.Errorf("judge unavailable")); got != "" {
		t.Fatalf("CompilationOutput() = %q, want empty", got)
	}
}
