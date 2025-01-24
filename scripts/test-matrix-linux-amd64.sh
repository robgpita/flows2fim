#!/bin/bash

# Implement command line arg that can take controls, fim, validate or nothing (if nothing run all)

# 

echo ">>>> Testing flows2fim controls"
flows2fim controls 
echo "<<<< flows2fim controls tests pass"

echo ">>>> Testing flows2fim fim"
flows2fim fim 
echo "<<<< flows2fim controls tests pass"

echo ">>>> Testing flows2fim validate"
flows2fim validate
echo "<<<< flows2fim validate tests pass"

