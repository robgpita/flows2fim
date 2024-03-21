# Flows2FIM

## Overview

## Purpose of Directories
- `cmd`: Contains executable applications.
- `pkg`: Houses reusable packages potentially useful in other projects.
- `internal`: For private application code not intended for external use.
- `scripts`: Includes useful scripts for building, testing, and more.

## Getting Started

1. Download fim-library from `s3://fimc-data/fim2d/prototype/2024_03_13/` to `testdata/library` folder.

2. Launch a docker container using `docker compose up` and run following commands inside the container

3. Run `go run main.go controls -db=testdata/reach_data.db -f testdata/flows_100yr.csv -c controls.csv -sid 8489318 -scs 0.0` This will create a controls.csv file

4. Run `go run main.go fim -c controls.csv -lib testdata/library -o output.vrt` This will create a VRT file. VRT can be tested by loading in QGIS.