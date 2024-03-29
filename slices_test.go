package direct

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func Test_sizeof(t *testing.T) {
	t.Log(Sizeof[int8]())
	t.Log(Sizeof[int16]())
	t.Log(Sizeof[int32]())
	t.Log(Sizeof[int64]())
	t.Log(Sizeof[string]())
	t.Log(Sizeof[[]byte]())
}

func Test_sizeof1(t *testing.T) {
	t.Log(Sizeof[Slice[string]]())
	t.Log(Sizeof[Slice[int]]())
}

func Test_newSlice(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	slice, err := MakeSlice[int32](&localMemory, 10)
	PanicErr(err)
	defer slice.Free(&localMemory)

	t.Log(memory)
	t.Log(Sizeof[sliceHeader]())
	t.Log(slice.header())
}

func TestSlice_Get(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer func() { localMemory.Destroy() }()

	var s Slice[int32]
	defer func() { s.Free(&localMemory) }()

	PanicErr(s.Append(5, &localMemory))
	PanicErr(s.Append(2, &localMemory))
	PanicErr(s.Append(0, &localMemory))

	Assert(s.Get(0) == 5)
	Assert(s.Get(1) == 2)
	Assert(s.Get(2) == 0)
}

func TestSlice_Iterate(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	var s Slice[int32]
	defer func() {
		s.Free(&localMemory)
		t.Log(localMemory.String())
	}()

	PanicErr(s.Append(5, &localMemory))
	PanicErr(s.Append(2, &localMemory))
	PanicErr(s.Append(0, &localMemory))

	s.Iterate(func(i int32) {
		t.Log(i)
	})

	t.Log(s.Length())
	Assert(s.Length() == 3)
	t.Log(s.Capacity())
}

func TestSlice_new2(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	var s Slice[int32]
	PanicErr(s.Append(5, &localMemory))
	s.Free(&localMemory)

	s = 0
	PanicErr(s.Append(5, &localMemory))
	s.Free(&localMemory)
}

func Test_BenchmarkMySlice_Append(t *testing.T) {
	memory := New(32 * 1024 * 1024)
	localMemory := memory.NewLocalMemory()
	for i := 0; i < 10000; i++ {
		var s Slice[int]
		for j := 0; j < 100; j++ {
			PanicErr(s.Append(j, &localMemory))
		}
		s.Free(&localMemory)
	}
	localMemory.Destroy()
	memory.Free()
}

func Test_BenchmarkMySlice_make(t *testing.T) {
	memory := New(32 * 1024 * 1024)
	localMemory := memory.NewLocalMemory()
	for i := 0; i < 10000; i++ {
		ss, _ := MakeSlice[string](&localMemory, 1024)
		_ = ss.Append("", &localMemory)
		ss.Free(&localMemory)
	}
	localMemory.Destroy()
	memory.Free()
}

func BenchmarkSlice_Append(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var s []int
		for j := 0; j < 100; j++ {
			s = append(s, j)
		}
	}
}

func BenchmarkMySlice_Append(b *testing.B) {
	memory := New(32 * 1024 * 1024)
	localMemory := memory.NewLocalMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var s Slice[int]
		for j := 0; j < 100; j++ {
			_ = s.Append(j, &localMemory)
		}
		s.Free(&localMemory)
	}
	b.StopTimer()
	localMemory.Destroy()
	memory.Free()
}

func BenchmarkSlice_make(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ss := make([]string, 0, 1024)
		ss = append(ss, "")
	}
}

func BenchmarkMySlice_make(b *testing.B) {
	memory := New(32 * 1024 * 1024)
	localMemory := memory.NewLocalMemory()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ss, _ := MakeSlice[string](&localMemory, 1024)
		_ = ss.Append("", &localMemory)
		ss.Free(&localMemory)
	}
	b.StopTimer()
	localMemory.Destroy()
	memory.Free()
}

func BenchmarkMySlice_make2(b *testing.B) {
	memory := New(32 * 1024 * 1024)
	localMemory := memory.NewLocalMemory()
	b.ResetTimer()
	ss, _ := MakeSlice[string](&localMemory, 1024)
	_ = ss.Append("", &localMemory)
	for i := 0; i < b.N; i++ {
		ss.Set(0, "")
	}
	ss.Free(&localMemory)
	b.StopTimer()
	localMemory.Destroy()
	memory.Free()
}

