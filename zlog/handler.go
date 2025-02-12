package zlog

import (
	"context"
	"log/slog"
	"sync"
)

// AttrExtractor is a function that retrieves or creates slog.Attr based
// on information/values found in the context.Context and the slog.Record.
type AttrExtractor func(ctx context.Context, record *slog.Record) slog.Attr

// HandlerOptions are options for a Handler
type HandlerOptions struct {
	// A list of functions to be called, each of which returns attributes
	// that should be prepended to the start of every log line with the context.
	// If left nil, the default ExtractPrepended function will be used.
	Prependers []AttrExtractor

	// A list of functions to be called, each of which will return attributes
	// that should be appended to the end of every log line with the context.
	// If left nil, the default ExtractAppended function will be used.
	Appenders []AttrExtractor
}

// Handler is a slog.Handler middleware that will prepend and append
// attributes to log lines. The attributes are extracted out of the log
// record's context by the provided AttrExtractor functions.
// It passes the final record and attributes off to the next handler when finished.
type Handler struct {
	next       slog.Handler
	goa        *groupOrAttrs
	prependers []AttrExtractor
	appenders  []AttrExtractor
}

// NewMiddleware creates a slog.Handler middleware
// that conforms to [github.com/samber/slog-multi.Middleware] interface.
// It can be used with slogmulti methods such as Pipe to easily setup
// a pipeline of slog handlers:
//
//	slog.SetDefault(slog.New(slogmulti.
//		Pipe(zlog.NewMiddleware(&zlog.HandlerOptions{})).
//		Pipe(zlog.NewOverwriteMiddleware(&slogdedup.OverwriteHandlerOptions{})).
//		Handler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})),
//	))
func NewMiddleware(opts *HandlerOptions) func(slog.Handler) slog.Handler {
	return func(next slog.Handler) slog.Handler {
		return NewHandler(next, opts)
	}
}

// NewHandler creates a slog.Handler middleware that will prepend and
// append attributes to log lines. The attributes are extracted out of the log
// record's context by the provided AttrExtractor functions.
// It passes the final record and attributes off to the next handler when finished.
// If opts is nil, the default options are used.
func NewHandler(next slog.Handler, opts *HandlerOptions) *Handler {
	if opts == nil {
		opts = &HandlerOptions{}
	}
	if opts.Prependers == nil {
		opts.Prependers = []AttrExtractor{ExtractPrepended}
	}
	if opts.Appenders == nil {
		opts.Appenders = []AttrExtractor{ExtractAppended}
	}

	return &Handler{
		next:       next,
		prependers: opts.Prependers,
		appenders:  opts.Appenders,
	}
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	if len(h.prependers) == 0 && len(h.appenders) == 0 && h.goa == nil {
		return h.next.Handle(ctx, r)
	}

	nBuf := r.NumAttrs() + len(h.appenders) + int(h.goa.getAttrNum())
	nFinal := len(h.prependers) + nBuf

	// slog.Record.AddAttrs iterate and copy every attr in the given attrs slice.
	// So if there is no group in the goa linked list, then it is safe to reuse
	// the tmp attrSlice, which is the most common case,
	// we can achieve better performance by reusing attrSlice.
	var tmpAttrs *attrSlice
	if h.goa == nil || h.goa.hasGroup == 0 {
		tmpAttrs = attrSlicePool.Get().(*attrSlice)
		tmpAttrs.ensureCap(nBuf, nFinal)
		defer tmpAttrs.recycle()
	} else {
		tmpAttrs = newAttrSlice(nBuf, nFinal)
	}

	// Collect all attributes from the record (which is the most recent).
	// These attributes are ordered from oldest to newest, and our collection will be too.
	r.Attrs(func(a slog.Attr) bool {
		tmpAttrs.appendAttrs(a)
		return true
	})

	// Add appended context attributes to the end.
	for _, f := range h.appenders {
		tmpAttrs.appendAttrs(f(ctx, &r))
	}

	// Iterate the goa (group or attributes) linked list, which is ordered from newest to oldest.
	for g := h.goa; g != nil; g = g.next {
		if g.group != "" {
			tmpAttrs.prependGroup(g.group)
		} else {
			tmpAttrs.prependAttrs(g.attrs...)
		}
	}

	// Add prepended context attributes and finalize the log attributes.
	final := tmpAttrs.final
	if len(h.prependers) > 0 {
		for _, f := range h.prependers {
			final = append(final, f(ctx, &r))
		}
		final = append(final, tmpAttrs.cur...)
	} else {
		final = tmpAttrs.cur
	}

	clone := slog.Record{
		Time:    r.Time,
		Message: r.Message,
		Level:   r.Level,
		PC:      r.PC,
	}
	clone.AddAttrs(final...)
	return h.next.Handle(ctx, clone)
}

