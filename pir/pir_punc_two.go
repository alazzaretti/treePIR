package pir

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"time"
	"math/rand"

	"checklist/psetggm"
)

type puncTwoClient struct {
	nRows  int
	RowLen int

	setSize int

	sets []SetKey

	hints []Row

	randSource         *rand.Rand
	origSetGen, setGen SetGenerator

	idxToSetIdx []int32
}

type PuncTwoHintReq struct {
	RandSeed           PRGKey
	NumHintsMultiplier int
}

type PuncTwoHintResp struct {
	NRows     int
	RowLen    int
	SetSize   int
	SetGenKey PRGKey
	Hints     []Row
}



var nextHeight []int


func initNextHeight(setSize int) {
	nextHeight = make([]int, setSize)
	psetggm.GetHeightsArr(setSize, nextHeight)
}

func getNextHeight() []int{
	return nextHeight
}

//unchanged
// func xorRowsFlatSlice(db *StaticDB, out []byte, indices Set) {
// 	for i := range indices {
// 		indices[i] *= db.RowLen
// 	}
// 	psetggm.XorBlocks(db.FlatDb, indices, out)

// }
//unchanged (for now)
// func dbElem(db StaticDB, i int) Row {
// 	if i < db.NumRows {
// 		return db.Row(i)
// 	} else {
// 		return make(Row, db.RowLen)
// 	}
// }


//unchanged (fornow)
//NOTE: REMOVED SEC_PARAM NUM hints, readd later
func NewPuncTwoHintReq(randSource *rand.Rand) *PuncTwoHintReq {
	req := &PuncTwoHintReq{
		RandSeed:           PRGKey{},
		NumHintsMultiplier: int(float64(SecParam) * math.Log(2)),
	}
	_, err := io.ReadFull(randSource, req.RandSeed[:])
	if err != nil {
		log.Fatalf("Failed to initialize random seed: %s", err)
	}
	return req
}

//use second setgenerator constructor for different punc, I think rest remains same: DONE

func (req *PuncTwoHintReq) Process(db StaticDB) (HintResp, error) {
	//why do we pick set size like this?
	// sqrt(db_size?)
	//guess it makes sense

	setSize := int(math.Round(math.Pow(float64(db.NumRows), 0.5)))
	
	initNextHeight(setSize)

	nHints := req.NumHintsMultiplier * db.NumRows / setSize
	//fmt.Println(nHints)
	hints := make([]Row, nHints)
	hintBuf := make([]byte, db.RowLen*nHints)
	setGen := NewSetGeneratorTwo(req.RandSeed, 0, db.NumRows, setSize)
	
	var pset PuncturableSet
	for i := 0; i < nHints; i++ {
		setGen.GenTwo(&pset)

		hints[i] = Row(hintBuf[db.RowLen*i : db.RowLen*(i+1)])

		//fmt.Println(i)
		//note -> this fucntion edits pset.elems!
		xorRowsFlatSlice(&db, hints[i], pset.elems)

		//fmt.Println(pset.elems)
	}

	return &PuncTwoHintResp{
		Hints:     hints,
		NRows:     db.NumRows,
		RowLen:    db.RowLen,
		SetSize:   setSize,
		SetGenKey: req.RandSeed,
	}, nil
}



//change set generator to second one, also we would like to
//change this initSets function so as to not incur linear space :DONE
func (resp *PuncTwoHintResp) InitClient(source *rand.Rand) Client {
	c := puncTwoClient{
		randSource: source,
		nRows:      resp.NRows,
		RowLen:     resp.RowLen,
		setSize:    resp.SetSize,
		hints:      resp.Hints,
		origSetGen: NewSetGeneratorTwo(resp.SetGenKey, 0, resp.NRows, resp.SetSize),
	}
	c.initSets()
	return &c
}
//unchanged
func (resp *PuncTwoHintResp) NumRows() int {
	return resp.NRows
}