func BenchmarkSlice_make_parallel(b *testing.B) {
	wg := sync.WaitGroup{}
	for k := 0; k < 32; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < b.N; i++ {
				ss := make([]string, 0, 1024)
				ss = append(ss, "")
			}
		}()
	}
	wg.Wait()
}

func BenchmarkMySlice_make_parallel(b *testing.B) {
	memory := New(256 * 1024 * 1024)
	b.ResetTimer()
	wg := sync.WaitGroup{}
	for k := 0; k < 32; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localMemory := memory.NewLocalMemory()
			for i := 0; i < b.N; i++ {
				ss, _ := MakeSlice[string](&localMemory, 1024)
				_ = ss.Append("", &localMemory)
				ss.Free(&localMemory)
			}
			localMemory.Destroy()
		}()
	}
	wg.Wait()
	b.StopTimer()
	memory.Free()
}

func Test_BenchmarkMySlice_make_parallel(t *testing.T) {
	memory := New(256 * 1024 * 1024)
	wg := sync.WaitGroup{}
	for k := 0; k < 32; k++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localMemory := memory.NewLocalMemory()
			for i := 0; i < 1000; i++ {
				ss, _ := MakeSlice[string](&localMemory, 1024)
				_ = ss.Append("", &localMemory)
				ss.Free(&localMemory)
			}
			localMemory.Destroy()
		}()
	}
	wg.Wait()
	memory.Free()
}

func TestMakeSliceWithLength(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	s, err := MakeSliceWithLength[int32](&localMemory, 10)
	PanicErr(err)
	defer func() { s.Free(&localMemory) }()

	Assert(s.Length() == 10, s.Length())
	Assert(s.Capacity() >= 10, s.Capacity())

	for i := SizeType(0); i < s.Length(); i++ {
		Assert(s.Get(i) == 0, s.Get(i))
	}
}

func TestSlice_String(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	s, err := MakeSliceWithLength[int32](&localMemory, 10)
	PanicErr(err)
	defer func() { s.Free(&localMemory) }()

	s.Set(0, 10)
	s.Set(5, 520)
	s.Set(9, 111)

	t.Log(s)
}

func Test2DSlice_String(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	var ss Slice[Slice[int32]]
	defer func() { ss.Free(&localMemory) }()
	{
		s, err := MakeSliceWithLength[int32](&localMemory, 10)
		PanicErr(err)
		defer func() { s.Free(&localMemory) }()

		s.Set(0, 10)
		s.Set(5, 520)
		s.Set(9, 111)

		t.Log(s)

		PanicErr(ss.Append(s, &localMemory))
	}

	{
		s, err := MakeSliceWithLength[int32](&localMemory, 5)
		PanicErr(err)
		defer func() { s.Free(&localMemory) }()

		s.Set(1, 4)
		s.Set(3, 8)

		t.Log(s)

		PanicErr(ss.Append(s, &localMemory))
	}

	t.Log(ss)
}

func TestIntersection(t *testing.T) {
	memory := New(1 * MB)
	localMemory := memory.NewLocalMemory()

	var intersection Slice[Slice[Slice[int32]]]

	var err error
	for i := 0; i < 3; i++ {
		t.Log("before", intersection)
		//intersection, err = calIntersection(intersection, &localMemory)
		intersection, err = calIntersectionParallel(intersection, &localMemory)
		PanicErr(err)
		t.Log("after", intersection)
	}

	freeIntersection(intersection, &localMemory)
	localMemory.Destroy()
	memory.Free()
}

func freeIntersection(intersection Slice[Slice[Slice[int32]]], m *LocalMemory) {
	intersection.Iterate(func(idPairs Slice[Slice[int32]]) {
		idPairs.Iterate(func(ids Slice[int32]) {
			ids.Free(m)
		})
		idPairs.Free(m)
	})
	intersection.Free(m)
}