func (h *Handler) WithGroup(name string) slog.Handler {
	clone := *h
	clone.goa = h.goa.withGroup(name)
	return &clone
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.goa = h.goa.withAttrs(attrs)
	return &clone
}

type attrSlice struct {
	buf   []slog.Attr // buffer enough to hold all attributes
	cur   []slog.Attr // current attr slice, sub-slice of buf
	final []slog.Attr // the final attributes
}

var attrSlicePool = sync.Pool{New: func() any { return &attrSlice{} }}

func newAttrSlice(nBuf, nFinal int) *attrSlice {
	n := nBuf + nFinal
	buf := make([]slog.Attr, 0, n)
	return &attrSlice{
		buf:   buf[:0:nBuf],
		cur:   buf[:0:nBuf],
		final: buf[nBuf:nBuf:n],
	}
}

func (p *attrSlice) ensureCap(nBuf, nFinal int) {
	if cap(p.buf) < nBuf {
		p.buf = make([]slog.Attr, 0, nBuf)
		p.cur = p.buf[:0]
	}
	if cap(p.final) < nFinal {
		p.final = make([]slog.Attr, 0, nFinal)
	}
}

func (p *attrSlice) recycle() {
	p.buf = p.buf[:0]
	p.cur = p.buf[:0]
	p.final = p.final[:0]
	attrSlicePool.Put(p)
}

func (p *attrSlice) appendAttrs(attrs ...slog.Attr) {
	p.cur = append(p.cur, attrs...)
}

func (p *attrSlice) prependAttrs(attrs ...slog.Attr) {
	n := len(attrs)
	copy(p.cur[n:len(p.cur)+n], p.cur)
	copy(p.cur[:n], attrs)
	p.cur = p.cur[:len(p.cur)+n]
}

func (p *attrSlice) prependGroup(group string) {
	cur := p.cur
	p.cur = append(p.cur, slog.Attr{
		Key:   group,
		Value: slog.GroupValue(cur...),
	})
	p.cur = p.cur[len(p.cur)-1:]
}

const maxUint16 = 1<<16 - 1

// groupOrAttrs holds either a group name or a list of slog.Attrs.
// It also holds a reference/link to its parent groupOrAttrs, forming a linked list.
type groupOrAttrs struct {
	group    string        // group name if non-empty
	attrs    []slog.Attr   // attrs if non-empty
	next     *groupOrAttrs // parent
	hasGroup uint16        // has group in the linked list
	nAttr    uint16        // estimate number of attrs in the linked list
}

// withGroup returns a new groupOrAttrs that includes the given group, and links to the old groupOrAttrs.
// Safe to call on a nil groupOrAttrs.
func (g *groupOrAttrs) withGroup(name string) *groupOrAttrs {
	// Empty-name groups are inlined as if they didn't exist.
	if name == "" {
		return g
	}
	if g.getAttrNum() == maxUint16 { // overflow uint16
		return g
	}
	return &groupOrAttrs{
		group:    name,
		next:     g,
		hasGroup: 1,
		nAttr:    g.getAttrNum() + 1,
	}
}

// withAttrs returns a new groupOrAttrs that includes the given attrs, and links to the old groupOrAttrs.
// Safe to call on a nil groupOrAttrs.
func (g *groupOrAttrs) withAttrs(attrs []slog.Attr) *groupOrAttrs {
	// Check empty attr slice or overflow uint16.
	n1, n2 := int(g.getAttrNum()), len(attrs)
	if n2 == 0 || n1+n2 > maxUint16 {
		return g
	}
	clone := &groupOrAttrs{
		attrs: attrs,
		next:  g,
		nAttr: uint16(n1 + n2),
	}
	if g != nil {
		clone.hasGroup = g.hasGroup
	}
	return clone
}

func (g *groupOrAttrs) getAttrNum() uint16 {
	if g == nil {
		return 0
	}
	return g.nAttr
}
