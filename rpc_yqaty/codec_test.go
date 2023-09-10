package rpc_yqaty

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

type Func struct {
}

type Struct1 struct {
	A int
	B int
}

type Struct2 struct {
	Mp map[int]int
}

type Struct3 struct {
	A int
	B string
	C uint
	D bool
}

type Struct4 struct {
	A  *Struct3
	B  Struct3
	Mp map[int]string
	Ar []int
	As []string
	//a  int
}

func TestMarshal(t *testing.T) {

	{
		A := 1
		if s, err := Marshal(A); s != "1" || err != nil {
			if err != nil {
				t.Error(err.Error())
			} else {
				t.Errorf("Marshal: convert %v into json format is %s !", A, s)
			}
		}
	}

	{
		A := "ada"
		if s, err := Marshal(A); s != "\"ada\"" || err != nil {
			if err != nil {
				t.Error(err.Error())
			} else {
				t.Errorf("Marshal: convert %v into json format is %s !", A, s)
			}
		}

	}

	{
		A := []int{1, 2, 3}
		if s, err := Marshal(A); s != "[1,2,3]" || err != nil {
			if err != nil {
				t.Error(err.Error())
			} else {
				t.Errorf("Marshal: convert %v into json format is %s !", A, s)
			}
		}
	}

	// {
	// 	A := map[int]string{1: "a", 2: "b", 3: "c"}
	// 	if s, err := Marshal(A); s != "{\"1\":\"a\",\"2\":\"b\",\"3\":\"c\"}" || err != nil {
	// 		if err != nil {
	// 			t.Error(err.Error())
	// 		} else {
	// 			t.Errorf("Marshal: convert %v into json format is %s !", A, s)
	// 		}
	// 	}
	// }

	{
		A := Struct1{1, 2}
		if s, err := Marshal(A); s != "{\"A\":1,\"B\":2}" || err != nil {
			if err != nil {
				t.Error(err.Error())
			} else {
				t.Errorf("Marshal: convert %v into json format is %s !", A, s)
			}
		}
	}

	// {
	// 	A := Struct2{map[int]int{1: 1, 2: 2, 3: 3}}
	// 	if s, err := Marshal(A); s != "{\"Mp\":{\"1\":1,\"2\":2,\"3\":3}}" || err != nil {
	// 		if err != nil {
	// 			t.Error(err.Error())
	// 		} else {
	// 			t.Errorf("Marshal: convert %v into json format is %s !", A, s)
	// 		}
	// 	}
	// }

	{
		A := Struct3{3, "ada", 5, true}
		if s, err := Marshal(A); s != "{\"A\":3,\"B\":\"ada\",\"C\":5,\"D\":true}" || err != nil {
			if err != nil {
				t.Error(err.Error())
			} else {
				t.Errorf("Marshal: convert %v into json format is %s !", A, s)
			}
		}
	}

	// {
	// 	A := Struct4{A: &Struct3{A: 1, B: "dsds", C: 3, D: true}, B: Struct3{A: 1, B: "dsds", C: 3, D: true}, Mp: map[int]string{1: "adas", 2: "sdsf"}, Ar: []int{1, 2, 3}, As: []string{"1", "2", "3"}}
	// 	if s, err := Marshal(A); s != "{\"A\":{\"A\":1,\"B\":\"dsds\",\"C\":3,\"D\":true},\"B\":{\"A\":1,\"B\":\"dsds\",\"C\":3,\"D\":true},\"Mp\":{\"1\":\"adas\",\"2\":\"sdsf\"},\"Ar\":[1,2,3],\"As\":[\"1\",\"2\",\"3\"]}" || err != nil {
	// 		if err != nil {
	// 			t.Error(err.Error())
	// 		} else {
	// 			t.Errorf("Marshal: convert %v into json format is %s !", A, s)
	// 		}
	// 	}
	// }

	{
		A := Struct2{}
		if s, err := Marshal(A); s != "{\"Mp\":null}" || err != nil {
			if err != nil {
				t.Error(err.Error())
			} else {
				t.Errorf("Marshal: convert %v into json format is %s !", A, s)
			}
		}
	}
}

