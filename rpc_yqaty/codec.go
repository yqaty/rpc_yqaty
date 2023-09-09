package rpc_yqaty

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/scanner"
)

type Encoder struct {
	s *bytes.Buffer
}

func (codec *Encoder) JSONEncode(data any) error {
	return codec.encode(reflect.ValueOf(data))
}

func (codec *Encoder) encode(data reflect.Value) error {
	if !data.CanInterface() {
		return nil
	}
	switch data.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprintf(codec.s, "%d", data.Int())
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fmt.Fprintf(codec.s, "%d", data.Uint())
		return nil

	case reflect.Float32, reflect.Float64:
		fmt.Fprintf(codec.s, "%f", data.Float())
		return nil

	case reflect.Bool:
		fmt.Fprintf(codec.s, "%t", data.Bool())
		return nil

	case reflect.Array, reflect.Slice:
		if data.IsZero() {
			codec.s.WriteString("null")
			return nil
		}
		codec.s.WriteString("[")
		for i := 0; i < data.Len(); i++ {
			if i > 0 {
				codec.s.WriteString(",")
			}
			if err := codec.encode(data.Index(i)); err != nil {
				return err
			}
		}
		codec.s.WriteString("]")
		return nil

	case reflect.Map:
		if data.IsZero() {
			codec.s.WriteString("null")
			return nil
		}
		keys := data.MapKeys()
		codec.s.WriteString("{")
		for i, key := range keys {
			if i > 0 {
				codec.s.WriteString(",")
			}
			if key.Kind() != reflect.String {
				codec.s.WriteString("\"")
			}
			if err := codec.encode(key); err != nil {
				return err
			}
			if key.Kind() != reflect.String {
				codec.s.WriteString("\"")
			}
			codec.s.WriteString(":")
			if err := codec.encode(data.MapIndex(key)); err != nil {
				return err
			}
		}
		codec.s.WriteString("}")
		return nil

	case reflect.String:
		if data.IsZero() {
			codec.s.WriteString("null")
			return nil
		}
		fmt.Fprintf(codec.s, "\"%s\"", data.String())
		return nil

	case reflect.Pointer, reflect.Interface:
		if data.IsZero() {
			codec.s.WriteString("null")
			return nil
		}
		if err := codec.encode(data.Elem()); err != nil {
			return err
		}
		return nil

	case reflect.Struct:
		codec.s.WriteString("{")
		flag := true
		for i := 0; i < data.NumField(); i++ {
			if !data.Field(i).CanInterface() {
				continue
			}
			if !flag {
				codec.s.WriteString(",")
			}
			flag = false
			fmt.Fprintf(codec.s, "\"%s\":", data.Type().Field(i).Name)
			if err := codec.encode(data.Field(i)); err != nil {
				return err
			}
		}
		codec.s.WriteString("}")
		return nil

	default:
		return fmt.Errorf("unsupported type: %v", data.Kind())
	}
}

type Decoder struct {
	s *scanner.Scanner
}

func (codec *Decoder) consume(s string) error {
	token := codec.s.Scan()
	if token == scanner.EOF {
		return errors.New("decode failed")
	}
	str := codec.s.TokenText()
	if str != s {
		return errors.New("decode failed")
	}
	return nil
}

func (codec *Decoder) JSONDecode(data any) error {
	vdata := reflect.ValueOf(data)
	if vdata.Kind() != reflect.Pointer || vdata.IsNil() {
		return errors.New("parameter must be a vaild pointer")
	}
	err := codec.decode(vdata)
	return err
}

func (codec *Decoder) Read() (string, error) {
	token := codec.s.Scan()
	if token == scanner.EOF {
		return "", errors.New("decode failed")
	}
	str := codec.s.TokenText()
	return str, nil
}

