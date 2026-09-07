package llm

import (
	"context"
	"errors"
	"reflect"
	"testing"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type streamingModelStub struct {
	chunks []*schema.Message
}

func (s *streamingModelStub) Generate(context.Context, []*schema.Message, ...einomodel.Option) (*schema.Message, error) {
	return schema.AssistantMessage("unused", nil), nil
}

func (s *streamingModelStub) Stream(context.Context, []*schema.Message, ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	return schema.StreamReaderFromArray(s.chunks), nil
}

func TestClientStreamEmitsDeltasAndReturnsCompleteAnswer(t *testing.T) {
	client := &Client{
		model: &streamingModelStub{chunks: []*schema.Message{
			schema.AssistantMessage("你好", nil),
			schema.AssistantMessage("，", nil),
			schema.AssistantMessage("这里是 ZeorCode。", nil),
		}},
		enabled: true,
	}

	var deltas []string
	answer, err := client.Stream(context.Background(), "system", []Message{{Role: "user", Content: "介绍一下"}}, func(delta string) error {
		deltas = append(deltas, delta)
		return nil
	})
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	if answer != "你好，这里是 ZeorCode。" {
		t.Fatalf("Stream() answer = %q", answer)
	}
	wantDeltas := []string{"你好", "，", "这里是 ZeorCode。"}
	if !reflect.DeepEqual(deltas, wantDeltas) {
		t.Fatalf("Stream() deltas = %#v, want %#v", deltas, wantDeltas)
	}
}

func TestClientStreamStopsWhenDeltaConsumerFails(t *testing.T) {
	client := &Client{
		model: &streamingModelStub{chunks: []*schema.Message{
			schema.AssistantMessage("first", nil),
			schema.AssistantMessage("second", nil),
		}},
		enabled: true,
	}
	wantErr := errors.New("client disconnected")
	calls := 0

	_, err := client.Stream(context.Background(), "system", nil, func(string) error {
		calls++
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Stream() error = %v, want %v", err, wantErr)
	}
	if calls != 1 {
		t.Fatalf("delta callback calls = %d, want 1", calls)
	}
}
