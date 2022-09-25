package pir

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"

	"checklist/psetggm"
)

type Present int

const Present_Yes Present = 0

type Set []int

type SetKey struct {
	id, shift uint32
}

type BaseGenerator interface {
	Eval(seed []byte, elems []int, opt_shift uint32)
	EvalOn(seed []byte, pos int, opt_shift uint32) int
	Punc(seed []byte, pos int) []byte
	EvalPunctured(pset []byte, hole int, elems []int)
	Distinct(elems []int) bool
}

type PuncturableSet struct {
	SetKey
	univSize, setSize int
	seed              [16]byte
	elems             Set
}

type PuncturedSet struct {
	UnivSize, SetSize int
	Keys              []byte
	Hole              int
	Shift             uint32
}

type SetGenerator struct {
	baseGen           BaseGenerator
	num               uint32
	idGen             cipher.Block
	univSize, setSize int
}

func NewSetGenerator(masterKey PRGKey, startId uint32, univSize int, setSize int) SetGenerator {
	aes, err := aes.NewCipher(masterKey[:])
	if err != nil {
		panic(err)
	}

	return SetGenerator{
		baseGen:  psetggm.NewGGMSetGeneratorC(univSize, setSize),
		num:      startId,
		idGen:    aes,
		univSize: univSize,
		setSize:  setSize,
	}
}


func NewSetGeneratorTwo(masterKey PRGKey, startId uint32, univSize int, setSize int) SetGenerator {
	aes, err := aes.NewCipher(masterKey[:])
	if err != nil {
		panic(err)
	}
	//for now, assume set size = sqrt(univ_size), change this later! could be problem
	return SetGenerator{
		baseGen:  psetggm.NewSecondGGMSetGeneratorC(univSize, setSize, setSize),
		num:      startId,
		idGen:    aes,
		univSize: univSize,
		setSize:  setSize,
	}
}



func (gen *SetGenerator) Gen(pset *PuncturableSet) {
	gen.gen(pset)

	block := make([]byte, 16)
	out := make([]byte, 16)
	block[0] = 0xBB
	binary.LittleEndian.PutUint32(block[1:], uint32(pset.id))
	gen.idGen.Encrypt(out, block)
	pset.shift = binary.LittleEndian.Uint32(out) % uint32(gen.setSize)

	for i := 0; i < len(pset.elems); i++ {
		pset.elems[i] = int((uint32(pset.elems[i]) + pset.shift) % uint32(gen.univSize))
	}
}


func (gen *SetGenerator) GenTwo(pset *PuncturableSet) {
	//look if need to change to gentwo 
	gen.gentwo(pset)

	//block := make([]byte, 16)
	//out := make([]byte, 16)
	//block[0] = 0xBB
	//binary.LittleEndian.PutUint32(block[1:], uint32(pset.id))
	//gen.idGen.Encrypt(out, block)
	//pset.shift = 0

	//no shift for now!
	
	//change this enumeration, will not work with our implementation as is: think about later
	//how to put this shift so we can genwith fast
	//pset.shift = binary.LittleEndian.Uint32(out) % uint32(gen.setSize)
	//for i := 0; i < len(pset.elems); i++ {
	//	pset.elems[i] = int((uint32(pset.elems[i]) + pset.shift) % uint32(gen.univSize))
	//}


}


func (gen *SetGenerator) GenTwoNoEval(pset *PuncturableSet) {
	//look if need to change to gentwo 
	gen.gentwonoeval(pset)

	block := make([]byte, 16)
	out := make([]byte, 16)
	block[0] = 0xBB
	binary.LittleEndian.PutUint32(block[1:], uint32(pset.id))
	gen.idGen.Encrypt(out, block)
	

	
	//change this enumeration, will not work with our implementation as is: think about later
	//how to put this shift so we can genwith fast
	//pset.shift=0
	pset.shift = binary.LittleEndian.Uint32(out) % uint32(gen.setSize)
	//for i := 0; i < len(pset.elems); i++ {
	//	pset.elems[i] = int((uint32(pset.elems[i]) + pset.shift) % uint32(gen.univSize))
	//}

}






func (gen *SetGenerator) GenWith(val int) (pset PuncturableSet) {
	gen.gen(&pset)

	block := make([]byte, 16)
	seed := make([]byte, 16)
	block[0] = 0xBB
	binary.LittleEndian.PutUint32(block[1:], uint32(pset.id))
	gen.idGen.Encrypt(seed, block)
	pos := binary.LittleEndian.Uint64(seed) % uint64(gen.setSize)
	

	pset.shift = uint32(MathMod(val-pset.elems[pos], gen.univSize))

	for i := 0; i < len(pset.elems); i++ {
		pset.elems[i] = int((uint32(pset.elems[i]) + pset.shift) % uint32(gen.univSize))
	}

	return pset
}