// calIntersection should free intersection
func calIntersection(intersection Slice[Slice[Slice[int32]]], m *LocalMemory) (Slice[Slice[Slice[int32]]], error) {
	if intersection.Length() < 1 {
		return generatePredicateIntersection(m)
	}
	var resultIntersection Slice[Slice[Slice[int32]]]
	var err error
	intersection.Iterate(func(idPairs Slice[Slice[int32]]) {
		var tempIdPairs Slice[Slice[Slice[int32]]]
		value2Index := make(map[int32]SizeType)
		idPairs.IterateIndex(func(tableIndex SizeType, ids Slice[int32]) {
			ids.Iterate(func(rowId int32) {
				value := rowId % 4
				index, ok := value2Index[value]
				if !ok {
					value2Index[value] = tempIdPairs.Length()
					index = tempIdPairs.Length()
					var temp Slice[Slice[int32]]
					temp, err = MakeSliceWithLength[Slice[int32]](m, 2)
					if err != nil {
						return
					}
					err = tempIdPairs.Append(temp, m)
					if err != nil {
						return
					}
				}
				temp := tempIdPairs.Get(index).Get(tableIndex)
				err = temp.Append(rowId, m)
				if err != nil {
					return
				}
				tempIdPairs.Get(index).Set(tableIndex, temp)
			})
		})
		tempIdPairs.Iterate(func(temp Slice[Slice[int32]]) {
			err = resultIntersection.Append(temp, m)
			if err != nil {
				return
			}
		})
		tempIdPairs.Free(m)
	})
	freeIntersection(intersection, m)
	return resultIntersection, err
}

func calIntersectionParallel(intersection Slice[Slice[Slice[int32]]], m *LocalMemory) (Slice[Slice[Slice[int32]]], error) {
	if intersection.Length() < 1 {
		return generatePredicateIntersection(m)
	}
	var resultIntersection Slice[Slice[Slice[int32]]]
	var theErr atomic.Value
	var wg sync.WaitGroup
	var lock sync.Mutex
	intersection.Iterate(func(idPairs Slice[Slice[int32]]) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			memory := m.NewLocalMemory()
			defer memory.Destroy()

			var tempIdPairs Slice[Slice[Slice[int32]]]
			value2Index := make(map[int32]SizeType)
			idPairs.IterateIndexBreakable(func(tableIndex SizeType, ids Slice[int32]) bool {
				if theErr.Load() != nil {
					return false
				}
				ids.IterateBreakable(func(rowId int32) bool {
					value := rowId % 4
					index, ok := value2Index[value]
					if !ok {
						value2Index[value] = tempIdPairs.Length()
						index = tempIdPairs.Length()
						temp, err := MakeSliceWithLength[Slice[int32]](&memory, 2)
						if err != nil {
							theErr.Store(err)
							return false
						}
						err = tempIdPairs.Append(temp, &memory)
						if err != nil {
							theErr.Store(err)
							return false
						}
					}
					temp := tempIdPairs.Get(index).Get(tableIndex)
					err := temp.Append(rowId, &memory)
					if err != nil {
						theErr.Store(err)
						return false
					}
					tempIdPairs.Get(index).Set(tableIndex, temp)

					time.Sleep(10 * time.Millisecond)
					return true
				})
				return true
			})
			if theErr.Load() != nil {
				return
			}
			tempIdPairs.IterateBreakable(func(temp Slice[Slice[int32]]) bool {
				lock.Lock()
				err := resultIntersection.Append(temp, &memory)
				lock.Unlock()
				if err != nil {
					theErr.Store(err)
					return false
				}
				return true
			})
			if theErr.Load() != nil {
				return
			}
			tempIdPairs.Free(&memory)
		}()
	})
	wg.Wait()
	err := theErr.Load()
	if err != nil {
		return 0, err.(error)
	}
	freeIntersection(intersection, m)
	return resultIntersection, nil
}

