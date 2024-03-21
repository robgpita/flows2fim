package fim

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// CheckGDALBuildVRTAvailable checks if gdalbuildvrt is available in the environment.
func CheckGDALBuildVRTAvailable() bool {
	cmd := exec.Command("gdalbuildvrt", "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func Run(args []string) {
	// Create a new flag set
	flags := flag.NewFlagSet("fim", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Println(`Usage of fim:
  GDAL VSI paths can be used, given GDAL mush have access to cloud creds.
  CLI flag syntax. The following forms are permitted:
  -flag
  --flag   // double dashes are also permitted
  -flag=x
  -flag x  // non-boolean flags only
  Arguments:`)
		flags.PrintDefaults()
	}

	var controlsFile, fimLibDir, outputVRT string
	var relative bool

	// Define flags using flags.StringVar
	flags.StringVar(&fimLibDir, "lib", ".", "Directory containing FIM Library. GDAL VSI paths can be used, given GDAL mush have access to cloud creds")
	flags.BoolVar(&relative, "rel", true, "If relative paths should be used in VRT")
	flags.StringVar(&controlsFile, "c", "", "Path to the conrols CSV file")
	flags.StringVar(&outputVRT, "o", "fim.vrt", "Output VRT file path")

	// Parse flags from the arguments
	if err := flags.Parse(args); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	// Validate required flags
	if controlsFile == "" || fimLibDir == "" || outputVRT == "" {
		fmt.Println("Missing required flags")
		flags.PrintDefaults()
		os.Exit(1)
	}

	// Check if gdalbuildvrt is available
	if !CheckGDALBuildVRTAvailable() {
		log.Fatalf("Error: gdalbuildvrt is not available. Please install GDAL and ensure gdalbuildvrt is in your PATH.")
	}

	// Get the absolute paths
	absOutputPath, err := filepath.Abs(outputVRT)
	if err != nil {
		log.Fatalf("Error getting absolute path for output VRT file: %v", err)
	}

	if strings.HasPrefix(fimLibDir, "/vsi") && !strings.HasPrefix(absOutputPath, "/vsi") {
		relative = false
		// This avoids
		// go run main.go fim -c outputs.csv -lib /vsis3/fimc-data/data -o output.vrt
		// [/app/output.vrt ../vsis3/fimc-data/data/8489318/z0_0/f_1560.tif]
	}

	// Processing CSV and creating VRT
	file, err := os.Open(controlsFile)
	if err != nil {
		log.Fatalf("Error opening CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error reading CSV file: %v", err)
	}

	var tifFiles []string
	for _, record := range records[1:] { // Skip header row
		reachID := record[0]
		flow, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Fatal(err)
		}
		controlStage, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Fatal(err)
		}

		folderName := fmt.Sprintf("z%.1f", controlStage)
		folderName = filepath.Join(fimLibDir, reachID, folderName)
		folderName = strings.Replace(folderName, ".", "_", -1) // Replace '.' with '_'
		fileName := fmt.Sprintf("f_%d.tif", int(flow))
		filePath := filepath.Join(folderName, fileName)

		if relative {
			filePath, err = filepath.Rel(filepath.Dir(absOutputPath), filePath)
			if err != nil {
				log.Fatal(err)
			}
		}

		if strings.HasPrefix(filePath, `\vsis3`) {
			filePath = strings.ReplaceAll(filePath, `\`, "/")
		}
		tifFiles = append(tifFiles, filePath)

	}

	args = append([]string{absOutputPath}, tifFiles...)
	// fmt.Println(args)
	cmd := exec.Command("gdalbuildvrt", args...)
	// Redirecting the output to the standard output of the Go program
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error running gdalbuildvrt: %v", err)
	}

	fmt.Printf("VRT created at %s\n", absOutputPath)
}