// no longer needed really since we commented out most of the logic
//will leave in case need to add stuff to set initialization later
//edits DONE
func (c *puncTwoClient) initSets() {

	c.sets = make([]SetKey, len(c.hints))
	// c.idxToSetIdx = make([]int32, c.nRows)
	// for i := range c.idxToSetIdx {
	// 	c.idxToSetIdx[i] = -1
	// }

	var pset PuncturableSet
	for i := 0; i < len(c.hints); i++ {
		
		//does this do sqrt(n) work? origSetGen.Gen? 
		//for original gen it does, we do gennoeval
		//now there is no enumeration of each set
		c.origSetGen.GenTwoNoEval(&pset)
		c.sets[i] = pset.SetKey


		//for _, j := range pset.elems {
		//	c.idxToSetIdx[j] = int32(i)
		//}
	}

	// Use a separate set generator with a new key for all future sets
	// since they must look random to the left server.
	var newSetGenKey PRGKey
	io.ReadFull(c.randSource, newSetGenKey[:])
	c.setGen = NewSetGeneratorTwo(newSetGenKey, c.origSetGen.num, c.nRows, c.setSize)
}

//unchanged, will not use for now, to sample failure prob

func (c *puncTwoClient) bernoulli(nHeads int, total int) bool {
	coin := c.randSource.Intn(total)
	return coin < nHeads
}
func (c *puncTwoClient) sample(odd1 int, odd2 int, total int) int {
	coin := c.randSource.Intn(total)
	if coin < odd1 {
		return 1
	} else if coin < odd1+odd2 {
		return 2
	} else {
		return 0
	}
}

//TODO - change, along with adapting set evaluation and all, 
//WILL NOT WORK LIKE THIS
//uses linear-sized hash map that points indices to sets....
//we need to iterate across all sets
func (c *puncTwoClient) findIndex(i int) (setIdx int) {
	//invalid index = bad query
	if i >= c.nRows {
		return -1
	}
	//removed, used hashmap
	// if setIdx := c.idxToSetIdx[MathMod(i, c.nRows)]; setIdx >= 0 {
	// 	return int(setIdx)
	// }
	
	var pset PuncturableSet
	

	for j := range c.sets {
		setGen := c.setGenForSet(j)
		setKeyNoShift := c.sets[j]
		//below needed if we are shifting sets, right now we are not
		//shift := setKeyNoShift.shift
		setKeyNoShift.shift = 0
		//check just one element of set:
		//specifically need logic to eval only one element of pset
		//probably need to add function to be able to 'evalAt'
		//new function though! :/
		//only for unpunctured sets
		//TODO^:

		//rewrote eval on to work like this, we evaluate the pset at the first log(n)/2 bits of i
		//note that this is not compatible with Punc that does the opposite

		output_index := setGen.EvalOn(setKeyNoShift, &pset, i);
		//fmt.Println(output_index)
		if output_index == i {
			//fmt.Println(j)
			return j
		}
		//dont need to iterate through sets since we can check in o(1) time
		//leaving becasue shift logic might be useful later if we decide to use
		//setGen.EvalInPlace(setKeyNoShift, &pset)
		//for _, v := range pset.elems {
			//shiftedV := int((uint32(v) + shift) % uint32(setGen.univSize))
			//if shiftedV == i {
				//return j
			//}

			// if shiftedV < c.nRows {
			// 	// upgrade invalid pointer to valid one
			// 	c.idxToSetIdx[shiftedV] = int32(j)
			// }
		//}
	}
	return -1
}



type PuncTwoQueryReq struct {
	PuncturedSet PuncturedSet
	ExtraElem    int
}

type PuncTwoQueryResp struct {
	Answer    []byte
	ExtraElem Row
}

type puncTwoQueryCtx struct {
	randCase int
	setIdx   int
	valPos int
}

//query: notes added throughout on what needs changing
//actually i think query is okay... all that's left is fix findIndex


//OPTIMIZE QUERY: We dont need to eval... after finding the set we can just 
//puncture literally

