#include "AES.h"
#include "pset_ggm.h"
#include "xor.h"
#include <vector>
#include <cstdint>
#include <iostream>
#include <stdio.h>

extern "C" {



int get_height(unsigned int v) {
    unsigned int r = 0; // r will be lg(v)
    --v;

    while (v >>= 1) 
    {
        r++;
    }
    return r+1;
}

typedef struct generator {
    unsigned int univ_size, set_size;
    __m128i *keys, *tmp;
} generator;

typedef struct new_generator {
    unsigned int univ_size, sqrt_univ_size, set_size;
    __m128i *keys, *tmp, *prp_key;
} new_generator;


unsigned int workspace_size(unsigned int univ_size, unsigned int set_size) {
    uint32_t height = get_height(set_size); 
    return sizeof(generator) + 2*(1<<height)*sizeof(__m128i)+32;
}


unsigned int new_workspace_size(unsigned int univ_size, unsigned int set_size) {
    uint32_t height = get_height(set_size);
    return sizeof(new_generator) + 2*(1<<height)*sizeof(__m128i)+32;
}





generator* pset_ggm_init(unsigned int univ_size, unsigned int set_size, uint8_t* workspace) {
    auto gen = (generator*)workspace;
    gen->univ_size = univ_size;
    gen->set_size = set_size;
    gen->keys = (__m128i*)(workspace + sizeof(generator));
    // align pointer
    gen->keys = (__m128i*)((((size_t)gen->keys-1)/16+1)*16);

    uint32_t height = get_height(set_size); 
    gen->tmp = gen->keys + (1<<height);

    return gen;
}
//changed to take in universe size and sqaure root of universe size... could do it in function but easier like this
// assume univ size is a perfect square 
new_generator* new_pset_ggm_init(unsigned int univ_size, unsigned int sqrt_univ_size, unsigned int set_size, uint8_t* workspace) {
    auto gen = (new_generator*) workspace;
    gen->univ_size = univ_size;
    gen->sqrt_univ_size = sqrt_univ_size;
    gen->set_size = set_size;
    gen->keys = (__m128i*)(workspace + sizeof(new_generator));
    // align pointer
    gen->keys = (__m128i*)((((size_t)gen->keys-1)/16+1)*16);


    uint32_t height = get_height(set_size); 
    gen->prp_key = gen->keys + (1<<height);

    gen->tmp = gen->prp_key + 1;

    return gen;
}




const __m128i one = _mm_setr_epi32(0, 0, 0, 1);

inline void expand(const __m128i& in, __m128i* out) {
    out[1] = _mm_xor_si128(in, one);
    mAesFixedKey.encryptECB(in, out[0]);
    mAesFixedKey.encryptECB(in, out[1]);
    out[0] = _mm_xor_si128(in, out[0]);
    out[1] = _mm_xor_si128(in, out[1]);
    out[1] = _mm_xor_si128(out[1], one);
}


void tree_eval_all(generator* gen, __m128i seed, long long unsigned int* out) {	
    uint32_t key_pos = 0;	
    uint32_t max_height = get_height(gen->set_size); 	
    uint32_t height = max_height;	
    // std::vector<__m128i> path_key(2*(max_height+1));	
    // path_key[0] = seed;	
    __m128i* keys = gen->keys;

    _mm_store_si128(keys, seed);

    uint32_t node = 0;	
    while (true) {	
        if (height == 0) {	
            out[node] = *(uint32_t*)(&keys[key_pos]) % gen->univ_size;	
            bool is_right = true;	
            // while 'is right child', go up	
            while ((node&1) == 1) {	
                ++height;	
                key_pos -= 1;	
                node >>= 1;	
            }	
            if (height >= max_height) {	
                return;	
            }	
            // move to right sibling	
            node += 1;	
            key_pos -= 1;	

            if ((node << height) >=  gen->set_size) {	
                return;	
            }	

            continue;	
        }	
        expand(keys[key_pos], &keys[key_pos+1]);	
        node <<= 1;	
        --height;	
        // first go to left child	
        key_pos += 2;	
    }	
}

void tree_eval_all2(generator* gen, __m128i seed, long long unsigned int* elems) {
    uint32_t max_depth = get_height(gen->set_size); 
    int num_layers = max_depth - 2;

    
    __m128i* keys = gen->keys;
    __m128i* tmp =  gen->tmp;


    _mm_store_si128(keys, seed);

    for (int depth = 0; depth < num_layers; depth++) {
        for (int i = 0; i < 1<<depth; i++) {
            __m128i key = _mm_load_si128(keys+i);
            _mm_store_si128(tmp + 2*i, key);
            key = _mm_xor_si128(key, one);
            _mm_store_si128(tmp + 2*i + 1, key);
        }
        mAesFixedKey.encryptECBBlocks(tmp, 1<<(depth+1), keys);
        for (int i = 0; i < 1<<(depth+1); i++) {
            __m128i key = _mm_load_si128(tmp+i);
            keys[i] = _mm_xor_si128(keys[i], key);
        }
    }

    const uint32_t* keys_as_elems = (uint32_t*)gen->keys;
    for (int i = 0; i < gen->set_size; i++) {
        //https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
        elems[i] =  ((uint64_t)keys_as_elems[i] * (uint64_t) gen->univ_size) >> 32; //gen->keys[i] % gen->univ_size; //
    }
}

//uses new approach we present in paper
//need to change so that each aes eval points to one element, not 4, or else it breaks
//this is not true as shown by implementation below
void new_tree_eval_all2(new_generator* gen, __m128i seed, long long unsigned int* elems, uint32_t val_shift) {
    uint32_t max_depth = get_height(gen->set_size); 
    int num_layers = max_depth - 2;

    
    __m128i* keys = gen->keys;
    __m128i* tmp =  gen->tmp;


    _mm_store_si128(keys, seed);

    for (int depth = 0; depth < num_layers; depth++) {
        for (int i = 0; i < 1<<depth; i++) {
            __m128i key = _mm_load_si128(keys+i);
            _mm_store_si128(tmp + 2*i, key);
            key = _mm_xor_si128(key, one);
            _mm_store_si128(tmp + 2*i + 1, key);
        }
        mAesFixedKey.encryptECBBlocks(tmp, 1<<(depth+1), keys);
        for (int i = 0; i < 1<<(depth+1); i++) {
            __m128i key = _mm_load_si128(tmp+i);
            keys[i] = _mm_xor_si128(keys[i], key);
        }
    }
    uint32_t real_height = get_height(gen->sqrt_univ_size);
   
    const uint32_t* keys_as_elems = (uint32_t*)gen->keys;
    for (int i = 0; i < gen->set_size; i++) {
        //https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/

        //here we concatenate so that elems[i] = i || F(i) for F with range 0,sqrt_univ_size, we assume set_size <= sqrt_univ_size <= 2^16
        //missing one step: sending current elems[i] through a PRP
        //shift of i and shift of othter part needs to be proportional to set size or we get out of bounds.
        //current code assumes set size is a power of two
        uint32_t mod_out = (((uint64_t)keys_as_elems[i] * (uint64_t) gen->sqrt_univ_size) >> 32);
        mod_out = (mod_out + val_shift) % (gen->sqrt_univ_size);
        
        elems[i] =  ((uint32_t) (i << real_height)) ^ (mod_out); //gen->keys[i] % gen->sqrt_univ_size;
        //std::cout << elems[i] << std::endl;
    }
}




void pset_ggm_eval(generator* gen, const uint8_t* seed, long long unsigned* elems) {
    tree_eval_all2(gen, toBlock(seed), elems);
}

void new_pset_ggm_eval(new_generator* gen, const uint8_t* seed, long long unsigned* elems, uint32_t val_shift) {
    new_tree_eval_all2(gen, toBlock(seed), elems, val_shift);
}


unsigned int pset_buffer_size(const generator* gen) {
    uint32_t height = get_height(gen->set_size); 
    if (height < 2) {
        return sizeof(__m128i);
    }
    return sizeof(__m128i)*(height-1);
}


unsigned int new_pset_buffer_size(const new_generator* gen) {
    uint32_t height = get_height(gen->set_size); 
    if (height < 2) {
        return sizeof(__m128i);
    }
    return sizeof(__m128i)*(height-1);
}



//fix: no need to save keys or anything....
long long unsigned new_pset_ggm_eval_on(new_generator* gen, const uint8_t* seed, unsigned int pos, uint8_t* pset, uint32_t val_shift) {

    __m128i* pset_keys = (__m128i*)pset;

    __m128i* keys = (__m128i*)gen->keys;
    __m128i* tmp = (__m128i*)gen->tmp;
    __m128i key = toBlock(seed);

    int depth = 0;
    uint32_t height = get_height(gen->set_size); 
    uint32_t real_height = height;
    pos = pos >> real_height;

    while (height > 2) {
        _mm_store_si128(tmp, key);
        key = _mm_xor_si128(key, one);
        _mm_store_si128(tmp + 1, key);
        mAesFixedKey.encryptECBBlocks(tmp, 2, keys);
        keys[1] = _mm_xor_si128(keys[1], key);
        key = _mm_xor_si128(key, one);
        keys[0] = _mm_xor_si128(keys[0], key);

        if ((pos & (1<<(height-1))) != 0) {
            //pset_keys[depth] = keys[0];
            key = keys[1];
        } else {
            //pset_keys[depth] = keys[1];
            key = keys[0];
        }
        depth++;
        height--;
    }
    //pset_keys[depth] = key;
    uint32_t* last_key = (uint32_t*)&key;//pset_keys[depth];

    //std::cout << last_key[0] <<std::endl;
    //std::cout << last_key[1] <<std::endl;
    //std::cout << last_key[2] <<std::endl;
    //std::cout << last_key[3] <<std::endl;
    uint32_t pos_idx = (pos & 0b11);
    //hardcoded to only suit 16bit i
    uint32_t mod_out = (((uint64_t)last_key[pos_idx] * (uint64_t) gen->sqrt_univ_size) >> 32);
    mod_out = (mod_out + val_shift) % (gen->sqrt_univ_size);
    //std::cout << "mod_out: "<< mod_out << std::endl;
    //std::cout << pos_idx<<", "<<last_key[0]<<", "<<mod_out <<", " << pos<<std::endl;
    return ((uint32_t) (pos << real_height)) ^ (mod_out);
}


void pset_ggm_punc(generator* gen, const uint8_t* seed, unsigned int pos, uint8_t* pset) {
    __m128i* pset_keys = (__m128i*)pset;

    __m128i* keys = (__m128i*)gen->keys;
    __m128i* tmp = (__m128i*)gen->tmp;
    __m128i key = toBlock(seed);

    int depth = 0;
    uint32_t height = get_height(gen->set_size); 

    while (height > 2) {
        _mm_store_si128(tmp, key);
        key = _mm_xor_si128(key, one);
        _mm_store_si128(tmp + 1, key);
        mAesFixedKey.encryptECBBlocks(tmp, 2, keys);
        keys[1] = _mm_xor_si128(keys[1], key);
        key = _mm_xor_si128(key, one);
        keys[0] = _mm_xor_si128(keys[0], key);

        if ((pos & (1<<(height-1))) != 0) {
            pset_keys[depth] = keys[0];
            key = keys[1];
        } else {
            pset_keys[depth] = keys[1];
            key = keys[0];
        }
        depth++;
        height--;
    }
    pset_keys[depth] = key;
    uint32_t* last_key = (uint32_t*)&pset_keys[depth];

    //std::cout << last_key[0] <<std::endl;
    //std::cout << last_key[1] <<std::endl;
    //std::cout << last_key[2] <<std::endl;
    //std::cout << last_key[3] <<std::endl;
    switch (pos & 0b11) {
        case 0:
            last_key[0] = 0;
            break;
        case 1:
            last_key[1] = 0;
            break;
        case 2:
            last_key[2] = 0;
            break;
        case 3:
            last_key[3] = 0;
            break;
    }
}

//they save one key per 'depth', we need to save an ordered array of keys from left to right
// one way to make this change is begin setting up space for the whole arraw
//and setting up two pointers (left,right), l = 0, r = depth of tree
//if pos is to the left of the current space, we save right key to arr[r] and set r -= 1
//if pos is to the right we save left key to arr[l] and set l += 1
//like this we can build the ordered array on the fly
//we don't worry about sending pos (in protocol they send it separate, we just change there to not send)


//NOTE!!! we actually are sent a value - not a pos, the name is misleading
//right now we puncture at the position where that value would be (without checking if it actually is)
//this can be modified without much overhead but for now now necessary
void new_pset_ggm_punc(new_generator* gen, const uint8_t* seed, unsigned int pos, uint8_t* pset) {

    __m128i* pset_keys = (__m128i*)pset;

    __m128i* keys = (__m128i*)gen->keys;
    __m128i* tmp = (__m128i*)gen->tmp;
    __m128i key = toBlock(seed);

    int depth = 0;
    uint32_t height = get_height(gen->set_size); 
    //changing value to pos:
    pos = (pos >> height);
    uint32_t left_point = 1;
    uint32_t right_point = height - 2;
    //std::cout << left_point << "|" << right_point <<std::endl;

    while (height > 2) {
        //store key @ address temp
        _mm_store_si128(tmp, key);
        //xor key with one, note one = 0,0,0,1
        key = _mm_xor_si128(key, one);
        //store (key xor one) at temp+1
        _mm_store_si128(tmp + 1, key);
        //encrypt key and (key xor one), save to keys
        mAesFixedKey.encryptECBBlocks(tmp, 2, keys);
        //let keys[1] = keys[1] xor key xor one
        keys[1] = _mm_xor_si128(keys[1], key);
        key = _mm_xor_si128(key, one);
        //let keys[0] = keys[0] xor key
        keys[0] = _mm_xor_si128(keys[0], key);

        

        //check if height-th bit of pos is 1
        //if yes save left key, recurse on right
        //if no, save right key, recurse on left
        //we change way of saving: if height-th bit of pos is 1, save key on left-most unfilled index
        //otherwise, sabe on right-most unfilled index
        if ((pos & (1<<(height-1))) != 0) {
            //pset_keys[depth] = keys[0];
            pset_keys[left_point++] = keys[0];
            key = keys[1];
        } else {
            //pset_keys[depth] = keys[1];
            pset_keys[right_point--] = keys[1];
            key = keys[0];
        }
        depth++;
        height--;
        //std::cout << left_point << "|" << right_point <<std::endl;
    }
   
    //pset_keys[depth] = key;
    //uint32_t* last_key = (uint32_t*)&pset_keys[depth];

    //change final step above to be fill at certain index: but actually poses problem, we didn't want this to be anywhere
    //have to think about it
    //I don't think we can work around it:
    //need to change all enumeration code to account for this
    //we can work around it: need to give 'punctured element' separate - and in order without the 'hole'
    //putting it on the first element (so that it is easy to iterate)
    pset_keys[0] = key;
    uint32_t* last_key = (uint32_t*)&pset_keys[0];


    // for (int i = 0; i < get_height(gen->set_size); i++) {
    //     uint32_t* iterative_key = (uint32_t*)&pset_keys[i];
        
    //     std::cout << iterative_key[2] <<std::endl;
    // }

    //what is this doing?
    //literally each key stores 4 elements! makes sense since 128 bit keys and 32 bit elements.
    // What does this mean for the tree?
    //specifically how does it change the tree's depth? log_4? how does it change our approach?
    //put 0 in front always:
    //on eval use same idea as tree, items are ordered, pos could be any of 4 places (0 starts in front)
    switch (pos & 0b11) {
        case 0:
            last_key[0] = 0;
            break;
        case 1:
            last_key[1] = last_key[0];
            last_key[0] = 0;
            break;
        case 2:
            last_key[2] = last_key[1];
            last_key[1] = last_key[0];
            last_key[0] = 0;
            break;
        case 3:
            last_key[3] = last_key[2];
            last_key[2] = last_key[1];
            last_key[1] = last_key[0];
            last_key[0] = 0;
            break;
    }

    // for (int i = 0; i < get_height(gen->set_size); i++) {
    //     uint32_t* iterative_key = (uint32_t*)&pset_keys[i];
    //     for (int j = 0; j < 4; j++)
    //     std::cout << iterative_key[j] <<std::endl;
    // }
}




void pset_ggm_eval_punc(generator* gen, const uint8_t* pset, unsigned int pos, long long unsigned int* elems) {
    uint32_t height = get_height(gen->set_size); 

    __m128i* keys = gen->keys;
    __m128i* tmp =  gen->tmp;

    const __m128i* pset_keys = (const __m128i*)pset;
    
    int depth = 0;
    while (height > 2) {
        for (int i = 0; i < 1<<depth; i++) {
            __m128i key = _mm_load_si128(keys+i);
            _mm_store_si128(tmp + 2*i, key);
            key = _mm_xor_si128(key, one);
            _mm_store_si128(tmp + 2*i + 1, key);
        }
        mAesFixedKey.encryptECBBlocks(tmp, 1<<(depth+1), keys);
        for (int i = 0; i < 1<<(depth+1); i++) {
            __m128i key = _mm_load_si128(tmp+i);
            keys[i] = _mm_xor_si128(keys[i], key);
        }
        height--;
        keys[(pos >> height)^1] = pset_keys[depth];
        depth++;
    }
    
    keys[(pos >> height)] = pset_keys[depth];
    
    size_t out_pos = 0;
    uint32_t* keys_as_elems = (uint32_t*)keys;
    for (int i = 0; i < gen->set_size; i++) {
        if (i == pos) {
            continue;
        }
        //https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
        elems[out_pos] = ((uint64_t)keys_as_elems[i] * (uint64_t) gen->univ_size) >> 32; 
        out_pos++;
    }  
}


//how to code function that keeps -> maybe split in three parts: 
//      1: evaluate first set where punc index = i (initially 0)
//      2: evaluate parity of set and iterate through four possibilities within key
//      3: iterate punctured key placement, adding and removing relevant elements, go from 0...root(n)-1
void new_pset_ggm_eval_punc(new_generator* gen, uint8_t* pset, unsigned int val_shift, long long unsigned int* elems, const uint32_t* next_height, 
    const uint8_t* db, unsigned int db_len, unsigned int block_len, 
    uint8_t* out) {
    

    uint32_t pos=0;
    uint32_t height = get_height(gen->set_size); 
    __m128i* keys = gen->keys;
    __m128i* tmp =  gen->tmp;

    __m128i* pset_keys = (__m128i*)pset;
    


    int depth = 0;
    while (height > 2) {
        for (int i = 0; i < 1<<depth; i++) {
            __m128i key = _mm_load_si128(keys+i);
            _mm_store_si128(tmp + 2*i, key);
            key = _mm_xor_si128(key, one);
            _mm_store_si128(tmp + 2*i + 1, key);
        }
        mAesFixedKey.encryptECBBlocks(tmp, 1<<(depth+1), keys);
        for (int i = 0; i < 1<<(depth+1); i++) {
            __m128i key = _mm_load_si128(tmp+i);
            keys[i] = _mm_xor_si128(keys[i], key);
        }
        height--;

        //TODO: fix this index - always put last key (figure out how) DONE!
        keys[(pos >> height)^1] = pset_keys[height-1];
        depth++;
    }
    //TODO:fix this: always put first key (figure out how) DONE!
    keys[(pos >> height)] = pset_keys[height-2];

    //std::cout << "final index: " << height - 2 <<std::endl;
    //Output: 0
    uint32_t real_height = get_height(gen->sqrt_univ_size);
    size_t out_pos = 0;
    uint32_t* keys_as_elems = (uint32_t*)keys;
    for (int i = 0; i < gen->set_size; i++) {
        if (i == pos) {
            continue;
        }
        //https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
        // elems[out_pos] = ((uint64_t)keys_as_elems[i] * (uint64_t) gen->univ_size) >> 32;


        uint32_t mod_out = (((uint64_t)keys_as_elems[i] * (uint64_t) gen->sqrt_univ_size) >> 32);
        mod_out = (mod_out + val_shift) % (gen->sqrt_univ_size);
        //std::cout << pos_idx<<", "<<last_key[0]<<", "<<mod_out <<", " << pos<<std::endl;
        elems[out_pos] =  ((uint32_t) (i << real_height)) ^ (mod_out);//((uint64_t) i << 16) ^ (((uint64_t)keys_as_elems[i] * (uint64_t) gen->sqrt_univ_size) >> 32); //gen->keys[i] % gen->sqrt_univ_size;
        //std::cout << elems[out_pos] << std::endl;
        out_pos++;
    }  
    //std::cout << (gen->sqrt_univ_size) << ", "<< (gen->set_size) <<std::endl;
    //we have at this point: set evaluated at puncture point = 0

    //TODO: save parity??? - DONE
    uint8_t* tmp_pointer = out;
    

    xor_rows(db, db_len, elems, (gen->set_size)-1, block_len, tmp_pointer, 0);

   
    //iterate for puncture = 0,1,2,3 - DONE below

    tmp_pointer = tmp_pointer+block_len;
    //std::cout << "block_len: "<< block_len <<std::endl;
    //return;
    for (int i = 0; i < 3; i++) {
        //TODO: FIX -> hardcoded for max 16 for log(sqrt(n))



        //shift does not change logic below:
        //remove current first half of eval and switch for new first half (another i)

        xor_rows(db, db_len, elems + i, 1, block_len, tmp_pointer, 1);

        elems[pos + i] =  (elems[pos+i] & ((gen->set_size)-1)) ^ ((pos + i) << real_height);
        
        xor_rows(db, db_len, elems+ i, 1, block_len, tmp_pointer, 2);
        // std::cout << tmp_pointer[0] << std::endl;
        //TODO: get parity, save (each step in loop represents different potential set)
        //std::cout << std::bitset<32>(elems[pos+i]) <<std::endl;
        tmp_pointer += block_len;
    }
    //printf("tmp_pointer: %p", (uint8_t*) tmp_pointer);

    //then: iterate over puncture = other points:
    //increment pos by 4
    //switch key ordering appropriately: in specific, we begin loop above at height = 2
    //then after we switch to height = 3
    //then we go to height = 2 again 

    

    pos += 4;
    int curr_point = 0;
    new_generator* tmp_gen = new new_generator(*gen);
    //int next_height[]= {3,4,3,5,3,4,3,6,3,4,3,5,3,4,3,7,3,4,3,5,3,4,3,6,3,4,3,5,3,4,3,8,3,4,3,5,3,4,3,6,3,4,3,5,3,4,3,7,3,4,3,5,3,4,3,6,3,4,3,5,3,4,3,9,
    //                      3,4,3,5,3,4,3,6,3,4,3,5,3,4,3,7,3,4,3,5,3,4,3,6,3,4,3,5,3,4,3,8,3,4,3,5,3,4,3,6,3,4,3,5,3,4,3,7,3,4,3,5,3,4,3,6,3,4,3,5,3,4,3};
    while (pos < gen->set_size) {
        // std::cout << "pos: " << pos <<std::endl;
        //swap keys for heights more than 3
        // as worked out in notebook -> this converts a order arrangement to a depth-first arrangement
        //to suit the function
        //allows us to recurse

        uint32_t swap_height = next_height[curr_point];
        //std::cout << "pos: " << pos << std::endl;
        //std::cout << "swap_height: " <<swap_height << std::endl;

        //organize keys in correct fashion to 'recurse'
        if (swap_height > 3) {
            //std::cout << "swapping keys" << std::endl;
            //swap key indices: pset_keys[swap_height - 2] <-> pset_keys[swap_height - 3]
            //std::swap<__m128i>(pset_keys+swap_height - 2, pset_keys + swap_height - 3);
            __m128i tmp_val = pset_keys[swap_height - 2];
            pset_keys[swap_height - 2] = pset_keys[swap_height - 3];
            pset_keys[swap_height - 3] = tmp_val;
            int i = 1;
            int j = swap_height -4;
            while (i < j) {
                //std::cout << "2nd swap" << std::endl;
                tmp_val = pset_keys[i];
                pset_keys[i] = pset_keys[j];
                pset_keys[j] = tmp_val;
                j-=1;
                i+=1;
            }
        }


        //recurse on the eval function: send only the first swap_height-1 keys,
        //this means that it will change 2^(swap_height) - 1 elements
        //the elements that it will change is from [pos- 2^(swap_height - 1)] up to [pos + 2^(swap_height-1) + 1]
        
        uint32_t offset = 1<<(swap_height -1);
        //std::cout << "offset: " << offset << std::endl;
        tmp_gen->set_size = offset << 1;
        //std::cout << "set_size: " << tmp_gen->set_size << std::endl;

        uint32_t real_off = pos - offset;

        //TODO: before recursing: we can do the following:
        //      -'remove' affected elements from parity count
        //      -after recursing, we add affected elements to parity count, save new parity (/parities)
        



        //note that pset does not need to change since set size defines amount of keys to be used and
        //we always want to use first keys (we swap accordingly for this)
        //to recurse we need to 'adjust pos' and also send an offset for the function to correctly
        //append the correct 'i'
        //std::cout << "new offset: " <<real_off <<", offsetted pos: " <<pos - real_off << std::endl;
        new_pset_ggm_eval_punc_helper(tmp_gen, pset,pos - real_off, val_shift, real_height, elems+real_off, real_off,db, db_len, block_len, tmp_pointer);

        //TODO: re-evaluate parities (only affected elements), and then iterate over 3 other pos indices 

        tmp_pointer += block_len;
        for (int i = 0; i < 3; i++) {
        //TODO: FIX -> hardcoded for max 16 size for log(sqrt(n))
            xor_rows(db, db_len, elems + pos + i, 1, block_len, tmp_pointer, 1);
            elems[pos + i] =  (elems[pos+i] & ((gen->set_size)-1)) ^ ((pos + i) << real_height);//(elems[pos+i] & (0b00000000000000001111111111111111)) ^ ((pos + i) << 16);
            xor_rows(db, db_len, elems + pos + i, 1, block_len, tmp_pointer, 2);
            //TODO: get parity, save (each step in loop represents different potential set)
            //std::cout << std::bitset<32>(elems[pos+i]) <<std::endl;
            tmp_pointer += block_len;
        }


        pos +=4;
        curr_point+=1;
    }
    //std::cout << "finish evalpunc loop" << std::endl;
}




void new_pset_ggm_eval_punc_helper(new_generator* gen, const uint8_t* pset, unsigned int pos, uint32_t val_shift, uint32_t real_height, long long unsigned int* elems,uint32_t offset,  
    const uint8_t* db, unsigned int db_len, unsigned int block_len, 
    uint8_t* out) {
    uint32_t height = get_height(gen->set_size); 
    __m128i* keys = gen->keys;
    __m128i* tmp =  gen->tmp;

    const __m128i* pset_keys = (const __m128i*)pset;

    int depth = 0;
    while (height > 2) {
        for (int i = 0; i < 1<<depth; i++) {
            __m128i key = _mm_load_si128(keys+i);
            _mm_store_si128(tmp + 2*i, key);
            key = _mm_xor_si128(key, one);
            _mm_store_si128(tmp + 2*i + 1, key);
        }
        mAesFixedKey.encryptECBBlocks(tmp, 1<<(depth+1), keys);
        for (int i = 0; i < 1<<(depth+1); i++) {
            __m128i key = _mm_load_si128(tmp+i);
            keys[i] = _mm_xor_si128(keys[i], key);
        }
        height--;

        //TODO: fix this index - always put last key (figure out how) DONE!
        keys[(pos >> height)^1] = pset_keys[height-1];
        depth++;
    }
    //TODO:fix this: always put first key (figure out how) DONE!
    keys[(pos >> height)] = pset_keys[height-2];

    //std::cout << "final index: " << height - 2 <<std::endl;
    //Output: 0
    size_t out_pos = 0;
    uint32_t* keys_as_elems = (uint32_t*)keys;
    xor_rows(db, db_len, elems, (gen->set_size)-1, block_len, out, 1);
    for (int i = 0; i < gen->set_size; i++) {
        if (i == pos) {
            continue;
        }
        //https://lemire.me/blog/2016/06/27/a-fast-alternative-to-the-modulo-reduction/
        // elems[out_pos] = ((uint64_t)keys_as_elems[i] * (uint64_t) gen->univ_size) >> 32;
        //TODO: note that we use sqrt of real univ size real, hopefully we kept it and never change it from
        //the other one so that it is actually different
        //actually thinking about htis, maybe this is why we dont get full randomness? should I fix it to 16?
        
        uint32_t mod_out = (((uint64_t)keys_as_elems[i] * (uint64_t) gen->sqrt_univ_size) >> 32);
        mod_out = (mod_out + val_shift) % (gen->sqrt_univ_size);
        elems[out_pos] =  ((uint32_t) ((offset+i) << real_height)) ^ (mod_out);
        
        //elems[out_pos] =  ((uint64_t) (offset+i) << 16) ^ (((uint64_t)keys_as_elems[i] * (uint64_t) gen->sqrt_univ_size) >> 32); //gen->keys[i] % gen->sqrt_univ_size;
        
        out_pos++;
    } 
    xor_rows(db, db_len, elems, (gen->set_size)-1, block_len, out, 2);
}








inline unsigned int fasthash(unsigned int elem, unsigned int range) {    
    return elem & (range-1); 
}

inline int round_to_power_of_2(unsigned int v) {
    v--;
    v |= v >> 1;
    v |= v >> 2;
    v |= v >> 4;
    v |= v >> 8;
    v |= v >> 16;
    v++;
    return v;
}

int distinct(generator* gen, const long long unsigned int* elems, unsigned int num_elems)
{   
    uint32_t* table = (uint32_t*)gen->tmp;
    int table_size = round_to_power_of_2(num_elems*4);

    for (int i = 0; i < table_size; i++) {
        table[i] = 0;
    }
    const uint32_t* end = table + table_size;

    for (int i = 0; i < num_elems; i++) {
        auto e = elems[i];
        unsigned int h = fasthash(e, table_size);
        uint32_t* ptr = table + h;

        for (;;) {
            const auto val = *ptr;
            if (val == 0) {
                *ptr = e;
                break;
            }
            if (val == e) {
                return false;
            }
            if (++ptr >= end) {
                ptr = table;
            } 
        }
    }
    return true;
}

int new_distinct(new_generator* gen, const long long unsigned int* elems, unsigned int num_elems)
{   
    uint32_t* table = (uint32_t*)gen->tmp;
    int table_size = round_to_power_of_2(num_elems*4);

    for (int i = 0; i < table_size; i++) {
        table[i] = 0;
    }
    const uint32_t* end = table + table_size;

    for (int i = 0; i < num_elems; i++) {
        auto e = elems[i];
        unsigned int h = fasthash(e, table_size);
        uint32_t* ptr = table + h;

        for (;;) {
            const auto val = *ptr;
            if (val == 0) {
                *ptr = e;
                break;
            }
            if (val == e) {
                return false;
            }
            if (++ptr >= end) {
                ptr = table;
            } 
        }
    }
    return true;
}


void get_heights_wrapper(uint32_t set_size, uint32_t* height_arr) {
    uint32_t height = get_height(set_size);
    uint32_t index = 0;
    get_heights_arr(height, height_arr, &index);
}

void get_heights_arr(uint32_t height, uint32_t* height_arr, uint32_t* index) {
        if (height == 3) {
            height_arr[*index] = height;
            *index = *index + 1;
            return;
        }
        get_heights_arr(height - 1, height_arr, index);
        height_arr[*index] = height;
        *index = *index + 1;
        get_heights_arr(height - 1, height_arr, index);
        return;
}




} // extern "C"
