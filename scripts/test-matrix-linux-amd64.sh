#! /bin/bash

# Implement command line arg that can take controls, fim, validate or nothing (if nothing run all)

flows_file=[ a.csv, b.txt , c]

echo "Testing flows2fim controls"

for file in ${@flows_file}
    flows2fim controls -f ${i}

echo "flows2fim controls test pass"


flows2fim fim 

flows2fim validate


