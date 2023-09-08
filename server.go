package rpc_yqaty

import (
	"bytes"
	"errors"
	"fmt"
	"go/token"
	"io"
	"net"
	"reflect"
	"sync"
	"text/scanner"
)

type MethodType struct {
	Method    reflect.Type
	ArgsType  reflect.Type
	ReplyType reflect.Type
	Value     reflect.Value
}

type Request struct {
	MethodName string
	Seq        uint64
}

type Response struct {
	Seq   uint64
	Error string
}

type Data struct {
	Reply any
}

type ServerCodec struct {
	conn    io.ReadWriteCloser
	encoder *Encoder
	decoder *Decoder
}

func (scodec *ServerCodec) ReadRequestHeader(req *Request) error {
	return scodec.decoder.JSONDecode(req)
}

func (scodec *ServerCodec) ReadRequestBody(data any) error {
	fmt.Println(reflect.TypeOf(data))
	return scodec.decoder.JSONDecode(data)
}

func (scodec *ServerCodec) WirteResponse(resp *Response, data *Data) error {
	scodec.encoder.s = new(bytes.Buffer)
	//fmt.Println(resp, reflect.TypeOf(resp.Reply))
	if err := scodec.encoder.JSONEncode(resp); err != nil {
		return err
	}
	if err := scodec.encoder.JSONEncode(data); err != nil {
		return err
	}
	scodec.encoder.s.WriteString(" ")
	scodec.conn.Write(scodec.encoder.s.Bytes())
	return nil
}

func (scodec *ServerCodec) Close() {
	scodec.conn.Close()
}

type Server struct {
	Mp map[string]*MethodType
}

func IsExportedOrBulitinType(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return token.IsExported(t.Name()) || t.PkgPath() == ""
}

func (server *Server) Register(name string, method any) error {
	methodtype := reflect.TypeOf(method)
	if methodtype.Kind() != reflect.Func {
		return errors.New("register: the second parameter should be a method")
	}
	if methodtype.NumIn() != 3 {
		return errors.New("register: needs exactly 3 parameter")
	}
	if methodtype.In(2).Kind() != reflect.Pointer {
		return errors.New("register: reply type needs to be a pointer")
	}
	if !IsExportedOrBulitinType(methodtype.In(2)) {
		return errors.New("register: reply needs to be exported")
	}
	if methodtype.NumOut() != 1 {
		return errors.New("register: needs to return exactly 1 parameter")
	}
	if methodtype.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
		return errors.New("register: needs to return type error")
	}
	_, ok := server.Mp[name]
	if ok {
		return errors.New("register: the name has been registered")
	}
	server.Mp[name] = &MethodType{methodtype, methodtype.In(1), methodtype.In(2), reflect.ValueOf(method)}
	return nil
}

func (server *Server) SendResponse(codec ServerCodec, sending *sync.Mutex, resp *Response, data *Data) {
	sending.Lock()
	codec.WirteResponse(resp, data)
	sending.Unlock()
}

func (server *Server) DealRequest(codec ServerCodec, sending *sync.Mutex, wg *sync.WaitGroup, req *Request, args reflect.Value) {
	defer wg.Done()
	method := server.Mp[req.MethodName]
	fun := method.Value
	reply := reflect.New(method.ReplyType.Elem())
	rcvr := reflect.New(method.Method.In(0).Elem())
	fmt.Println("WE", reply)
	rerrors := fun.Call([]reflect.Value{rcvr, args, reply})
	var err error
	if rerrors[0].Interface() != nil {
		err = rerrors[0].Interface().(error)
	}
	fmt.Println(reply.Elem(), err)
	if err != nil {
		server.SendResponse(codec, sending, &Response{req.Seq, err.Error()}, &Data{nil})
		return
	}
	fmt.Println(reply)
	server.SendResponse(codec, sending, &Response{req.Seq, ""}, &Data{reply.Interface()})
}

func (server *Server) ServeConn(codec ServerCodec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		var req Request
		err := codec.ReadRequestHeader(&req)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				codec.ReadRequestBody(nil)
				server.SendResponse(codec, sending, &Response{req.Seq, err.Error()}, &Data{nil})
				continue
			}
		}
		method, ok := server.Mp[req.MethodName]
		if !ok {
			codec.ReadRequestBody(nil)
			server.SendResponse(codec, sending, &Response{req.Seq, "the name has not been register"}, &Data{nil})
			continue
		}
		args := reflect.New(method.ArgsType)
		err = codec.ReadRequestBody(args.Interface())
		fmt.Println(err)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				server.SendResponse(codec, sending, &Response{req.Seq, err.Error()}, &Data{nil})
				continue
			}
		}
		wg.Add(1)
		go server.DealRequest(codec, sending, wg, &req, args.Elem())
	}
	wg.Wait()
	codec.Close()
}

func (server *Server) InitCodec(conn io.ReadWriteCloser) {
	codec := ServerCodec{conn: conn, decoder: &Decoder{&scanner.Scanner{}}, encoder: &Encoder{&bytes.Buffer{}}}
	codec.decoder.s.Init(conn)
	server.ServeConn(codec)
}

func (server *Server) Accept(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			return err
		}
		go server.InitCodec(conn)
	}
}

func GetServer() *Server {
	server := Server{}
	server.Mp = make(map[string]*MethodType)
	return &server
}
