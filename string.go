package direct

import (
	"fmt"
	"github.com/madokast/direct/memory"
	"github.com/madokast/direct/utils"
	"reflect"
	"strings"
	"sync/atomic"
	"unsafe"
)

// String represents an un-modifiable string in managed_memory.
type String struct {
	ptr    memory.Pointer
	length SizeType
	holder Slice[byte]
}

func (s *String) Move() (moved String) {
	moved = *s
	*s = emptyString
	return moved
}

func (s String) Length() SizeType {
	return s.length
}

func (s String) AsGoString() string {
	return *((*string)(unsafe.Pointer(&s)))
}

func (s String) CopyToGoString() string {
	var sb strings.Builder
	sb.WriteString(s.AsGoString())
	return sb.String()
}

func (s String) Equal(s2 String) bool {
	return s.AsGoString() == s2.AsGoString()
}

func isString[T any]() bool {
	return reflect.TypeOf((*T)(nil)).Elem() == stringType
}

func equalString[str any](s1, s2 str) bool {
	if utils.Asserted {
		if !isString[str]() {
			panic(fmt.Sprintf("call equalString using non-string type s1=%v, s2=%v, type(s2)=%T", s1, s2, fmt.Sprintf("%T", s1)))
		}
	}
	return *((*string)(unsafe.Pointer(&s1))) == *((*string)(unsafe.Pointer(&s2)))
}

func (s String) Hashcode() (hash SizeType) {
	hash = 2166136261
	const prime32 = 16777619
	i := SizeType(0)
	for ; i < s.length; i++ {
		hash *= prime32
		//logger.Debug("at", fmt.Sprintf("%c", *PointerAs[byte](s.ptr)))
		hash ^= SizeType(*memory.PointerAs[byte](s.ptr))
		s.ptr++
	}
	return hash
}

func hashString[str any](s str) SizeType {
	if utils.Asserted {
		if !isString[str]() {
			panic(fmt.Sprintf("call hashString using non-string type s=%v, type(s)=%T", s, fmt.Sprintf("%T", s)))
		}
	}
	return ((*String)(unsafe.Pointer(&s))).Hashcode()
}

func (s String) String() string {
	return s.CopyToGoString()
}

func (s String) Free() {
	if s.holder.pointer().IsNotNull() {
		header := s.holder.header()
		cnt := atomic.AddInt32(memory.PointerAs[int32](header.elementBasePointer), -1)
		if utils.Asserted {
			if cnt < 0 {
				panic(fmt.Sprintf("string holder cnt <(%d) 0", cnt))
			}
		}
		if utils.Debug {
			fmt.Printf("free string %s in holder count %d\n", s.AsGoString(), cnt)
		}
		if cnt == 0 {
			s.holder.Free()
		}
	}
}

func (s *String) Nove() (moved String) {
	moved = *s
	*s = emptyString
	return moved
}

func (s String) Moved() bool {
	return s == emptyString
}

var emptyString = String{}
var stringType = reflect.TypeOf(emptyString)

func init() {
	if memory.Sizeof[String]()-memory.Sizeof[Slice[byte]]() != memory.Sizeof[string]() {
		panic(fmt.Sprintf("string size is not correct %d %d", memory.Sizeof[String](), memory.Sizeof[memory.Word]()))
	}
}
