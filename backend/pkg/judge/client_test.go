package judge

import (
	"encoding/json"
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

func TestRunUsesConfiguredOutputLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload struct {
			Cmd []struct {
				Files []struct {
					Max int `json:"max"`
				} `json:"files"`
			} `json:"cmd"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if got := payload.Cmd[0].Files[1].Max; got != 2048 {
			t.Fatalf("stdout max = %d, want 2048", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"status":"Accepted","files":{"stdout":"ok","stderr":""},"fileIds":{}}]`)
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL)
	verdict, err := client.Run(&Artifact{Language: LanguageCPP, FileID: "artifact"}, "", Limits{OutputBytes: 2048})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if verdict.Stdout != "ok" {
		t.Fatalf("Run() stdout = %q, want ok", verdict.Stdout)
	}
}

func TestCompilationOutputIgnoresInfrastructureErrors(t *testing.T) {
	if got := CompilationOutput(fmt.Errorf("judge unavailable")); got != "" {
		t.Fatalf("CompilationOutput() = %q, want empty", got)
	}
}