//TODO: Adjust the genwith
func (c *puncTwoClient) Query(i int) ([]QueryReq, ReconstructFunc) {
	if len(c.hints) < 1 {
		panic("No stored hints. Did you forget to call InitHint?")
	}
	var ctx puncTwoQueryCtx
	//obviously findIndex needs to be changed (as stated above)
	ctx.setIdx = c.findIndex(i);

	ctx.valPos = GetPos(i, c.setSize)
	if ctx.setIdx < 0 {
		return nil, nil
	}
	i = MathMod(i, c.nRows)

	
	//stays the same if setIdx is coded consistently with this 
	pset := c.eval(ctx.setIdx)
	
	//logic to pick what set to send where: good for us maybe? will leave for now and
	//if needed I'll hardcode a case for testing
	//punc algorithm is 
	var puncSetL, puncSetR PuncturedSet
	var extraL, extraR int
	ctx.randCase = c.sample(c.setSize-1, c.setSize-1, c.nRows)
	
	//hardcoding for now so that I can test properly: remove later
	//ctx.randCase = 0

	//need to change random member except to not use set since we do not evaluate
	switch ctx.randCase {
	case 0:
		start := time.Now()
		newSet := c.setGen.GenWithTwo(i)
		elapsed := time.Since(start)
		fmt.Printf("genwith took %s \n", elapsed)
		extraL = c.randomMemberExcept(newSet, i)
		extraR = c.randomMemberExcept(pset, i)

		puncSetL = c.setGen.PuncTwo(newSet, i)
		puncSetR = c.setGen.PuncTwo(pset, i)
		if ctx.setIdx >= 0 {
			c.replaceSet(ctx.setIdx, newSet)
		}
	case 1:
		newSet := c.setGen.GenWithTwo(i)
		extraR = c.randomMemberExcept(newSet, i)
		extraL = c.randomMemberExcept(newSet, extraR)
		puncSetL = c.setGen.PuncTwo(newSet, extraR)
		puncSetR = c.setGen.PuncTwo(newSet, i)
	case 2:
		newSet := c.setGen.GenWithTwo(i)
		extraL = c.randomMemberExcept(newSet, i)
		extraR = c.randomMemberExcept(newSet, extraL)
		puncSetL = c.setGen.PuncTwo(newSet, i)
		puncSetR = c.setGen.PuncTwo(newSet, extraL)
	}
	return []QueryReq{
			&PuncTwoQueryReq{PuncturedSet: puncSetL, ExtraElem: extraL},
			&PuncTwoQueryReq{PuncturedSet: puncSetR, ExtraElem: extraR},
		},
		func(resps []interface{}) (Row, error) {
			queryResps := make([]*PuncTwoQueryResp, len(resps))
			var ok bool
			for i, r := range resps {
				if queryResps[i], ok = r.(*PuncTwoQueryResp); !ok {
					return nil, fmt.Errorf("Invalid response type: %T, expected: *PuncTwoQueryResp", r)
				}
			}
			//look at reconstruct?
			//probably needs changes
			return c.reconstruct(ctx, queryResps)
		}
}
//unchanged if we code consistent to it
func (c *puncTwoClient) eval(setIdx int) PuncturableSet {
	if c.sets[setIdx].id < c.origSetGen.num {

		return c.origSetGen.EvalTwo(c.sets[setIdx])
	} else {
		return c.setGen.EvalTwo(c.sets[setIdx])
	}
}
//unchanged if we code consistent to it
func (c *puncTwoClient) setGenForSet(setIdx int) *SetGenerator {
	if c.sets[setIdx].id < c.origSetGen.num {
		return &c.origSetGen
	} else {
		return &c.setGen
	}
}
//mostly unneeded since we no longer groom hashmap,
//but will keep it here for now
func (c *puncTwoClient) replaceSet(setIdx int, newSet PuncturableSet) {
	//old logic to groom hashmap
	//pset := c.eval(setIdx)
	// for _, idx := range pset.elems {
	// 	if idx < c.nRows && c.idxToSetIdx[idx] == int32(setIdx) {
	// 		c.idxToSetIdx[idx] = -1
	// 	}
	// }

	c.sets[setIdx] = newSet.SetKey

	//old logic to groom hashmap
	// for _, v := range newSet.elems {
	// 	c.idxToSetIdx[v] = int32(setIdx)
	// }
}
//I think this is fine to stay, already changed to call new funcs
func (c *puncTwoClient) DummyQuery() []QueryReq {
	newSet := c.setGen.GenWithTwo(0)
	extra := c.randomMemberExcept(newSet, 0)
	puncSet := c.setGen.PuncTwo(newSet, 0)
	q := PuncTwoQueryReq{PuncturedSet: puncSet, ExtraElem: extra}
	return []QueryReq{&q, &q}
}




