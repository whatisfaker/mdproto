package mdproto

import (
	"encoding/json"
	"fmt"
	"testing"
)

type AAA struct {
	F1  int8    `mdp:"100,omitempty"`
	F2  int16   `mdp:"101"`
	F3  int32   `mdp:"102"`
	F4  int64   `mdp:"103"`
	F11 []int8  `mdp:"200,omitempty"`
	F22 []int16 `mdp:"201"`
	F33 []int32 `mdp:"202"`
	F44 []int64 `mdp:"203"`
	FS  string  `mdp:"251,utf8,omitempty"`
	FS2 string  `mdp:"252"`
}

type TestObject struct {
	Result    int8   `mdp:"60"`
	UserType  int8   `mdp:"61"`
	IsActive  int8   `mdp:"62"`
	IP        string `mdp:"130"`
	Key       string `mdp:"131"`
	UserID    int32  `mdp:"132"`
	Port      int32  `mdp:"133"`
	ServerID  int32  `mdp:"134"`
	Token     int8   `mdp:"135"`
	F1        int32  `mdp:"136"`
	F2        int32  `mdp:"137"`
	UserType2 int8   `mdp:"138"`
}

func Test2(t *testing.T) {
	b := []int8{60, 0, 61, 12, 62, 1, -126, 0, 20, -2, -1, 0, 49, 0, 50, 0, 55, 0, 46, 0, 48, 0, 46, 0, 48, 0, 46, 0, 49, -125, 0, 66, -2, -1, 0, 100, 0, 57, 0, 50, 0, 52, 0, 98, 0, 99, 0, 49, 0, 99, 0, 52, 0, 49, 0, 50, 0, 52, 0, 52, 0, 52, 0, 51, 0, 56, 0, 97, 0, 49, 0, 101, 0, 52, 0, 50, 0, 101, 0, 100, 0, 97, 0, 53, 0, 55, 0, 98, 0, 52, 0, 97, 0, 49, 0, 57, 0, 53, -124, 0, 15, 80, 62, -123, 0, 0, 31, 120, -122, 1, 49, 84, 17, -121, 0, 66, -2, -1, 0, 55, 0, 102, 0, 100, 0, 49, 0, 49, 0, 98, 0, 50, 0, 50, 0, 53, 0, 56, 0, 101, 0, 56, 0, 52, 0, 101, 0, 50, 0, 54, 0, 57, 0, 48, 0, 51, 0, 100, 0, 53, 0, 51, 0, 49, 0, 51, 0, 53, 0, 49, 0, 53, 0, 102, 0, 49, 0, 57, 0, 100, 0, 101, -120, 0, 0, 0, 0, -119, 0, 0, 0, 0, -118, 1}
	bb := make([]byte, 0)
	for _, k := range b {
		bb = append(bb, byte(k))
	}
	fmt.Println(bb)
	var o TestObject
	err := Unmarshal(bb, &o)
	if err != nil {
		t.Error(err)
		return
	}
	c, _ := json.Marshal(o)
	t.Log(string(c))
}

func TestEncode(t *testing.T) {
	// bbs := []byte("测试1")
	// for _, bb := range bbs {
	// 	fmt.Printf("%x ", bb)
	// }
	// fmt.Println()
	// d := unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewEncoder()
	// bbs, _ = d.Bytes(bbs)
	// for _, bb := range bbs {
	// 	fmt.Printf("%x ", bb)
	// }
	// fmt.Println()
	b, err := Marshal(&AAA{
		F1:  1,
		F2:  32767,
		F3:  65535,
		F4:  100000000,
		F22: []int16{11, 12},
		F33: []int32{21, 22, 23},
		F44: []int64{31, 32, 33, 34},
		FS:  "测试1",
		FS2: "测试1",
	})
	if err != nil {
		t.Error(err)
		return
	}
	var c AAA
	err = Unmarshal(b, &c)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(b, c)
}
