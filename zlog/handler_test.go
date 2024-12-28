package zlog

import (
	"context"
	"encoding/json"
	"log/slog"
	"slices"
	"strings"
	"testing"

	"github.com/jxskiss/gopkg/v2/zlog/internal/test"
)

type logLine struct {
	Source struct {
		Function string `json:"function"`
		File     string `json:"file"`
		Line     int    `json:"line"`
	} `json:"source"`
}

func TestHandler(t *testing.T) {
	t.Parallel()

	tester := &test.Handler{}
	h := NewHandler(tester, nil)

	ctx := PrependAttrs(nil, "prepend1", "arg1", slog.String("prepend1", "arg2"))
	ctx = PrependAttrs(ctx, "prepend2", "arg1", "prepend2", "arg2")
	PrependAttrs(ctx, "prepend3", "arg1", "prepend3", "arg2") // Ensure we aren't overwriting the parent context
	ctx = AppendAttrs(ctx, "append1", "arg1", "append1", "arg2")
	ctx = AppendAttrs(ctx, slog.String("append2", "arg1"), "append2", "arg2")
	AppendAttrs(ctx, "append3", "arg1", "append3", "arg2") // Ensure we aren't overwriting the parent context
	AppendAttrs(nil, "append4", "arg1", "badkey")
	AppendAttrs(ctx, int64(123))

	l := slog.New(h)

	l = l.With("with1", "arg1", "with1", "arg2").With()
	l = l.WithGroup("group1").WithGroup("")
	l = l.With("with2", "arg1", "with2", "arg2")

	l.InfoContext(ctx, "main message", "main1", "arg1", "main1", "arg2")

	expectedText := `time=2023-09-29T13:00:59.000Z level=INFO msg="main message" prepend1=arg1 prepend1=arg2 prepend2=arg1 prepend2=arg2 with1=arg1 with1=arg2 group1.with2=arg1 group1.with2=arg2 group1.main1=arg1 group1.main1=arg2 group1.append1=arg1 group1.append1=arg2 group1.append2=arg1 group1.append2=arg2
`
	if s := tester.String(); s != expectedText {
		t.Errorf("Expected:\n%s\nGot:\n%s\n", expectedText, s)
	}

	b, err := tester.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	expectedJSON := `{"time":"2023-09-29T13:00:59Z","level":"INFO","msg":"main message","prepend1":"arg1","prepend1":"arg2","prepend2":"arg1","prepend2":"arg2","with1":"arg1","with1":"arg2","group1":{"with2":"arg1","with2":"arg2","main1":"arg1","main1":"arg2","append1":"arg1","append1":"arg2","append2":"arg1","append2":"arg2"}}
`
	if string(b) != expectedJSON {
		t.Errorf("Expected:\n%s\nGot:\n%s\n", expectedText, string(b))
	}

	// Check the source location fields
	tester.Source = true
	b, err = tester.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var unmarshalled logLine
	err = json.Unmarshal(b, &unmarshalled)
	if err != nil {
		t.Fatal(err)
	}

	if unmarshalled.Source.Function != "github.com/jxskiss/gopkg/v2/zlog.TestHandler" ||
		!strings.HasSuffix(unmarshalled.Source.File, "zlog/handler_test.go") ||
		unmarshalled.Source.Line != 43 {
		t.Errorf("Expected source fields are incorrect: %#+v\n", unmarshalled)
	}
}

func TestHandlerMultipleAttrExtractor(t *testing.T) {
	t.Parallel()

	tester := &test.Handler{}
	h := NewMiddleware(&HandlerOptions{
		Prependers: []AttrExtractor{
			ExtractPrepended,
			func(ctx context.Context, _ *slog.Record) slog.Attr {
				if v, ok := ctx.Value(prependKey).([]slog.Attr); ok {
					v = slices.Clone(v)
					for i := 0; i < len(v); i++ {
						v[i].Key += "^"
					}
					return slog.Attr{Value: slog.GroupValue(v...)}
				}
				return slog.Attr{}
			},
			func(_ context.Context, _ *slog.Record) slog.Attr {
				return slog.Attr{}
			},
		},
		Appenders: []AttrExtractor{
			ExtractAppended,
			func(ctx context.Context, _ *slog.Record) slog.Attr {
				if v, ok := ctx.Value(appendKey).([]slog.Attr); ok {
					v = slices.Clone(v)
					for i := 0; i < len(v); i++ {
						v[i].Key += "*"
					}
					return slog.Attr{Value: slog.GroupValue(v...)}
				}
				return slog.Attr{}
			},
			func(_ context.Context, _ *slog.Record) slog.Attr {
				return slog.Attr{}
			},
		},
	})(tester)

	ctx := PrependAttrs(nil, "prepend1", "arg1", slog.String("prepend1", "arg2"))
	ctx = PrependAttrs(ctx, "prepend2", "arg1", "prepend2", "arg2")
	PrependAttrs(ctx, "prepend3", "arg1", "prepend3", "arg2") // Ensure we aren't overwriting the parent context
	ctx = AppendAttrs(ctx, "append1", "arg1", "append1", "arg2")
	ctx = AppendAttrs(ctx, slog.String("append2", "arg1"), "append2", "arg2")
	AppendAttrs(ctx, "append3", "arg1", "append3", "arg2") // Ensure we aren't overwriting the parent context
	AppendAttrs(nil, "append4", "arg1", "badkey")
	AppendAttrs(ctx, int64(123))

	l := slog.New(h)

	l = l.With("with1", "arg1", "with1", "arg2")
	l = l.WithGroup("group1")
	l = l.With("with2", "arg1", "with2", "arg2")

	l.InfoContext(ctx, "main message", "main1", "arg1", "main1", "arg2")

	expectedText := `time=2023-09-29T13:00:59.000Z level=INFO msg="main message" prepend1=arg1 prepend1=arg2 prepend2=arg1 prepend2=arg2 prepend1^=arg1 prepend1^=arg2 prepend2^=arg1 prepend2^=arg2 with1=arg1 with1=arg2 group1.with2=arg1 group1.with2=arg2 group1.main1=arg1 group1.main1=arg2 group1.append1=arg1 group1.append1=arg2 group1.append2=arg1 group1.append2=arg2 group1.append1*=arg1 group1.append1*=arg2 group1.append2*=arg1 group1.append2*=arg2
`
	if s := tester.String(); s != expectedText {
		t.Errorf("Expected:\n%s\nGot:\n%s\n", expectedText, s)
	}

	b, err := tester.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	expectedJSON := `{"time":"2023-09-29T13:00:59Z","level":"INFO","msg":"main message","prepend1":"arg1","prepend1":"arg2","prepend2":"arg1","prepend2":"arg2","prepend1^":"arg1","prepend1^":"arg2","prepend2^":"arg1","prepend2^":"arg2","with1":"arg1","with1":"arg2","group1":{"with2":"arg1","with2":"arg2","main1":"arg1","main1":"arg2","append1":"arg1","append1":"arg2","append2":"arg1","append2":"arg2","append1*":"arg1","append1*":"arg2","append2*":"arg1","append2*":"arg2"}}
`
	if string(b) != expectedJSON {
		t.Errorf("Expected:\n%s\nGot:\n%s\n", expectedText, string(b))
	}
}
