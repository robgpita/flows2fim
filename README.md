# Flows2FIM

## Overview

## Purpose of Directories
- `cmd`: Contains executable applications.
- `pkg`: Houses reusable packages potentially useful in other projects.
- `internal`: For private application code not intended for external use.
- `scripts`: Includes useful scripts for building, testing, and more.

## Getting Started

`docker run -it -v /home/abdul.siddiqui/workbench/repos/fim_utilities/utils/flows2inundation/data_ops:/app -v /home/abdul.siddiqui/data/outputs:/data --workdir /app golang:1.22.1`

`go run main.go controls -db=reach_data.db -f flows_huc12.csv -c outputs.csv -sid 8489318 -scs 0.0`

`flows2fim fim -c outputs.csv -lib /vsis3/fimc-data/fim2d/prototype/2024_03_13/ -o output.vrt -rel=false`