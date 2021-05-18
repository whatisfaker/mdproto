package mdproto

import (
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/unicode"
)

var (
	ErrUnsupportedStruct    = errors.New("struct is not supported")
	ErrDuplicateDefinition  = errors.New("duplicate item id in struct")
	ErrUnsupportStrEncoding = errors.New("unsupport string encoding")
	DefaultTagName          = "mdp"
)

const (
	strEncodingUTF8  = "utf8"
	strEncodingUTF16 = "utf16"
)

func getTag(name string, tag reflect.StructTag) (byte, string, bool, error) {
	value, ok := tag.Lookup(DefaultTagName)
	if !ok {
		return 0, "", false, nil
	}
	value = strings.TrimSpace(value)
	if value == "-" {
		return 0, "", false, nil
	}
	values := strings.Split(value, ",")
	l2 := len(values)
	ext := ""
	omitempty := false
	var ib byte
	if l2 > 0 {
		itemID, err := strconv.Atoi(values[0])
		if err != nil {
			return 0, "", false, fmt.Errorf("field %s mdp tag(%s) is incorrect", name, value)
		}
		ib = byte(itemID)
	}
	if l2 > 1 {
		v := strings.TrimSpace(strings.ToLower(values[1]))
		if v == "omitempty" {
			omitempty = true
		} else if v == strEncodingUTF8 || v == strEncodingUTF16 {
			ext = v
		}
	}
	if l2 > 2 {
		v := strings.TrimSpace(strings.ToLower(values[2]))
		if v == "omitempty" {
			omitempty = true
		}
	}
	return ib, ext, omitempty, nil
}

func Encode(v interface{}) ([]byte, error) {
	s := reflect.ValueOf(v)
	switch s.Type().Kind() {
	case reflect.Struct:
	case reflect.Ptr:
		s = s.Elem()
		if s.Type().Kind() != reflect.Struct {
			return nil, ErrUnsupportedStruct
		}
	default:
		return nil, ErrUnsupportedStruct
	}
	l := s.NumField()
	var ret []byte
	uniqueItems := make(map[byte]bool)
	for i := 0; i < l; i++ {
		itemID, ext, omit, err := getTag(s.Type().Field(i).Name, s.Type().Field(i).Tag)
		if err != nil {
			return nil, err
		}
		if itemID == 0 {
			continue
		}
		if _, ok := uniqueItems[itemID]; ok {
			return nil, ErrDuplicateDefinition
		}
		uniqueItems[itemID] = true
		b := []byte{itemID}
		f := s.Field(i).Interface()
		switch rs := f.(type) {
		case int8:
			if rs == 0 && omit {
				continue
			}
			b = append(b, (byte)(rs))
		case int16:
			if rs == 0 && omit {
				continue
			}
			bb := make([]byte, 2)
			binary.BigEndian.PutUint16(bb, uint16(rs))
			b = append(b, bb...)
		case int32:
			if rs == 0 && omit {
				continue
			}
			bb := make([]byte, 4)
			binary.BigEndian.PutUint32(bb, uint32(rs))
			b = append(b, bb...)
		case int64:
			if rs == 0 && omit {
				continue
			}
			bb := make([]byte, 8)
			binary.BigEndian.PutUint64(bb, uint64(rs))
			b = append(b, bb...)
		case []int8:
			if rs == nil && omit {
				continue
			}
			bb := make([]byte, 2)
			binary.BigEndian.PutUint16(bb, uint16(len(rs)))
			b = append(b, bb...)
			for _, ch := range rs {
				b = append(b, (byte)(ch))
			}
		case []int16:
			if rs == nil && omit {
				continue
			}
			bb := make([]byte, 2+len(rs)*2)
			binary.BigEndian.PutUint16(bb[0:2], uint16(len(rs)))
			for i, ch := range rs {
				binary.BigEndian.PutUint16(bb[2+i*2:i*2+4], uint16(ch))
			}
			b = append(b, bb...)
		case []int32:
			if rs == nil && omit {
				continue
			}
			bb := make([]byte, 2+len(rs)*4)
			binary.BigEndian.PutUint16(bb[0:2], uint16(len(rs)))
			for i, ch := range rs {
				binary.BigEndian.PutUint32(bb[2+i*4:i*4+6], uint32(ch))
			}
			b = append(b, bb...)
		case []int64:
			if rs == nil && omit {
				continue
			}
			bb := make([]byte, 2+len(rs)*8)
			binary.BigEndian.PutUint16(bb[0:2], uint16(len(rs)))
			for i, ch := range rs {
				binary.BigEndian.PutUint64(bb[2+i*8:i*8+10], uint64(ch))
			}
			b = append(b, bb...)
		case string:
			if rs == "" && omit {
				continue
			}
			var bts []byte
			if ext != "" && ext != strEncodingUTF8 && ext != strEncodingUTF16 {
				return nil, ErrUnsupportStrEncoding
			}
			if ext == strEncodingUTF8 {
				bts = []byte(rs)
			} else {
				var err error
				encoder := unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewEncoder()
				bts, err = encoder.Bytes([]byte(rs))
				if err != nil {
					return nil, fmt.Errorf("convert string to []bytes(UTF16) error %v", err)
				}
			}
			length := len(bts)
			bb := make([]byte, 2)
			binary.BigEndian.PutUint16(bb, uint16(length))
			b = append(b, bb...)
			b = append(b, bts...)
		default:
			return nil, fmt.Errorf("unsupport value type %s", s.Field(i).Type().Name())
		}
		ret = append(ret, b...)
	}
	return ret, nil
}