//definitely need to change this!! to process all root(n) sets
//also see what Fastanswer does, had not noticed it is used here
//change answer to use our new punceval thing!
func (q *PuncTwoQueryReq) Process(db StaticDB) (interface{}, error) {


	//how do iset up to get back sqrt(n) answers????

	resp := PuncTwoQueryResp{Answer: /*make(Row, db.RowLen)}*/make([]byte, (q.PuncturedSet.SetSize+1)*db.RowLen)}



	psetggm.FastAnswerTwo(q.PuncturedSet.Keys, q.PuncturedSet.UnivSize, q.PuncturedSet.SetSize, int(q.PuncturedSet.Shift),
		getNextHeight(),db.FlatDb, db.RowLen, resp.Answer)

	resp.ExtraElem = db.FlatDb[db.RowLen*q.ExtraElem : db.RowLen*q.ExtraElem+db.RowLen]

	return &resp, nil
}


//takes in response and outputs the parity we want
//definitely needs editing, dependent on process and queryreq above
func (c *puncTwoClient) reconstruct(ctx puncTwoQueryCtx, resp []*PuncTwoQueryResp) (Row, error) {
	if len(resp) != 2 {
		return nil, fmt.Errorf("Unexpected number of answers: have: %d, want: 2", len(resp))
	}

	out := make(Row, len(c.hints[0]))
	if ctx.setIdx < 0 {
		return nil, errors.New("couldn't find element in collection")
	}
	//again uses randomness thing explained in paper to not fail
	//make sure this is consistent or hardcode ctx.randcase to test

	//gets me actual index that I am interested within all parities
	realidx := len(c.hints[0])*(ctx.valPos)


	switch ctx.randCase {
	case 0:
		hint := c.hints[ctx.setIdx]
		xorInto(out, hint)
		xorInto(out, resp[Right].Answer[realidx:realidx+16])

		// Update hint with refresh info
		xorInto(hint, hint)
		xorInto(hint, resp[Left].Answer[realidx:realidx+16])
		xorInto(hint, out)

	case 1:
		xorInto(out, out)
		xorInto(out, resp[Left].Answer)
		xorInto(out, resp[Right].Answer)
		xorInto(out, resp[Right].ExtraElem)
	case 2:
		xorInto(out, out)
		xorInto(out, resp[Left].Answer)
		xorInto(out, resp[Right].Answer)
		xorInto(out, resp[Left].ExtraElem)
	}
	//fmt.Println(out)
	return out, nil
}

//is this for testing?not sure where this would be used
func (c *puncTwoClient) NumCovered() int {
	covered := make(map[int]bool)
	for j := range c.sets {
		for _, elem := range c.eval(j).elems {
			covered[elem] = true
		}
	}
	return len(covered)
}

//TODO: Change: pick val from rand source:
//:::::::check if val right shifted is equal to idx shifted
// Sample a random element of the set that is not equal to `idx`.
func (c *puncTwoClient) randomMemberExcept(set PuncturableSet, idx int) int {
	for {
		// TODO: If this is slow, use a more clever way to
		// pick the random element.
		//
		// Use rejection sampling.
		// val := set.elems[c.randSource.Intn(c.setSize)]
		// if val != idx {
		// 	return val
		// }

		//note: can do this in C if this is slow
		height := psetggm.GetHeight(c.setSize)
		val := c.randSource.Intn(c.setSize)
		idx2 := idx >> height
		if val != idx2 {
			return (val << height)
		}
	}
}
func GetPos(idx int, setSize int) int {
	height := psetggm.GetHeight(setSize)
	return idx >> height
}

//fine not to change
func (c *puncTwoClient) StateSize() (bitsPerKey, fixedBytes int) {
	return int(math.Log2(float64(len(c.hints)))), len(c.hints) * c.RowLen
}