func TestUnMarshal(t *testing.T) {

	{
		s := "1"
		A := new(int)
		if err := UnMarshal(s, A); fmt.Sprintf("%v", *A) != "1" || err != nil {
			if err != nil {
				t.Error(err.Error())
			} else {
				t.Errorf("UnMarshal: convert json format %s into object is %v !", s, *A)
			}
		}
	}

	{
		s := "\"ada\""
		A := new(string)
		if err := UnMarshal(s, A); fmt.Sprintf("%v", *A) != "ada" || err != nil {
			if err != nil {
				t.Error(err.Error())
			} else {
				t.Errorf("UnMarshal: convert json format %s into object is %v !", s, *A)
			}
		}
	}

	{
		s := "[1,2,3]"
		A := make([]int, 3)
		if err := UnMarshal(s, &A); fmt.Sprintf("%v", A) != "[1 2 3]" || err != nil {
			if err != nil {
				t.Error(err.Error())
			} else {
				t.Errorf("UnMarshal: convert json format %s into object is %v !", s, A)
			}
		}
	}

	// {
	// 	A := map[int]string{1: "a", 2: "b", 3: "c"}
	// 	if s, err := Marshal(A); s != "{\"1\":\"a\",\"2\":\"b\",\"3\":\"c\"}" || err != nil {
	// 		if err != nil {
	// 			t.Error(err.Error())
	// 		} else {
	// 			t.Errorf("Marshal: convert %v into json format is %s !", A, s)
	// 		}
	// 	}
	// }

	{
		s := "{\"A\":1,\"B\":2}"
		A := &Struct1{}
		if err := UnMarshal(s, A); fmt.Sprintf("%v", *A) != "{1 2}" || err != nil {
			if err != nil {
				t.Error(err.Error())
			} else {
				t.Errorf("UnMarshal: convert json format %s into object is %v !", s, *A)
			}
		}
	}

	{
		s := "{\"Mp\":{\"1\":1,\"2\":2,\"3\":3}}"
		A := &Struct2{}
		A.Mp = make(map[int]int)
		if err := UnMarshal(s, A); fmt.Sprintf("%v", *A) != "{map[1:1 2:2 3:3]}" || err != nil {
			if err != nil {
				t.Error(err.Error())
			} else {
				t.Errorf("UnMarshal: convert json format %s into object is %v !", s, *A)
			}
		}
	}

	{
		s := "{\"A\":3,\"B\":\"ada\",\"C\":5,\"D\":true}"
		A := &Struct3{3, "ada", 5, true}
		if err := UnMarshal(s, A); fmt.Sprintf("%v", *A) != "{3 ada 5 true}" || err != nil {
			if err != nil {
				t.Error(err.Error())
			} else {
				t.Errorf("UnMarshal: convert json format %s into object is %v !", s, *A)
			}
		}
	}

	// {
	// 	s := "{\"A\":{\"A\":1,\"B\":\"dsds\",\"C\":3,\"D\":true},\"B\":{\"A\":1,\"B\":\"dsds\",\"C\":3,\"D\":true},\"Mp\":{\"1\":\"adas\",\"2\":\"sdsf\"},\"Ar\":[1,2,3],\"As\":[\"1\",\"2\",\"3\"]}"
	// 	A := &Struct4{A: &Struct3{}}
	// 	A.Mp = make(map[int]string)
	// 	A.As = make([]string, 3)
	// 	A.Ar = make([]int, 3)
	// 	if err := UnMarshal(s, A); fmt.Sprintf("%v", *A) != "{3 ada 5 true}" || err != nil {
	// 		if err != nil {
	// 			t.Error(err.Error())
	// 		} else {
	// 			t.Errorf("UnMarshal: convert json format %s into object is %v !", s, *A)
	// 		}
	// 	}
	// }

	{
		s := "{\"Mp\":null}"
		A := &Struct2{}
		//A.Mp = make(map[int]int)
		if err := UnMarshal(s, A); fmt.Sprintf("%v", *A) != "{map[]}" || err != nil {
			if err != nil {
				t.Error(err.Error())
			} else {
				t.Errorf("UnMarshal: convert json format %s into object is %v !", s, *A)
			}
		}
	}
}

func (f *Func) Add(A Struct1, B *int) error {
	*B = A.A + A.B
	return nil
}

func (f *Func) Map(A Struct1, B *Struct2) error {
	mp := make(map[int]int)
	mp[A.A] = A.B
	B.Mp = mp
	return nil
}

func (f *Func) Get(A Struct1, B *Struct4) error {
	*B = Struct4{A: &Struct3{A: 1, B: "dsds", C: 3, D: true}, B: Struct3{A: 1, B: "dsds", C: 3, D: true}, Mp: map[int]string{1: "adas", 2: "sdsf"}, Ar: []int{1, 2, 3}, As: []string{"1", "2", "3"}}
	return nil
}

func ServerTest() {
	server := GetServer()
	server.Register("add", (*Func).Add)
	server.Register("map", (*Func).Map)
	server.Register("get", (*Func).Get)
	server.Accept("127.0.0.1:9090")
}

func call(t *testing.T, ans any, client *Client, name string, args any, reply any, wg *sync.WaitGroup) {
	t.Helper()
	defer wg.Done()
	err := client.Call(name, args, reply)
	if err != nil {
		//fmt.Println("Error: ", name, args, err)
		t.Errorf("Error: %v %v %v", name, args, err)
		return
	}
	if fmt.Sprintf("%v", ans) != fmt.Sprintf("%v", reflect.ValueOf(reply).Elem()) {
		t.Errorf("expect %v, output %v", ans, reflect.ValueOf(reply).Elem())
	}
	//fmt.Println("Successfully: ", name, args, reflect.ValueOf(reply).Elem())
}

func TestAll(t *testing.T) {

	go ServerTest()
	time.Sleep(time.Second)
	client := GetClient()
	err := client.Dial("127.0.0.1:9090")
	if err != nil {
		fmt.Println(err)
		return
	}
	wg := new(sync.WaitGroup)

	wg.Add(1)
	test1_A := new(int)
	ans1 := 3
	go call(t, ans1, client, "add", Struct1{1, 2}, test1_A, wg)

	wg.Add(1)
	test2_A := &Struct2{}
	test2_A.Mp = make(map[int]int)
	ans2 := Struct2{map[int]int{1: 2}}
	go call(t, ans2, client, "map", Struct1{1, 2}, test2_A, wg)

	/*wg.Add(1)
	test3_A := &Struct4{A: &Struct3{}}
	test3_A.Mp = make(map[int]string)
	test3_A.Ar = make([]int, 3)
	test3_A.As = make([]string, 3)
	go call(t, client, "get", Struct1{1, 2}, test3_A, wg)*/

	wg.Wait()
	//fmt.Println(*test3_A.A)

	err = client.Close()
	if err != nil {
		fmt.Println(err)
	}
}
