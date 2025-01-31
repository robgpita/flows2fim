#!/bin/bash

# This script tests the built flows2fim with a  combination of parameters 

num_test_cases=4

printf "(1/${num_test_cases})\n\t>>>> Test the help usage statement <<<<\n\n" 
flows2fim --help

printf "(1/${num_test_cases})\n\t>>>> Test the version usage statement <<<<\n\n"
flows2fim --version

printf "###### TEST CASES FOR flows2fim controls  ######"
printf "(1/${num_test_cases})\n\t>>>> Test contorls -db parameter <<<<\n\n"