//change LOGIC TO GEN UNTIL WE FIND, or sync shift to our imp: as is generates random new set
//for easy implementation, we can just loop through this current logic and add check
func (gen *SetGenerator) GenWithTwo(val int) (pset PuncturableSet) {
	//why does this work if no pset is passed?
	//not sure but okay

	gen.genwithtwonoeval(&pset, val)

	//block := make([]byte, 16)
	//seed := make([]byte, 16)
	//block[0] = 0xBB
	//binary.LittleEndian.PutUint32(block[1:], uint32(pset.id))
	//gen.idGen.Encrypt(seed, block)
	

	//pset.shift =0


	//notice these shifts, look at how to adapt to our thing later:
	//i think it should be pretty easy just need some (basic) new logic
	
	//pos := binary.LittleEndian.Uint64(seed) % uint64(gen.setSize)
	//pset.shift = uint32(MathMod(val-pset.elems[pos], gen.univSize))

	//for i := 0; i < len(pset.elems); i++ {
	//	pset.elems[i] = int((uint32(pset.elems[i]) + pset.shift) % uint32(gen.univSize))
	//}

	return pset
}



func (gen *SetGenerator) gen(pset *PuncturableSet) {
	pset.univSize = gen.univSize
	pset.setSize = gen.setSize
	if len(pset.elems) != pset.setSize {
		pset.elems = make([]int, pset.setSize)
	}
	var block [16]byte

	for {
		block[0] = 0xAA
		binary.LittleEndian.PutUint32(block[1:], uint32(gen.num))
		pset.id = gen.num
		gen.num++

		gen.idGen.Encrypt(pset.seed[:], block[:])
		gen.baseGen.Eval(pset.seed[:], pset.elems,0)

		if gen.baseGen.Distinct(pset.elems) {
			return
		}
	}
}


func (gen *SetGenerator) gentwo(pset *PuncturableSet) {
	pset.univSize = gen.univSize

	pset.setSize = gen.setSize
	if len(pset.elems) != pset.setSize {
		pset.elems = make([]int, pset.setSize)
	}
	var block [16]byte


	block[0] = 0xAA
	binary.LittleEndian.PutUint32(block[1:], uint32(gen.num))
	pset.id = gen.num
	gen.num++
	gen.idGen.Encrypt(pset.seed[:], block[:])


	block2 := make([]byte, 16)
	out := make([]byte, 16)
	block2[0] = 0xBB
	binary.LittleEndian.PutUint32(block2[1:], uint32(pset.id))
	gen.idGen.Encrypt(out, block2)
	

	shift := binary.LittleEndian.Uint32(out) % uint32(gen.setSize)

	gen.baseGen.Eval(pset.seed[:], pset.elems, shift)
	pset.shift = shift


	return

}

func (gen *SetGenerator) gentwonoeval(pset *PuncturableSet) {
	pset.univSize = gen.univSize
	pset.setSize = gen.setSize
	if len(pset.elems) != pset.setSize {
		pset.elems = make([]int, pset.setSize)
	}
	var block [16]byte

	for {
		block[0] = 0xAA
		binary.LittleEndian.PutUint32(block[1:], uint32(gen.num))
		pset.id = gen.num
		gen.num++

		gen.idGen.Encrypt(pset.seed[:], block[:])
		return
	}
}
func (gen *SetGenerator) genwithtwonoeval(pset *PuncturableSet, index int) {
	pset.univSize = gen.univSize
	pset.setSize = gen.setSize
	if len(pset.elems) != pset.setSize {
		pset.elems = make([]int, pset.setSize)
	}
	var block [16]byte

	
	block[0] = 0xAA
	binary.LittleEndian.PutUint32(block[1:], uint32(gen.num))
	pset.id = gen.num
	gen.num++

	gen.idGen.Encrypt(pset.seed[:], block[:])
	val1 := gen.baseGen.EvalOn(pset.seed[:], index, 0) % pset.setSize //pass in shift of 0 to eval on because we will decide shift later

	val2 := index % pset.setSize
	pset.shift = uint32(MathMod((val2 - val1), pset.setSize))
	//fmt.Println(pset.shift)
	//fmt.Printf("val1: %d, val2: %d, shift: %d \n", val1,val2,pset.shift)
}





func (gen *SetGenerator) Eval(key SetKey) PuncturableSet {
	pset := PuncturableSet{
		SetKey:   key,
		univSize: gen.univSize,
		setSize:  gen.setSize,
		elems:    make([]int, gen.setSize)}

	gen.EvalInPlace(key, &pset)
	return pset
}

func (gen *SetGenerator) EvalTwo(key SetKey) PuncturableSet {
	pset := PuncturableSet{
		SetKey:   key,
		univSize: gen.univSize,
		setSize:  gen.setSize,
		elems:    make([]int, gen.setSize)}


	gen.EvalInPlaceTwo(key, &pset)
	return pset
}




