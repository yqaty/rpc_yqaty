# rpc_yqaty

A rpc framework written in go.

---

## Usage

### Server

```go

type Func struct{}

type Struct1 struct{
    A int
    B int
}

func (f *Func) Add(A Struct1, B *int) error {
    *B = A.A + A.B
    return nil
}

func main(){
    server := GetServer()
    server.Register("add", (*Func).Add)
    server.Accept("127.0.0.1:9090") // tcp
}
```
### Client

```go

func main(){
    client := GetClient()
    client.Dial("127.0.0.1:9090")
    A := new(int)
    client.Call("add", Struct1{1, 2}, A)
    client.Close()
}

```
