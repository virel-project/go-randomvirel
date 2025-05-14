package randomvirel

import (
	"bytes"
	"sync"
)

var hash_fullmode = false

type Seed [32]byte

var globMut sync.Mutex
var globDataset Dataset
var globCache Cache
var globSeed []byte
var globNumThreads int

var threads chan VM

var flags = GetFlags()

func InitHash(numthreads int, fullmode bool) {
	globMut.Lock()
	defer globMut.Unlock()

	// check that the hash hasn't already been initialized
	if globNumThreads != 0 {
		return
	}

	if numthreads < 1 {
		numthreads = 1
	}

	globNumThreads = numthreads
	hash_fullmode = fullmode
	threads = make(chan VM, globNumThreads+1)
	for i := 0; i < globNumThreads; i++ {
		threads <- nil
	}
	if hash_fullmode {
		flags |= FlagFullMEM
	}
}

// PowHash is a high-level function which takes a seed and some data, and returns the hash.
// You must initialize the VM with the InitHash function before using PowHash.
func PowHash(seed Seed, data []byte) [32]byte {
	return PowHashArbitrarySeed(seed[:], data)
}

func PowHashArbitrarySeed(seed, data []byte) [32]byte {
	var curVM VM
	func() {
		globMut.Lock()
		defer globMut.Unlock()

		if !bytes.Equal(globSeed, seed) || globCache == nil {
			var err error
			var shouldAlloc bool = globCache == nil

			if shouldAlloc {
				globCache, err = AllocCache(flags)
				if err != nil {
					panic(err)
				}
			}
			InitCache(globCache, seed[:])

			if hash_fullmode {
				if shouldAlloc {
					globDataset, err = AllocDataset(flags)
					if err != nil {
						panic(err)
					}
				}
				InitDatasetMultithread(globDataset, globCache, uint64(globNumThreads))
			}
			globSeed = seed

			for i := 0; i < globNumThreads; i++ {
				thread := <-threads
				if thread != nil {
					DestroyVM(thread)
				}
				if hash_fullmode {
					thread, err = CreateVM(globCache, globDataset, flags)
					if err != nil {
						panic(err)
					}
				} else {
					thread, err = CreateLightVM(globCache, flags)
					if err != nil {
						panic(err)
					}
				}
				threads <- thread
			}
			curVM = <-threads
		} else {
			curVM = <-threads
		}
	}()

	h := CalculateHash(curVM, data)
	threads <- curVM
	return h
}