func (codec *Decoder) decode(data reflect.Value) error {
	if !data.CanInterface() {
		return nil
	}
	switch data.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str, err := codec.Read()
		if err != nil {
			return err
		}
		if str == "null" {
			return nil
		}
		if str[0] == '"' {
			str = str[1 : len(str)-1]
		}
		i, err := strconv.Atoi(str)
		if err != nil {
			return err
		}
		data.SetInt(int64(i))
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		str, err := codec.Read()
		if err != nil {
			return err
		}
		if str == "null" {
			return nil
		}
		if str[0] == '"' {
			str = str[1 : len(str)-1]
		}
		i, err := strconv.Atoi(str)
		if err != nil {
			return err
		}
		data.SetUint(uint64(i))
		return nil

	case reflect.Float32, reflect.Float64:
		str, err := codec.Read()
		if err != nil {
			return err
		}
		if str == "null" {
			return nil
		}
		if str[0] == '"' {
			str = str[1 : len(str)-1]
		}
		f, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
		data.SetFloat(float64(f))
		return nil

	case reflect.Bool:
		str, err := codec.Read()
		if err != nil {
			return err
		}
		if str == "null" {
			return nil
		}
		if str[0] == '"' {
			str = str[1 : len(str)-1]
		}
		b, err := strconv.ParseBool(str)
		if err != nil {
			return err
		}
		data.SetBool(b)
		return nil

	case reflect.Array:
		str, err := codec.Read()
		if err != nil {
			return err
		}
		if str == "null" {
			return nil
		}
		if str != "[" {
			return errors.New("decode failed")
		}
		for i := 0; i < data.Len(); i++ {
			if i > 0 {
				if err := codec.consume(","); err != nil {
					return err
				}
			}
			if err := codec.decode(data.Index(i)); err != nil {
				return err
			}
		}
		if err := codec.consume("]"); err != nil {
			return err
		}
		return nil

	case reflect.Slice:
		str, err := codec.Read()
		if err != nil {
			return err
		}
		if str == "null" {
			return nil
		}
		if str != "[" {
			return errors.New("decode failed")
		}
		cnt := 0
		for str != "]" {
			val := reflect.New(data.Type().Elem()).Elem()
			err := codec.decode(val)
			if err != nil {
				return err
			}
			data.Index(cnt).Set(val)
			str, err = codec.Read()
			if err != nil {
				return err
			}
			if str != "," && str != "]" {
				return errors.New("decode failed")
			}
			cnt++
		}
		return nil

	case reflect.Map:
		str, err := codec.Read()
		if err != nil {
			return err
		}
		if str == "null" {
			return nil
		}
		if str != "{" {
			return errors.New("decode failed")
		}
		for str != "}" {
			key := reflect.New(data.Type().Key()).Elem()
			err := codec.decode(key)
			if err != nil {
				return err
			}
			err = codec.consume(":")
			if err != nil {
				return err
			}
			val := reflect.New(data.Type().Elem()).Elem()
			err = codec.decode(val)
			if err != nil {
				return err
			}
			data.SetMapIndex(key, val)
			str, err = codec.Read()
			if err != nil {
				return err
			}
		}
		return nil

	case reflect.String:
		str, err := codec.Read()
		if err != nil {
			return err
		}
		if str == "null" {
			return nil
		}
		data.SetString(str[1 : len(str)-1])
		return nil

	case reflect.Pointer:
		if err := codec.decode(data.Elem()); err != nil {
			return err
		}
		return nil

	case reflect.Struct:
		str, err := codec.Read()
		if err != nil {
			return err
		}
		if str == "null" {
			return nil
		}
		if str != "{" {
			return errors.New("decode failed")
		}
		for str != "}" {
			name, err := codec.Read()
			if err != nil {
				return err
			}
			name = name[1 : len(name)-1]
			err = codec.consume(":")
			if err != nil {
				return err
			}
			_, bo := data.Type().FieldByName(name)
			if bo {
				if err := codec.decode(data.FieldByName(name)); err != nil {
					return err
				}
			}
			str, err = codec.Read()
			if err != nil {
				return err
			}
		}
		return nil

	case reflect.Interface:
		err := codec.decode(data.Elem())
		if err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("unsupported type: %v", data.Kind())
	}
}

func Marshal(arg any) (string, error) {
	codec := &Encoder{new(bytes.Buffer)}
	err := codec.JSONEncode(arg)
	return codec.s.String(), err
}

func UnMarshal(s string, rec any) error {
	codec := &Decoder{new(scanner.Scanner)}
	codec.s.Init(strings.NewReader(s))
	return codec.JSONDecode(rec)
}
