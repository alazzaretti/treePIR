package pir

import (
	"math/rand"
	"testing"
	"gotest.tools/assert"
	"fmt"
	"time"
)

//benchmarks for 2^24 elements query time
//punctwo: 2.5ms
//punc: 326.916µs but them 3ms
//dpf: 281.46725ms

//projections for 2^32 elements
//punctwo: 200ms
//punc 12ms
//dpf: 56s
func TestPIRPuncTwo(t *testing.T) {
	dbSize := 16777216
	db := MakeDB(dbSize, 16)

	client := NewPIRReader(RandSource(), Server(db), Server(db))

	err := client.Init(PuncTwo)
	assert.NilError(t, err)
	start := time.Now()
	val, err := client.Read(521)
	elapsed := time.Since(start)
	fmt.Printf("punctwo took %s \n", elapsed)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(521))

	// // Test refreshing by reading the same item again
	val, err = client.Read(521)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(521))
	//read every other index
	start = time.Now()
	for i:= 0; i<dbSize;i+=4097 {
		val,err=client.Read(i)
		assert.NilError(t, err)

		
		
		assert.DeepEqual(t, val, db.Row(i))
		val,err=client.Read(i)
		assert.NilError(t, err)

		
		
		assert.DeepEqual(t, val, db.Row(i))

		// fmt.Println(i)
		// fmt.Println(val)
		// fmt.Println(db.Row(i))
	}
	elapsed = time.Since(start)
	fmt.Printf("punctwo on %d indices took %s \n", (2*dbSize)/4097,elapsed)

	//test pirpunc in same test
	fmt.Println("punc now-------")
	err2 := client.Init(Punc)
	assert.NilError(t, err2)

	start2 := time.Now()
	val2, err2 := client.Read(793)

	elapsed2 := time.Since(start2)
	fmt.Printf("punc took %s \n", elapsed2)
	assert.NilError(t, err2)
	assert.DeepEqual(t, val2, db.Row(793))

	// Test refreshing by reading the same item again
	val2, err2 = client.Read(793)
	assert.NilError(t, err2)
	assert.DeepEqual(t, val2, db.Row(793))
}


func TestPIRPunc(t *testing.T) {
	db := MakeDB(4096, 16)

	client := NewPIRReader(RandSource(), Server(db), Server(db))

	err := client.Init(Punc)
	assert.NilError(t, err)

	start := time.Now()
	val, err := client.Read(793)

	elapsed := time.Since(start)
	fmt.Printf("punc took %s \n", elapsed)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(793))

	// Test refreshing by reading the same item again
	val, err = client.Read(793)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(793))

}

func TestMatrix(t *testing.T) {
	db := MakeDB(65536, 16)

	client := NewPIRReader(RandSource(), Server(db), Server(db))

	err := client.Init(Matrix)
	assert.NilError(t, err)

	val, err := client.Read(0x23)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(0x23))
}

func TestDPF(t *testing.T) {
	db := MakeDB(65536, 16)

	client := NewPIRReader(RandSource(), Server(db), Server(db))

	err := client.Init(DPF)
	assert.NilError(t, err)
	start := time.Now()
	val, err := client.Read(900)
	elapsed := time.Since(start)
	fmt.Printf("dpf took %s \n", elapsed)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(900))
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