type val struct {
	Value reflect.Value
	Ext   string
}

func Decode(b []byte, v interface{}) error {
	s := reflect.ValueOf(v)
	switch s.Type().Kind() {
	case reflect.Ptr:
		s = s.Elem()
		if s.Type().Kind() != reflect.Struct {
			return ErrUnsupportedStruct
		}
	default:
		return ErrUnsupportedStruct
	}
	l := s.NumField()
	md := make(map[byte]*val)
	for i := 0; i < l; i++ {
		itemID, encoding, _, err := getTag(s.Type().Field(i).Name, s.Type().Field(i).Tag)
		if err != nil {
			return err
		}
		if itemID == 0 {
			continue
		}
		fd := s.Field(i)
		if _, ok := md[itemID]; ok {
			return ErrDuplicateDefinition
		}
		md[itemID] = &val{
			Value: fd,
			Ext:   encoding,
		}
	}
	start := 0
	end := len(b)
	for start < end {
		offset, err := split(b[start:], md)
		if err != nil {
			return err
		}
		start += offset
	}
	return nil
}

func split(data []byte, dict map[byte]*val) (int, error) {
	itemID := data[0]
	val, ok := dict[itemID]
	if !ok {
		return 1, nil
	}
	begin, end := 1, 1
	switch val.Value.Interface().(type) {
	case int8:
		end += 1
		val.Value.SetInt(int64(data[begin]))
	case int16:
		end += 2
		v := binary.BigEndian.Uint16(data[begin:end])
		val.Value.SetInt(int64(v))
	case int32:
		end += 4
		v := binary.BigEndian.Uint32(data[begin:end])
		val.Value.SetInt(int64(v))
	case int64:
		end += 8
		v := binary.BigEndian.Uint64(data[begin:end])
		val.Value.SetInt(int64(v))
	case []int8:
		end += 2
		l := binary.BigEndian.Uint16(data[begin:end])
		v := make([]int8, l)
		for i := 0; i < int(l); i++ {
			v[i] = (int8)(data[end])
			end += 1
		}
		val.Value.Set(reflect.ValueOf(v))
	case []int16:
		end += 2
		l := binary.BigEndian.Uint16(data[begin:end])
		v := make([]int16, l)
		for i := 0; i < int(l); i++ {
			vv := binary.BigEndian.Uint16(data[end : end+2])
			v[i] = int16(vv)
			end += 2
		}
		val.Value.Set(reflect.ValueOf(v))
	case []int32:
		end += 2
		l := binary.BigEndian.Uint16(data[begin:end])
		v := make([]int32, l)
		for i := 0; i < int(l); i++ {
			vv := binary.BigEndian.Uint32(data[end : end+4])
			v[i] = int32(vv)
			end += 4
		}
		val.Value.Set(reflect.ValueOf(v))
	case []int64:
		end += 2
		l := binary.BigEndian.Uint16(data[begin:end])
		v := make([]int64, l)
		for i := 0; i < int(l); i++ {
			vv := binary.BigEndian.Uint64(data[end : end+8])
			v[i] = int64(vv)
			end += 8
		}
		val.Value.Set(reflect.ValueOf(v))
	case string:
		end += 2
		l := binary.BigEndian.Uint16(data[begin:end])
		if val.Ext != "" && val.Ext != strEncodingUTF8 && val.Ext != strEncodingUTF16 {
			return 0, ErrUnsupportStrEncoding
		}
		if val.Ext == strEncodingUTF8 {
			val.Value.SetString(string(data[end : end+int(l)]))
		} else {
			var err error
			decoder := unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder()
			bts, err := decoder.Bytes(data[end : end+int(l)])
			if err != nil {
				return 0, fmt.Errorf("decode []byte(UTF16) to string error %v", err)
			}
			val.Value.SetString(string(bts))
		}
		end += int(l)
	default:
		return 0, fmt.Errorf("unsupport value type %v", val.Value.Type())
	}
	return end, nil
}