func (gen *SetGenerator) EvalInPlace(key SetKey, pset *PuncturableSet) {
	pset.SetKey = key
	pset.univSize = gen.univSize
	pset.setSize = gen.setSize
	if len(pset.elems) != gen.setSize {
		pset.elems = make([]int, gen.setSize)
	}

	var block [16]byte

	block[0] = 0xAA
	binary.LittleEndian.PutUint32(block[1:], key.id)

	gen.idGen.Encrypt(pset.seed[:], block[:])
	gen.baseGen.Eval(pset.seed[:], pset.elems, 0)

	if key.shift == 0 {
		return
	}
	for i := 0; i < len(pset.elems); i++ {
		pset.elems[i] = int((uint32(pset.elems[i]) + key.shift) % uint32(gen.univSize))
	}
}



func (gen *SetGenerator) EvalInPlaceTwo(key SetKey, pset *PuncturableSet) {
	//maybe need sqrt_univ_size in pset?
	pset.SetKey = key
	pset.univSize = gen.univSize
	pset.setSize = gen.setSize
	if len(pset.elems) != gen.setSize {
		pset.elems = make([]int, gen.setSize)
	}

	var block [16]byte

	block[0] = 0xAA
	binary.LittleEndian.PutUint32(block[1:], key.id)

	gen.idGen.Encrypt(pset.seed[:], block[:])
	gen.baseGen.Eval(pset.seed[:], pset.elems, key.shift)
	//fmt.Println(key.shift)

	if key.shift == 0 {
		return
	}
	//fmt.Println("make sure we never see this in evalinplacetwo")
	//note how shift is done, use that same logic here as well (maybe write helper for this)
	// for i := 0; i < len(pset.elems); i++ {
	// 	pset.elems[i] = int((uint32(pset.elems[i]) + key.shift) % uint32(gen.univSize))
	// }
}

func (gen *SetGenerator) EvalOn(key SetKey, pset *PuncturableSet, idx int) int {
	pset.SetKey = key
	pset.univSize = gen.univSize
	pset.setSize = gen.setSize
	var block [16]byte

	block[0] = 0xAA
	binary.LittleEndian.PutUint32(block[1:], key.id)

	gen.idGen.Encrypt(pset.seed[:], block[:])

	return gen.baseGen.EvalOn(pset.seed[:], idx, pset.shift)

}







func (gen *SetGenerator) Punc(pset PuncturableSet, idx int) PuncturedSet {
	for pos, elem := range pset.elems {
		if elem == idx {
			return PuncturedSet{
				UnivSize: pset.univSize,
				SetSize:  pset.setSize - 1,
				Hole:     pos,
				Shift:    pset.shift,
				Keys:     gen.baseGen.Punc(pset.seed[:], pos)}
		}
	}
	panic(fmt.Sprintf("Failed to find idx: %d in pset: %v", idx, pset.elems))
}

//COMPLETELY Change logic, does not need thing since we can find member easily
//not sure what I meant above, but I think punc remains unchanged after we find set?
func (gen *SetGenerator) PuncTwo(pset PuncturableSet, idx int) PuncturedSet {
	//for pos, elem := range pset.elems {
	//	if elem == idx {
	//		return PuncturedSet{
	//			UnivSize: pset.univSize,
	//			SetSize:  pset.setSize - 1,
	//			Hole:     pos,
	//			Shift:    pset.shift,
	//			Keys:     gen.baseGen.Punc(pset.seed[:], pos)}
	//	}
	//}
	//panic(fmt.Sprintf("Failed to find idx: %d in pset: %v", idx, pset.elems))
	return PuncturedSet {
		UnivSize: pset.univSize,
		SetSize: pset.setSize - 1,
		Hole: 0,
		Shift: pset.shift,
		Keys: gen.baseGen.Punc(pset.seed[:], idx)}
}




func (pset *PuncturedSet) Eval() Set {
	baseGen := psetggm.NewGGMSetGeneratorC(pset.UnivSize, pset.SetSize+1)
	elems := make([]int, pset.SetSize)
	baseGen.EvalPunctured(pset.Keys, pset.Hole, elems)
	for i := 0; i < len(elems); i++ {
		elems[i] = int((uint32(elems[i]) + pset.Shift) % uint32(pset.UnivSize))
	}
	return elems
}
//why does this create a new set?
//here is where we need to pass in the database and such but really having a hard time understanding what's going on
//i think it needs a new baseGen since it does not have access to the baseGen client side (this runs on server)


//commented out because we don't USE! we use fastanswer instead.

// func (pset *PuncturedSet) EvalTwo() Set {
// 	baseGen := psetggm.NewSecondGGMSetGeneratorC(pset.UnivSize, pset.SetSize+1)
// 	elems := make([]int, pset.SetSize)
// 	baseGen.EvalPunctured(pset.Keys, pset.Hole, elems)

// 	//shifts always 0 for now so below not needed?	look into how to add later though
// 	// for i := 0; i < len(elems); i++ {
// 	// 	elems[i] = int((uint32(elems[i]) + pset.Shift) % uint32(pset.UnivSize))
// 	// }
// 	return elems
// }


// Go's % operator follows C semantics and can produce
// negative values if it's given a negative argument.
// We need an arithmetic mod operator.
func MathMod(x int, mod int) int {
	out := x % mod

	// TODO: This is not a constant-time operation.
	if out < 0 {
		out = out + mod
	}

	return out
}
