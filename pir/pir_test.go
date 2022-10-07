package pir

import (
	"math/rand"
	"testing"
	"gotest.tools/assert"
	"fmt"
	"time"
)

//DPF STUFF SAVING:
//dpf with 2^22 indicies took 84.394316ms 
// dpf with 2^14 indicies took 344.57µs 
// dpf with 2^11 indicies took 52.259µs 
//2^28 benchmarking:
//
//punctwo took 16.921788ms 
//punctwo on 2047 indices took 47.17515586s = 23ms per query
//
//punc took 2.246068ms 
//punc on 2047 indices took 1m59.024780883s = 58.1ms per query
//
//dpf took 5.48s per query avg (4m45s for 52 queries)

//

//sticky stuff talk to babis:
//1. implementation: good perf. but cannot load more than 2^28 elems? is it okay to benchmark like this?
//      a. practical sing.serv pir does not actually have polylog bandwidth 
//          (although there is the dottling scheme that does so we can use that?)
//      b. w/ dpf: very good performance, maybe just benchmark that and mention sing. serv

//2. failure case: sticky: makes protocol more complication - 
//       just need to think about it a bit more

//3. angel implementation actually is secure! do we need to benchmark it or can we just use
//     their numbers or something?
//
//(4. grants recs, can send fb research text, do you need anything else?)


//2^26: number of wikipedia pages
//avg wikipedia article size: 357.142857 bytes
//2^34: number of passwords: hash is 256 bits

//benchmarks for 2^24 elements query time
//punctwo: 2.5ms 
//punc: 326.916µs     but them 3ms
//dpf: 281.46725ms

//2^26: avg of 0.00885225885s per query over 32760 queries
//checklist does about 1ms, so ~10x? logn/2?

//2^28 x 32: 16ms so 2^30 = 32ms and 2^32 ~64 ms?

