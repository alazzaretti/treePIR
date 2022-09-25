package pir

import (
	"math/rand"
	"testing"
	"gotest.tools/assert"
	"fmt"
	"time"
)

//benchmarks for 2^24 elements query time
//punctwo: 11.058542ms
//punc: 326.916µs but them 3ms
//dpf: 281.46725ms

//projections for 2^32 elements
//punctwo: 500ms
//punc 12ms
//dpf: 56s
func TestPIRPuncTwo(t *testing.T) {
	db := MakeDB(16777216, 16)

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

}


func TestPIRPunc(t *testing.T) {
	db := MakeDB(16777216, 16)

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
	db := MakeDB(4096, 16)

	client := NewPIRReader(RandSource(), Server(db), Server(db))

	err := client.Init(Matrix)
	assert.NilError(t, err)

	val, err := client.Read(0x7)
	assert.NilError(t, err)
	assert.DeepEqual(t, val, db.Row(7))
}

func TestDPF(t *testing.T) {
	db := MakeDB(4096, 16)

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
