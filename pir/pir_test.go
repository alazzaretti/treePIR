package pir

import (
	"math/rand"
	"testing"
	"gotest.tools/assert"
	"fmt"
	"time"
	//"io"
)







func TestPIRPuncTwo(t *testing.T) {
	// if testing.Short() {
 //    t.Skip("skipping test in short mode.")
 //  	}

	dbSizePower := 20
    elmSize := 32
    dbSize := int(1<<dbSizePower)
    fmt.Printf("treepir, dbSize = 2^%d, elmSize = %d bytes\n",dbSizePower,elmSize)
    db := MakeDB(dbSize, elmSize)

	client := NewPIRReader(RandSource(), Server(db), Server(db))
	start := time.Now()
	err := client.Init(TreePIR)
	elapsed := time.Since(start)
	fmt.Printf("treepir init took %s \n", elapsed)
	assert.NilError(t, err)
	start = time.Now()
	val, err := client.Read(3)
	elapsed = time.Since(start)
	fmt.Printf("treepir took %s \n", elapsed)
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

	fmt.Printf("treepir on %d indices took %s \n", (2*dbSize)/262155,elapsed)
	assert.NilError(t, err)
}


func TestPIRPunc(t *testing.T) {
	dbSizePower := 20
    elmSize := 32
    dbSize := int(1<<dbSizePower)
    fmt.Printf("treepir, dbSize = 2^%d, elmSize = %d bytes\n",dbSizePower,elmSize)
    db := MakeDB(dbSize, elmSize)

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
	}
	}
	
	elapsed = time.Since(start)

	fmt.Printf("punc on %d indices took %s \n", (2*dbSize)/262155,elapsed)
	assert.NilError(t, err)

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
