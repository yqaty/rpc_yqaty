package rpc_yqaty

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"text/scanner"
	"time"
)

type ClientCodec struct {
	conn    io.ReadWriteCloser
	encoder *Encoder
	decoder *Decoder
}

func (codec *ClientCodec) WriteRequest(req *Request, data any) error {
	codec.encoder.s = new(bytes.Buffer)
	if err := codec.encoder.JSONEncode(req); err != nil {
		fmt.Println(err)
		return err
	}
	if err := codec.encoder.JSONEncode(data); err != nil {
		return err
	}
	codec.encoder.s.WriteString(" ")
	_, err := codec.conn.Write(codec.encoder.s.Bytes())
	return err
}

func (codec *ClientCodec) ReadResponseHeader(resp *Response) error {
	return codec.decoder.JSONDecode(resp)
}

func (codec *ClientCodec) ReadResponseBody(data *Data) error {
	return codec.decoder.JSONDecode(data)
}

type Query struct {
	Method string
	Args   any
	Reply  any
	Error  error
	seq    uint64
	Done   chan *Query
}

func (query *Query) done() {
	select {
	case query.Done <- query:
	default:
		log.Println("discarding Call reply due to insufficient Done chan capacity")
	}
}

type Client struct {
	mutex   sync.Mutex
	codec   ClientCodec
	pending map[uint64]*Query
	seq     uint64
	closing bool
}

func (client *Client) SendRequest(req *Request, data any) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	if err := client.codec.WriteRequest(req, data); err != nil {
		return err
	}
	return nil
}

func (client *Client) InitCodec(conn io.ReadWriteCloser) error {
	client.codec = ClientCodec{conn: conn, decoder: &Decoder{&scanner.Scanner{}}, encoder: &Encoder{&bytes.Buffer{}}}
	client.codec.decoder.s.Init(conn)
	return nil
}

func (client *Client) Listen() {
	var err error
	for {
		resp := Response{}
		err = client.codec.ReadResponseHeader(&resp)
		if err != nil {
			break
		}
		client.mutex.Lock()
		query := client.pending[resp.Seq]
		delete(client.pending, resp.Seq)
		client.mutex.Unlock()
		if query == nil {
			client.codec.ReadResponseBody(nil)
			continue
		}
		if resp.Error != "" {
			client.codec.ReadResponseBody(nil)
			query.Error = errors.New(resp.Error)
			query.done()
			continue
		}
		data := Data{}
		data.Reply = query.Reply
		err = client.codec.ReadResponseBody(&data)
		if err != nil {
			break
		}
		query.Reply = data.Reply
		query.done()
	}
	client.mutex.Lock()
	client.closing = true
	for _, query := range client.pending {
		query.Error = err
		query.done()
	}
	client.mutex.Unlock()
}

func (client *Client) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	client.InitCodec(conn)
	go client.Listen()
	return nil
}

func (client *Client) Deal(query *Query) {
	client.mutex.Lock()
	if client.closing {
		query.done()
		query.Error = errors.New("the connection is shut down")
		return
	}
	client.seq++
	client.pending[client.seq] = query
	query.seq = client.seq
	req := Request{MethodName: query.Method, Seq: client.seq}
	client.mutex.Unlock()
	client.SendRequest(&req, query.Args)
}

func (client *Client) Call(name string, args any, reply any) error {
	query := Query{Method: name, Args: args, Reply: reply, Error: nil, Done: make(chan *Query, 1)}
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()
	go client.Deal(&query)
	cnt := 0
	select {
	case <-query.Done:
		break
	case <-tick.C:
		client.mutex.Lock()
		delete(client.pending, query.seq)
		client.mutex.Unlock()
		if cnt++; cnt >= 1 {
			query.Error = errors.New("can not receive response")
			break
		}
		query = Query{Method: name, Args: args, Reply: reply, Error: nil, Done: make(chan *Query, 1)}
		go client.Deal(&query)
	}
	return query.Error
}

func (client *Client) Close() error {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	if client.closing {
		return errors.New("the connect is shut down")
	}
	client.closing = true
	return client.codec.conn.Close()
}

func GetClient() *Client {
	client := Client{}
	client.pending = make(map[uint64]*Query)
	return &client
}
