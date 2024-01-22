package direct

import (
	"math/rand"
	"strconv"
	"testing"
)

func TestNewStringFactory(t *testing.T) {
	memory := New(1 * MB)
	defer memory.Free()

	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	factory := NewStringFactory()
	defer factory.Destroy(&concurrentMemory)

	s1, err := factory.CreateFromGoString("hello", &concurrentMemory)
	PanicErr(err)
	defer s1.Free(&concurrentMemory)

	t.Log(s1.Length())
	t.Log(s1.AsGoString())
	t.Log(s1.CopyToGoString())
	t.Log(s1)

	Assert(s1.CopyToGoString() == "hello")
}

func TestString_Hashcode(t *testing.T) {
	memory := New(1 * MB)
	defer memory.Free()

	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	factory := NewStringFactory()
	defer factory.Destroy(&concurrentMemory)

	s1, err := factory.CreateFromGoString("hello", &concurrentMemory)
	PanicErr(err)
	defer s1.Free(&concurrentMemory)

	Assert(s1.CopyToGoString() == "hello")

	t.Log(s1.Hashcode())
}

func Test_IsString(t *testing.T) {
	Assert(isString[String]())
	Assert(!isString[string]())
}

func Benchmark_IsString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Assert(isString[String]())
	}
}

func TestStringFactory_CreateFromGoString(t *testing.T) {
	memory := New(1 * MB)
	defer memory.Free()

	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	factory := NewStringFactory()
	defer factory.Destroy(&concurrentMemory)

	m, err := MakeMap[String, SizeType](0, &concurrentMemory)
	PanicErr(err)
	defer m.Free(&concurrentMemory)

	for _, s := range []string{"hello", ",", " ", "world", "!"} {
		ss, err := factory.CreateFromGoString(s, &concurrentMemory)
		PanicErr(err)

		err = m.Put(ss, ss.Length(), &concurrentMemory)
		PanicErr(err)

		t.Logf("%v", m)
	}

	iter := m.Iterator()
	for iter.Next() {
		iter.KeyRef().Free(&concurrentMemory)
	}
}

func TestStringFactory_CreateFromGoString1000(t *testing.T) {
	memory := New(128 * MB)
	defer memory.Free()

	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	factory := NewStringFactory()
	defer factory.Destroy(&concurrentMemory)

	m, err := MakeMap[String, SizeType](0, &concurrentMemory)
	PanicErr(err)
	defer m.Free(&concurrentMemory)

	for i := 0; i < 1000; i++ {
		s := strconv.Itoa(rand.Int())
		ss, err := factory.CreateFromGoString(s, &concurrentMemory)
		PanicErr(err)
		err = m.Put(ss, ss.Length(), &concurrentMemory)
		PanicErr(err)
	}

	iter := m.Iterator()
	for iter.Next() {
		iter.KeyRef().Free(&concurrentMemory)
	}
}

func TestStringFactory_CreateFromGoString_bigString(t *testing.T) {
	memory := New(16 * MB)
	defer memory.Free()

	concurrentMemory := memory.NewConcurrentMemory()
	defer concurrentMemory.Destroy()

	factory := NewStringFactory()
	defer factory.Destroy(&concurrentMemory)

	m, err := MakeMap[String, SizeType](0, &concurrentMemory)
	PanicErr(err)
	defer m.Free(&concurrentMemory)

	for _, s := range []string{"hello", string(make([]byte, 1024)), " ", string(make([]byte, 4096)), "!"} {
		ss, err := factory.CreateFromGoString(s, &concurrentMemory)
		PanicErr(err)

		err = m.Put(ss, ss.Length(), &concurrentMemory)
		PanicErr(err)

		t.Logf("%v", m)
	}

	iter := m.Iterator()
	for iter.Next() {
		iter.KeyRef().Free(&concurrentMemory)
	}
}