//projections for 2^32 elements
//punctwo: 100ms
//punc 8ms
//dpf: 56s
func TestPIRPuncTwo(t *testing.T) {
	dbSize := 268435456
	fmt.Printf("dbSize = 2^28")
	db := MakeDB(dbSize, 32)

	client := NewPIRReader(RandSource(), Server(db), Server(db))
	start := time.Now()
	err := client.Init(PuncTwo)
	elapsed := time.Since(start)
	fmt.Printf("punctwo init took %s \n", elapsed)
	assert.NilError(t, err)
	start = time.Now()
	val, err := client.Read(3)
	elapsed = time.Since(start)
	fmt.Printf("punctwo took %s \n", elapsed)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(3))
	// // Test refreshing by reading the same item again
	val, err = client.Read(3)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(3))
	//read every other index
	start = time.Now()
	for j := 0; j < 2; j++ {
		for i:= 0; i<dbSize;i+=262155 {
	 	val,err=client.Read(i)
	 	//assert.NilError(t, err)

		
		
	 	//assert.DeepEqual(t, val, db.Row(i))
	 	//val,err=client.Read(i)
	 	//assert.NilError(t, err)

		
		
	 	//assert.DeepEqual(t, val, db.Row(i))

	// 	// fmt.Println(i)
	// 	// fmt.Println(val)
	// 	// fmt.Println(db.Row(i))
	}
	}
	
	elapsed = time.Since(start)

	fmt.Printf("punctwo on %d indices took %s \n", (2*dbSize)/262155,elapsed)
	assert.NilError(t, err)
	//test pirpunc in same test

	
	// fmt.Println("punc now-------")
	// err2 := client.Init(Punc)
	// assert.NilError(t, err2)

	// start2 := time.Now()
	// val2, err2 := client.Read(793)

	// elapsed2 := time.Since(start2)
	// fmt.Printf("punc took %s \n", elapsed2)
	// assert.NilError(t, err2)
	// assert.DeepEqual(t, val2, db.Row(793))

	// // Test refreshing by reading the same item again
	// val2, err2 = client.Read(793)
	// assert.NilError(t, err2)
	// assert.DeepEqual(t, val2, db.Row(793))
}
func TestPIRPuncTwoPartTwo(t *testing.T) {
	dbSize := 4194304
	fmt.Printf("dbSize = 2^22, blocksize 256 bytes")
	db := MakeDB(dbSize, 256)

	client := NewPIRReader(RandSource(), Server(db), Server(db))
	start := time.Now()
	err := client.Init(PuncTwo)
	elapsed := time.Since(start)
	fmt.Printf("punctwo init took %s \n", elapsed)
	assert.NilError(t, err)
	start = time.Now()
	val, err := client.Read(3)
	elapsed = time.Since(start)
	fmt.Printf("punctwo took %s \n", elapsed)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(3))
	// // Test refreshing by reading the same item again
	val, err = client.Read(3)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(3))
	//read every other index
	start = time.Now()
	for j := 0; j < 2; j++ {
		for i:= 0; i<dbSize;i+=4097 {
	 	val,err=client.Read(i)
	 	//assert.NilError(t, err)

		
		
	 	//assert.DeepEqual(t, val, db.Row(i))
	 	//val,err=client.Read(i)
	 	//assert.NilError(t, err)

		
		
	 	//assert.DeepEqual(t, val, db.Row(i))

	// 	// fmt.Println(i)
	// 	// fmt.Println(val)
	// 	// fmt.Println(db.Row(i))
	}
	}
	
	elapsed = time.Since(start)

	fmt.Printf("punctwo on %d indices took %s \n", (2*dbSize)/4097,elapsed)
	assert.NilError(t, err)
	//test pirpunc in same test

	
	// fmt.Println("punc now-------")
	// err2 := client.Init(Punc)
	// assert.NilError(t, err2)

	// start2 := time.Now()
	// val2, err2 := client.Read(793)

	// elapsed2 := time.Since(start2)
	// fmt.Printf("punc took %s \n", elapsed2)
	// assert.NilError(t, err2)
	// assert.DeepEqual(t, val2, db.Row(793))

	// // Test refreshing by reading the same item again
	// val2, err2 = client.Read(793)
	// assert.NilError(t, err2)
	// assert.DeepEqual(t, val2, db.Row(793))
}

