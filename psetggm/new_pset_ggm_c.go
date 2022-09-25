package psetggm

/*
#cgo amd64 CXXFLAGS: -msse2 -msse -march=native -maes -Ofast -std=c++11
#cgo arm64 CXXFLAGS: -march=armv8-a+fp+simd+crypto+crc -Ofast -std=c++11
#cgo LDFLAGS: -static-libstdc++
#include "pset_ggm.h"
#include "xor.h"
#include "answer.h"
*/
import "C"
import (
	"unsafe"
)

type SecondGGMSetGeneratorC struct {
	workspace []byte
	cgen      *C.new_generator
}

func NewSecondGGMSetGeneratorC(univSize, setSize, sqrtUnivSize int) *SecondGGMSetGeneratorC {
	size := C.new_workspace_size(C.uint(univSize), C.uint(setSize))
	gen := SecondGGMSetGeneratorC{
		workspace: make([]byte, size),
	}
	gen.cgen = C.new_pset_ggm_init(C.uint(univSize), C.uint(sqrtUnivSize), C.uint(setSize),
		(*C.uchar)(&gen.workspace[0]))
	return &gen
}

func (gen *SecondGGMSetGeneratorC) Eval(seed []byte, elems []int) {
	C.new_pset_ggm_eval(gen.cgen, (*C.uchar)(&seed[0]), (*C.ulonglong)(unsafe.Pointer(&elems[0])))
}

func (gen *SecondGGMSetGeneratorC) EvalOn(seed []byte, pos int) int {
	pset := make([]byte, C.new_pset_buffer_size(gen.cgen))
	return int(C.new_pset_ggm_eval_on(gen.cgen, (*C.uchar)(&seed[0]), C.uint(pos), (*C.uchar)(&pset[0])))
}

func (gen *SecondGGMSetGeneratorC) Punc(seed []byte, pos int) []byte {
	pset := make([]byte, C.new_pset_buffer_size(gen.cgen))
	C.new_pset_ggm_punc(gen.cgen, (*C.uchar)(&seed[0]), C.uint(pos), (*C.uchar)(&pset[0]))
	return pset
}

func (gen *SecondGGMSetGeneratorC) EvalPunctured(pset []byte, hole int, elems []int) {
	//C.pset_ggm_eval_punc(gen.cgen, (*C.uchar)(&pset[0]), C.uint(hole), (*C.ulonglong)(unsafe.Pointer(&elems[0])), (*C.uint)(&next_height[0]), (*C.uchar)(&db[0]))
	return;
}

// func XorBlocks(db []byte, offsets []int, out []byte) {
// 	C.xor_rows((*C.uchar)(&db[0]), C.uint(len(db)), (*C.ulonglong)(unsafe.Pointer(&offsets[0])), C.uint(len(offsets)), C.uint(len(out)), (*C.uchar)(&out[0]))
// }

// func XorHashesByBitVector(db []byte, indexing []byte, out []byte) {
// 	C.xor_hashes_by_bit_vector((*C.uchar)(&db[0]), C.uint(len(db)),
// 		(*C.uchar)(&indexing[0]), (*C.uchar)(&out[0]))
// }

func (gen *SecondGGMSetGeneratorC) Distinct(elems []int) bool {
	return (C.new_distinct(gen.cgen, (*C.ulonglong)(unsafe.Pointer(&elems[0])), C.uint(len(elems))) != 0)
}

func FastAnswerTwo(pset []byte, univSize, setSize, shift int, next_height []int,db []byte, rowLen int, out []byte) {
	C.new_answer((*C.uchar)(&pset[0]), C.uint(univSize), C.uint(setSize), C.uint(shift), (*C.uint)(unsafe.Pointer(&next_height[0])),
		(*C.uchar)(&db[0]), C.uint(len(db)), C.uint(rowLen),C.uint(rowLen),(*C.uchar)(&out[0])) //C.uint(len(out)), (*C.uchar)(&out[0]))
}

func GetHeightsArr(setSize int, heightArr []int) {
	C.get_heights_wrapper(C.uint(setSize), (*C.uint)(unsafe.Pointer(&heightArr[0])))
}

func GetHeight(setSize int) int {
	return int(C.get_height(C.uint(setSize)))
}
