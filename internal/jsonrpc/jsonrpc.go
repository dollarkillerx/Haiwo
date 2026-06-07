package jsonrpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Handler func(context.Context, json.RawMessage) (any, *Error)

type Peer struct {
	conn     *websocket.Conn
	mu       sync.Mutex
	nextID   atomic.Int64
	handlers map[string]Handler
	pending  map[string]chan Response
	closed   chan struct{}
}

func NewPeer(conn *websocket.Conn) *Peer {
	return &Peer{
		conn:     conn,
		handlers: make(map[string]Handler),
		pending:  make(map[string]chan Response),
		closed:   make(chan struct{}),
	}
}

func (p *Peer) Handle(method string, h Handler) {
	p.handlers[method] = h
}

func (p *Peer) Run(ctx context.Context) error {
	defer close(p.closed)
	for {
		_, data, err := p.conn.ReadMessage()
		if err != nil {
			return err
		}
		var probe struct {
			Method string          `json:"method"`
			ID     json.RawMessage `json:"id"`
			Result json.RawMessage `json:"result"`
			Error  *Error          `json:"error"`
		}
		if err := json.Unmarshal(data, &probe); err != nil {
			continue
		}
		if probe.Method != "" {
			var req Request
			if err := json.Unmarshal(data, &req); err == nil {
				go p.dispatch(ctx, req)
			}
			continue
		}
		if len(probe.ID) > 0 {
			key := string(probe.ID)
			p.mu.Lock()
			ch := p.pending[key]
			delete(p.pending, key)
			p.mu.Unlock()
			if ch != nil {
				ch <- Response{JSONRPC: "2.0", ID: probe.ID, Result: probe.Result, Error: probe.Error}
			}
		}
	}
}

func (p *Peer) Call(ctx context.Context, method string, params any, result any) error {
	id := p.nextID.Add(1)
	idRaw := json.RawMessage(fmt.Sprintf("%d", id))
	paramsRaw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	ch := make(chan Response, 1)
	p.mu.Lock()
	p.pending[string(idRaw)] = ch
	p.mu.Unlock()
	if err := p.write(Request{JSONRPC: "2.0", ID: idRaw, Method: method, Params: paramsRaw}); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.closed:
		return errors.New("jsonrpc peer closed")
	case resp := <-ch:
		if resp.Error != nil {
			return errors.New(resp.Error.Message)
		}
		if result != nil && len(resp.Result) > 0 {
			return json.Unmarshal(resp.Result, result)
		}
		return nil
	}
}

func (p *Peer) Notify(method string, params any) error {
	paramsRaw, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return p.write(Request{JSONRPC: "2.0", Method: method, Params: paramsRaw})
}

func (p *Peer) Close() error {
	return p.conn.Close()
}

func (p *Peer) dispatch(ctx context.Context, req Request) {
	h := p.handlers[req.Method]
	if h == nil {
		p.respond(req.ID, nil, &Error{Code: -32601, Message: "method not found"})
		return
	}
	result, rpcErr := h(ctx, req.Params)
	p.respond(req.ID, result, rpcErr)
}

func (p *Peer) respond(id json.RawMessage, result any, rpcErr *Error) {
	if len(id) == 0 {
		return
	}
	var raw json.RawMessage
	if rpcErr == nil {
		raw, _ = json.Marshal(result)
	}
	_ = p.write(Response{JSONRPC: "2.0", ID: id, Result: raw, Error: rpcErr})
}

func (p *Peer) write(v any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.conn.WriteJSON(v)
}
