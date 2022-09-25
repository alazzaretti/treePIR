#include "pset_ggm.h"

#include <chrono> 
#include <cstdint>
#include <iostream>
#include <unordered_set>
#include <vector>
#include <cstring>
#include "intrinsics.h"
#include <iostream>
#include <stdio.h>

using namespace std::chrono; 



int main(int argc, char** argv) {
    enum ARGS {
        PROGRAM_NAME = 0,
        SET_SIZE,
        NUM_ARGS
    };

    if (argc < NUM_ARGS) {
        printf("Usage: %s <SET_SIZE>\n", argv[PROGRAM_NAME]);
        return 1;
    }
    //added 2 factor to univ_size, added sqrt_univ_size var for later section
    uint32_t univ_size = 64*64;//*1024*1024;

    uint32_t sqrt_univ_size = 64;//*1024;
    uint32_t set_size = atoi(argv[SET_SIZE]);

    const uint8_t seed[] = {0, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0};

    unsigned int worksize = workspace_size(univ_size, set_size);
    std::vector<uint8_t> workspace(worksize);

    generator* gen = pset_ggm_init(univ_size, set_size, workspace.data());

    std::vector<long long unsigned int> out(set_size);
    auto start = high_resolution_clock::now(); 
    for (int i=0; i < 100000; ++i) {
        *(int*)(seed) = i;
        pset_ggm_eval(gen, seed, out.data());
    }
    auto stop = high_resolution_clock::now(); 

    std::cout   << "Eval time: " 
                << duration_cast<nanoseconds>(stop - start).count()/100000
                << "ns"
                << std::endl;
    std::cout << "set: ";
    for (int i =0; i < set_size; ++i) {
        std::cout << out[i] << ", ";
    }
    std::cout << std::endl;

    
    unsigned int pset_size = pset_buffer_size(gen);
    //initializes vector pset of size pset_size
    std::vector<uint8_t> pset(pset_size);

    std::vector<long long unsigned int> pelems(set_size);

    for (int pos = 0; pos < set_size; ++pos) {
        pset_ggm_punc(gen, seed, pos, pset.data());
        pset_ggm_eval_punc(gen, pset.data(), pos, pelems.data());
        int compare_pos = 0;
        for (int i =0; i < set_size; ++i) {
            if (i == pos) {
                continue;
            }
            //std::cout  << compare_pos << "th element: " << pelems[compare_pos] << ", ";
            if ((i != pos) && (out[i] != pelems[compare_pos])) {
                std::cout << "Differ: " << out[i] << ", " << pelems[compare_pos] << ". ";
            }
            compare_pos++;
        }
        std::cout << std::endl;
    } 





    ///////////////////////------------------------------------------\\\\\\\\\\\\\\\\\\\\\\\\\\

    //tests using new psets! same args and all, checking if it will work


    unsigned int new_worksize = new_workspace_size(univ_size, set_size);
    std::vector<uint8_t> new_workspace(new_worksize);

    new_generator* new_gen = new_pset_ggm_init(univ_size, sqrt_univ_size, set_size, new_workspace.data());

    std::vector<long long unsigned int> new_out(set_size);
    start = high_resolution_clock::now(); 
    for (int i=0; i < 1/*00000*/; ++i) {
        *(int*)(seed) = i;
        new_pset_ggm_eval(new_gen, seed, new_out.data());
    }
    stop = high_resolution_clock::now(); 

    std::cout   << "Eval time: " 
                << duration_cast<nanoseconds>(stop - start).count()/100000
                << "ns"
                << std::endl;
    std::cout << "set: ";
    for (int i =0; i < set_size; ++i) {
        std::cout << std::bitset<32>(new_out[i]) << ", ";
    }
    std::cout << std::endl;

    unsigned int new_pset_size = new_pset_buffer_size(new_gen);
    std::vector<uint8_t> new_pset(new_pset_size);
    std::vector<long long unsigned int> new_pelems(set_size);


    int pos = set_size - 1;


    new_pset_ggm_punc(new_gen, seed, pos, new_pset.data());



    std::cout << "tree arrays:" <<std::endl;
    uint32_t* tree_arr = new uint32_t[set_size];
    uint32_t tree_height = get_height(set_size);
    uint32_t index = 0;
    get_heights_arr(tree_height,tree_arr, &index);

    // for (int i = 0; i < 50; i++) {
    //     std::cout << tree_arr[i] << ", ";
    // }
    // std::cout << std::endl;
    

    uint32_t* tree_arr2 = new uint32_t[set_size];
    get_heights_wrapper(set_size, tree_arr2);
    // for (int i = 0; i < 256; i++) {
    //     std::cout << tree_arr2[i] << ", ";
    // }
    // std::cout << std::endl;


    char* db = new char[univ_size];
    for (int i = 0; i < univ_size; i++) {
        db[i] = rand();
    }


    pos= 0;

    start = high_resolution_clock::now(); 
    //new_pset_ggm_eval_punc(new_gen, new_pset.data(), pos, new_pelems.data(),tree_arr,db); // PASS IN PARITY POINTER AS WELL to count parities
    stop = high_resolution_clock::now(); 

    std::cout   << "Eval time: " 
                << duration_cast<nanoseconds>(stop - start).count()/100000
                << "ns"
                << std::endl;

    

    // std::cout << "after set: ";
    // for (int i =0; i < set_size-1; ++i) {
    //     std::cout << i<<"th element: " <<std::bitset<32>(new_pelems[i]) << std::endl;
    // }

    pos = set_size - 1;

    // std::cout << std::endl;
    // for (int i =0; i < set_size - 1; ++i) {
    //     std::cout << i<<": "<<(std::bitset<32>(new_pelems[i]) == std::bitset<32>(new_out[i+ ((uint32_t) i >=pos)]))<< std::endl;
    // }
    // std::cout << std::endl;



    //tree generation -- testing creating array for values the heights array that I hardcoded on the fly:
    //this function can be used to generate the array of size set_size at setup and then we dont have to worry about it going out of bounds




    //XOR:: testing things to code to adjust this weird xor function.
    uint8_t test_db[40] = {1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,
        27,28,29,30,31,32,33,34,35,36,37,38,39,40};

    std::cout << "here" << std::endl;
    std::cout << (char) test_db[0] <<std::endl;

    uint8_t* pt_db = ((uint8_t*) (test_db)) +5;
    std::cout << pt_db[0] <<std::endl;





    return 0;
}
