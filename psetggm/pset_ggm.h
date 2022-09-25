#include <stdint.h>
#include "intrinsics.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef struct generator generator;
typedef struct new_generator new_generator;



int get_height(unsigned int v);
unsigned int workspace_size(unsigned int univ_size, unsigned int set_size);
generator* pset_ggm_init(unsigned int univ_size, unsigned int set_size, uint8_t* workspace);
void pset_ggm_eval(generator* gen, const uint8_t* seed, long long unsigned int* elems);

unsigned int pset_buffer_size(const generator* gen);
void pset_ggm_punc(generator* gen, const uint8_t* seed, unsigned int pos, uint8_t* pset);
void pset_ggm_eval_punc(generator* gen, const uint8_t* pset, unsigned int pos, long long unsigned int* elems);

int distinct(generator* gen, const long long unsigned int* elems, unsigned int num_elems);





unsigned int new_workspace_size(unsigned int univ_size, unsigned int set_size);
new_generator* new_pset_ggm_init(unsigned int univ_size, unsigned int sqrt_univ_size, unsigned int set_size, uint8_t* workspace);
void new_pset_ggm_eval(new_generator* gen, const uint8_t* seed, long long unsigned int* elems);

unsigned int new_pset_buffer_size(const new_generator* gen);
long long unsigned new_pset_ggm_eval_on(new_generator* gen, const uint8_t* seed, unsigned int pos, uint8_t* pset);
void new_pset_ggm_punc(new_generator* gen, const uint8_t* seed, unsigned int pos, uint8_t* pset);
void new_pset_ggm_eval_punc(new_generator* gen, uint8_t* pset, unsigned int pos, long long unsigned int* elems, const uint32_t* next_height, 
    const uint8_t* db, unsigned int db_len, unsigned int block_len, 
    uint8_t* out);
void new_pset_ggm_eval_punc_helper(new_generator* gen, const uint8_t* pset, unsigned int pos, uint32_t real_height, long long unsigned int* elems,uint32_t offset,  
    const uint8_t* db, unsigned int db_len, unsigned int block_len, 
    uint8_t* out);
int new_distinct(new_generator* gen, const long long unsigned int* elems, unsigned int num_elems);


void get_heights_arr(uint32_t height, uint32_t* height_arr, uint32_t* index);
void get_heights_wrapper(uint32_t set_size, uint32_t* height_arr);


#ifdef __cplusplus
}
#endif