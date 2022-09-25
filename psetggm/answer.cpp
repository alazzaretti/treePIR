#include "answer.h"

#include "pset_ggm.h"
#include "xor.h"

#include <vector>

extern "C" {

void answer(const uint8_t* pset, unsigned int pos, unsigned int univ_size, unsigned int set_size, unsigned int shift,
    const uint8_t* db, unsigned int db_len, unsigned int row_len, unsigned int block_len, 
    uint8_t* out) {
    
    auto worksize = workspace_size(univ_size, set_size+1);
    auto workspace = (uint8_t*)malloc(worksize+set_size*sizeof(long long unsigned int));
    auto gen = pset_ggm_init(univ_size, set_size+1, workspace);

    auto elems = (long long unsigned int*)(workspace+worksize);
    pset_ggm_eval_punc(gen, pset, pos, elems);

    for (int i = 0; i < set_size; i++) 
        elems[i] = ((elems[i]+shift)%univ_size);//*row_len;


    xor_rows(db, db_len, elems, set_size, block_len, out, 0);

    free(workspace);
}

void new_answer(uint8_t* pset, unsigned int univ_size, unsigned int set_size, unsigned int shift, const unsigned int* next_height,
    const uint8_t* db, unsigned int db_len, unsigned int row_len, unsigned int block_len, 
    uint8_t* out) {

    auto worksize = new_workspace_size(univ_size, set_size + 1);
    auto workspace = (uint8_t*)malloc(worksize+set_size*sizeof(long long unsigned int));
    
    //should be sqrt_univ_size on second parameter but it should be equal for our implementation so its fine 
    //for most usecases, but can break if called on different set_size \neq sqrt(N)
    auto gen = new_pset_ggm_init(univ_size, set_size+1, set_size+1, workspace);
    auto elems = (long long unsigned int*)(workspace+worksize);


    //todo, fix function to do operation above in a smart way for each xor.
    //might require re-writing xor rows to take in an input 'beginning value'

    new_pset_ggm_eval_punc(gen, pset, shift, elems, next_height, db, db_len, block_len, out);
    free(workspace);

}



} // extern "C"