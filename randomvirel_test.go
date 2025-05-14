package randomvirel

import (
	"bytes"
	"encoding/hex"
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
		hexDec("80cf38cee556e27b1a304b2bdc1aa3b93a6f2e3ed1a6a0f2b4e7fbd3f2270a75"),
	},
	{
		[]byte("test key 000"),
		[]byte("Lorem ipsum dolor sit amet"),
		hexDec("fb622344fd64e7a2f93b6b660b97d1e57668075a42a329fc5e302a3122098dbb"),
	},
	{
		[]byte("test key 000"),
		[]byte("sed do eiusmod tempor incididunt ut labore et dolore magna aliqua"),
		hexDec("2534becac5c7594708d435b231e8c3482bf46679a92d3124a59ce82335fc0205"),
	},
	{
		[]byte("test key 001"),
		[]byte("sed do eiusmod tempor incididunt ut labore et dolore magna aliqua"),
		hexDec("66759fa6969545fb3dcc6a6f124cb837c81bff6364d7f7e0af822a4f764b15e9"),
	},
	{
		[]byte("test key 001"),
		hexDec("0b0b98bea7e805e0010a2126d287a2a0cc833d312cb786385a7c2f9de69d25537f584a9bc9977b00000000666fd8753bf61a8631f12984e3fd44f4014eca629276817b56f32e9b68bd82f416"),
		hexDec("3bbb57a343c6e504adc5afd1673c2d595cbbe33512657f06079eecc292482019"),
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

		var hashCorrect = make([]byte, hex.DecodedLen(len(tp[2])))
		_, err := hex.Decode(hashCorrect, tp[2])
		if err != nil {
			t.Log(err)
		}

		if !bytes.Equal(hash[:], hashCorrect) {
			t.Fatalf("full mode: incorrect hash: expected %x, got %x", hashCorrect, hash)
		}
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