func generatePredicateIntersection(m *LocalMemory) (Slice[Slice[Slice[int32]]], error) {
	var intersection Slice[Slice[Slice[int32]]]
	var leftPli = map[int32][]int32{
		0: {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		1: {11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		3: {100, 200},
		4: {55, 66, 77, 88, 99},
	}
	for _, int32s := range leftPli {
		idPairs, err := MakeSliceWithLength[Slice[int32]](m, 2)
		if err != nil {
			return 0, err
		}
		s1, err := MakeSliceFromGoSlice(m, int32s)
		if err != nil {
			return 0, err
		}
		s2, err := MakeSliceFromGoSlice(m, int32s)
		if err != nil {
			return 0, err
		}
		idPairs.Set(0, s1)
		idPairs.Set(1, s2)

		err = intersection.Append(idPairs, m)
		if err != nil {
			return 0, err
		}
	}

	return intersection, nil
}

func TestMakeSliceFromGoSlice(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	var ss = PanicErr1(MakeSliceFromGoSlice(&localMemory, []string{"hello", ", ", "world", "!"}))
	defer func() { ss.Free(&localMemory) }()

	t.Log(ss)
}

func TestMakeSliceFromGoSlice1000(t *testing.T) {
	memory := New(1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	var temp []int
	for i := 0; i < 1000; i++ {
		temp = append(temp, rand.Int())
	}

	var ss = PanicErr1(MakeSliceFromGoSlice(&localMemory, temp))
	defer func() { ss.Free(&localMemory) }()

	if len(temp) != int(ss.Length()) {
		panic(fmt.Sprintf("%d != %d", len(temp), int(ss.Length())))
	}

	for i, e := range temp {
		if e != ss.Get(SizeType(i)) {
			panic(fmt.Sprintf("%d != %d", e, ss.Get(SizeType(i))))
		}
	}
}

func TestMakeSliceFromGoSliceEmpty(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	var temp []string
	var ss = PanicErr1(MakeSliceFromGoSlice(&localMemory, temp))
	defer func() { ss.Free(&localMemory) }()

	t.Log(ss)
}

func TestSlice_Copy(t *testing.T) {
	memory := New(4 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	var ss = PanicErr1(MakeSliceFromGoSlice(&localMemory, []string{"hello", ", ", "world", "!"}))
	defer func() { ss.Free(&localMemory) }()

	t.Log(ss)

	ss2, err := ss.Copy(&localMemory)
	PanicErr(err)
	defer func() { ss2.Free(&localMemory) }()
	t.Log(ss2)
}

func TestMakeSliceCopy1000(t *testing.T) {
	memory := New(1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	var temp []int
	for i := 0; i < 1000; i++ {
		temp = append(temp, rand.Int())
	}

	var ss = PanicErr1(MakeSliceFromGoSlice(&localMemory, temp))
	defer func() { ss.Free(&localMemory) }()

	if len(temp) != int(ss.Length()) {
		panic(fmt.Sprintf("%d != %d", len(temp), int(ss.Length())))
	}

	ss2, err := ss.Copy(&localMemory)
	PanicErr(err)
	defer func() { ss2.Free(&localMemory) }()

	if len(temp) != int(ss2.Length()) {
		panic(fmt.Sprintf("%d != %d", len(temp), int(ss2.Length())))
	}

	t.Log(len(temp), ss.Length(), ss2.Length())

	for i, e := range temp {
		if e != ss.Get(SizeType(i)) {
			panic(fmt.Sprintf("%d != %d", e, ss.Get(SizeType(i))))
		}
		if e != ss2.Get(SizeType(i)) {
			panic(fmt.Sprintf("%d != %d", e, ss2.Get(SizeType(i))))
		}
	}
}

func TestSlice_AppendBatch(t *testing.T) {
	memory := New(1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	var ss = PanicErr1(MakeSliceFromGoSlice(&localMemory, []int{1, 2, 3}))
	defer func() { ss.Free(&localMemory) }()

	var s2 Slice[int]
	defer func() { s2.Free(&localMemory) }()
	err := s2.AppendBatch(ss, &localMemory)
	PanicErr(err)

	t.Log(ss)
	t.Log(s2)
}

func TestSlice_AppendBatch2(t *testing.T) {
	memory := New(1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	var ss = PanicErr1(MakeSliceFromGoSlice(&localMemory, []int{1, 2, 3}))
	defer func() { ss.Free(&localMemory) }()

	var s2 Slice[int]
	defer func() { s2.Free(&localMemory) }()
	err := s2.AppendBatch(ss, &localMemory)
	PanicErr(err)
	err = s2.AppendBatch(ss, &localMemory)
	PanicErr(err)

	t.Log(ss)
	t.Log(s2)
}

func TestSlice_AppendBatchSelf(t *testing.T) {
	memory := New(1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	var ss = PanicErr1(MakeSliceFromGoSlice(&localMemory, []int{1, 2, 3}))
	defer func() { ss.Free(&localMemory) }()

	err := ss.AppendBatch(ss, &localMemory)
	PanicErr(err)

	t.Log(ss)
}

func TestSlice_AppendBatch1000(t *testing.T) {
	memory := New(1024 * 1024)
	defer memory.Free()
	localMemory := memory.NewLocalMemory()
	defer localMemory.Destroy()

	gs := make([]int, 500)
	for i := range gs {
		gs[i] = 22
	}

	var ss = PanicErr1(MakeSliceFromGoSlice(&localMemory, gs))
	defer func() { ss.Free(&localMemory) }()

	err := ss.AppendBatch(ss, &localMemory)
	PanicErr(err)
	t.Log(ss.Length())
	Assert(ss.Length() == 1000)
	ss.Iterate(func(i int) {
		if i != 22 {
			panic(i)
		}
	})
}

func TestSlice_Iterator(t *testing.T) {
	memory := New(1 * KB)
	defer memory.Free()
	local := memory.NewLocalMemory()
	defer local.Destroy()
	slice := PanicErr1(MakeSliceFromGoSlice(&local, []int{1, 2, 5, 10}))
	iter := slice.Iterator()
	for iter.Next() {
		t.Log(iter.Index(), iter.Value())
	}
	slice.Free(&local)
}

func TestSlice_Move(t *testing.T) {
	memory := New(1 * MB)
	defer memory.Free()
	local := memory.NewLocalMemory()
	defer local.Destroy()

	var ss Slice[Slice[int]]
	defer func() {
		ss.Iterate(func(s Slice[int]) {
			s.Free(&local)
		})
		ss.Free(&local)
	}()
	for i := 0; i < 10; i++ {
		var s Slice[int]
		defer func() { s.Free(&local) }()
		for j := 0; j < i; j++ {
			_ = s.Append(j, &local)
		}
		if i%2 == 0 {
			_ = ss.Append(s.Move(), &local)
		}
	}

	// [[0 1 2 3 4 5 6 7 8 9] [0 1 2 3 4 5 6 7 8 9] [0 1 2 3 4 5 6 7 8 9] [0 1 2 3 4 5 6 7 8 9] [0 1 2 3 4 5 6 7 8 9]]
	t.Log(ss)
}

func TestSlice_DoubleFreeCheck(t *testing.T) {
	memory := New(1 * KB)
	defer memory.Free()
	local := memory.NewLocalMemory()
	defer local.Destroy()
	slice := PanicErr1(MakeSliceFromGoSlice(&local, []int{1, 2, 5, 10}))
	t.Log(slice)
	slice.Free(&local)

	if asserted {
		t.Log("========== THE FOLLOWING ERROR IS OK ============")
		var err any
		func() {
			defer func() {
				err = recover()
			}()
			slice.Free(&local)
		}()

		Assert(err != nil, err)
	}
}

func Benchmark_MultiSliceAppendGo(b *testing.B) {
	rand.Seed(1)
	size := 100 * 100 * 100
	var values = make([]int, size)
	for i := 0; i < size; i++ {
		values[i] = rand.Int()
	}

	for i := 0; i < b.N; i++ {
		m := map[int]int{}
		var ss [][]int
		for _, value := range values {
			key := value % 128
			index, ok := m[key]
			if !ok {
				index = len(ss)
				ss = append(ss, nil)
				m[key] = index
			}
			ss[index] = append(ss[index], value)
		}
		m = map[int]int{}
		ss = nil
	}
}

func Benchmark_MultiSliceAppendRock(b *testing.B) {
	rand.Seed(1)
	size := 100 * 100 * 100
	var values = make([]int, size)
	for i := 0; i < size; i++ {
		values[i] = rand.Int()
	}
	memory := New(1 * GB)
	local := memory.NewLocalMemory()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := PanicErr1(MakeMap[int, SizeType](0, &local))
		var ss Slice[Slice[int]] = nullSlice

		for _, value := range values {
			key := value % 128
			index, ok := m.Get2(key)
			if !ok {
				index = ss.Length()
				PanicErr(ss.Append(nullSlice, &local))
				PanicErr(m.DirectPut(key, index, &local))
			}
			PanicErr(ss.RefAt(index).Append(value, &local))
		}
		ssIter := ss.Iterator()
		for ssIter.Next() {
			ssIter.Ref().Free(&local)
		}
		ss.Free(&local)
		m.Free(&local)
	}

	b.StopTimer()
	local.Destroy()
	memory.Free()
}
