package randomvirel

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"runtime"
	"testing"
)

func hexDec(a string) []byte {
	b, _ := hex.DecodeString(a)
	return b
}

var tests = [][3][]byte{
	{
		[]byte("test key 000"),
		[]byte("This is a test"),
		hexDec("a90e9e2a6bb74a0e3a23da303805c36a150a67d293f8fe8bc6f982d7b445568f"),
	},
	{
		[]byte("test key 000"),
		[]byte("Lorem ipsum dolor sit amet"),
		hexDec("eb0aad09fcefb60bb991d2459a97c5098b0d9e8def7033400a46884864890064"),
	},
	{
		[]byte("test key 000"),
		[]byte("sed do eiusmod tempor incididunt ut labore et dolore magna aliqua"),
		hexDec("e4e48f5fe3bb58d9f8e07f8d130bb4e441b98e7102751400437551426e3f337c"),
	},
	{
		[]byte("test key 001"),
		[]byte("sed do eiusmod tempor incididunt ut labore et dolore magna aliqua"),
		hexDec("56052c254fc0e0dfbc80cb38fb6cb47872e52cd68048442f9e539e49116202c0"),
	},
	{
		[]byte("test key 001"),
		hexDec("0b0b98bea7e805e0010a2126d287a2a0cc833d312cb786385a7c2f9de69d25537f584a9bc9977b00000000666fd8753bf61a8631f12984e3fd44f4014eca629276817b56f32e9b68bd82f416"),
		hexDec("6c6e908fbb14c3d60126221cc4e1ced7a9d812c565efdd761f52e98f8dbe15e4"),
	},
}

func TestMain(t *testing.T) {
	flags := GetFlags()
	runtime.GOMAXPROCS(runtime.NumCPU())

	for _, tp := range tests {
		cache, _ := AllocCache(flags)
		t.Log("AllocCache finished")
		InitCache(cache, tp[0])
		t.Log("InitCache finished")

		vm, _ := CreateLightVM(cache, flags)
		hash := CalculateHash(vm, tp[1])

		if !bytes.Equal(hash[:], tp[2]) {
			t.Logf("light mode: incorrect hash: expected %x, got %x", tp[2], hash)
			t.Fail()
		}

		DestroyVM(vm)
		ReleaseCache(cache)
	}

	t.Log("using full mode")

	InitHash(runtime.NumCPU(), true)

	for _, tp := range tests {
		hash := PowHashArbitrarySeed(tp[0], tp[1])

		var hashCorrect = tp[2]

		if !bytes.Equal(hash[:], hashCorrect) {
			t.Fatalf("full mode: incorrect hash: expected %x, got %x", hashCorrect, hash)
		}
	}
}

func TestReuse(t *testing.T) {
	InitHash(len(tests), false)

	doneChan := make(chan bool)

	for _, z := range tests {
		tp := z
		go func() {
			h := PowHashArbitrarySeed(tp[0], tp[1])
			if !bytes.Equal(h[:], tp[2]) {
				panic(fmt.Errorf("seed %s: got hash %x, expected %x", tp[0], h, tp[2]))
			}

			doneChan <- true
		}()
	}

	for i := 0; i < len(tests); i++ {
		<-doneChan
	}
}

func BenchmarkCalculateHash(b *testing.B) {
	flags := GetFlags()
	cache, _ := AllocCache(flags | FlagFullMEM)
	ds, _ := AllocDataset(flags | FlagFullMEM)
	InitCache(cache, []byte("cache"))
	var workerNum = uint64(runtime.NumCPU())
	InitDatasetMultithread(ds, cache, workerNum)
	vm, _ := CreateVM(cache, ds, flags|FlagFullMEM)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		CalculateHash(vm, []byte("hash input"))
	}

	DestroyVM(vm)
}