func TestPIRPunc(t *testing.T) {
	dbSize := 268435456
	fmt.Printf("dbSize = 2^28, 32 bits")
	db := MakeDB(dbSize, 32)

	client := NewPIRReader(RandSource(), Server(db), Server(db))
	start := time.Now()
	err := client.Init(Punc)
	elapsed := time.Since(start)
	fmt.Printf("punc init took %s \n", elapsed)
	assert.NilError(t, err)
	start = time.Now()
	val, err := client.Read(3)
	elapsed = time.Since(start)
	fmt.Printf("punc took %s \n", elapsed)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(3))
	// // Test refreshing by reading the same item again
	val, err = client.Read(3)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(3))
	//read every other index
	start = time.Now()
	for j := 0; j < 2; j++ {
		for i:= 0; i<dbSize;i+=262155 {
	 	val,err=client.Read(i)
	 	//assert.NilError(t, err)

		
		
	 	//assert.DeepEqual(t, val, db.Row(i))
	 	//val,err=client.Read(i)
	 	//assert.NilError(t, err)

		
		
	 	//assert.DeepEqual(t, val, db.Row(i))

	// 	// fmt.Println(i)
	// 	// fmt.Println(val)
	// 	// fmt.Println(db.Row(i))
	}
	}
	
	elapsed = time.Since(start)

	fmt.Printf("punc on %d indices took %s \n", (2*dbSize)/262155,elapsed)
	assert.NilError(t, err)

}
func TestPIRPuncPartTwo(t *testing.T) {
	dbSize := 4194304
	fmt.Printf("dbSize = 2^22, blocksize 256 bytes")
	db := MakeDB(dbSize, 256)

	client := NewPIRReader(RandSource(), Server(db), Server(db))
	start := time.Now()
	err := client.Init(Punc)
	elapsed := time.Since(start)
	fmt.Printf("punc init took %s \n", elapsed)
	assert.NilError(t, err)
	start = time.Now()
	val, err := client.Read(3)
	elapsed = time.Since(start)
	fmt.Printf("punc took %s \n", elapsed)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(3))
	// // Test refreshing by reading the same item again
	val, err = client.Read(3)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(3))
	//read every other index
	start = time.Now()
	for j := 0; j < 2; j++ {
		for i:= 0; i<dbSize;i+=4097 {
	 	val,err=client.Read(i)
	 	//assert.NilError(t, err)

		
		
	 	//assert.DeepEqual(t, val, db.Row(i))
	 	//val,err=client.Read(i)
	 	//assert.NilError(t, err)

		
		
	 	//assert.DeepEqual(t, val, db.Row(i))

	// 	// fmt.Println(i)
	// 	// fmt.Println(val)
	// 	// fmt.Println(db.Row(i))
	}
	}
	
	elapsed = time.Since(start)

	fmt.Printf("punc on %d indices took %s \n", (2*dbSize)/4097,elapsed)
	assert.NilError(t, err)
	//test pirpunc in same test

	
	// fmt.Println("punc now-------")
	// err2 := client.Init(Punc)
	// assert.NilError(t, err2)

	// start2 := time.Now()
	// val2, err2 := client.Read(793)

	// elapsed2 := time.Since(start2)
	// fmt.Printf("punc took %s \n", elapsed2)
	// assert.NilError(t, err2)
	// assert.DeepEqual(t, val2, db.Row(793))

	// // Test refreshing by reading the same item again
	// val2, err2 = client.Read(793)
	// assert.NilError(t, err2)
	// assert.DeepEqual(t, val2, db.Row(793))
}


func TestMatrix(t *testing.T) {
	db := MakeDB(256, 16)

	client := NewPIRReader(RandSource(), Server(db), Server(db))

	err := client.Init(Matrix)
	assert.NilError(t, err)

	val, err := client.Read(0x23)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(0x23))
}

func TestDPF(t *testing.T) {
	db := MakeDB(4194304, 16)

	client := NewPIRReader(RandSource(), Server(db), Server(db))

	err := client.Init(DPF)
	assert.NilError(t, err)
	start := time.Now()
	val, err := client.Read(502)
	elapsed := time.Since(start)
	fmt.Printf("dpf with 2^22 indicies took %s \n", elapsed)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(502))


	db = MakeDB(16384, 16)

	client = NewPIRReader(RandSource(), Server(db), Server(db))

	err = client.Init(DPF)
	assert.NilError(t, err)
	start = time.Now()
	val, err = client.Read(502)
	elapsed = time.Since(start)
	fmt.Printf("dpf with 2^14 indicies took %s \n", elapsed)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(502))

	db = MakeDB(2048, 16)

	client = NewPIRReader(RandSource(), Server(db), Server(db))

	err = client.Init(DPF)
	assert.NilError(t, err)
	start = time.Now()
	val, err = client.Read(502)
	elapsed = time.Since(start)
	fmt.Printf("dpf with 2^11 indicies took %s \n", elapsed)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(502))

}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randStringBytes(r *rand.Rand, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[i%len(letterBytes)]
	}
	return string(b)
}

func TestSample(t *testing.T) {
	client := puncClient{randSource: RandSource()}
	assert.Equal(t, 1, client.sample(10, 0, 10))
	assert.Equal(t, 2, client.sample(0, 10, 10))
	assert.Equal(t, 0, client.sample(0, 0, 10))
	count := make([]int, 3)
	for i := 0; i < 1000; i++ {
		count[client.sample(10, 10, 30)]++
	}
	for _, c := range count {
		assert.Check(t, c < 380)
		assert.Check(t, c > 280)
	}
}
