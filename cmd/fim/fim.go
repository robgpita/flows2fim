package fim

import (
	"encoding/csv"
	"flag"
	"flows2fim/pkg/utils"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var usage string = `Usage of fim:
Given a control table and a fim library folder, create a flood inundation VRT or a merged TIFF for the control conditions.
GDAL VSI paths can be used, given GDAL must have access to cloud creds.

FIM Library Specifications:
- All maps should have same CRS, Resolution, vertical units (if any), and nodata value
- Should have following folder structure:
.
├── 2821866
│   ├── z_nd
│   │   ├── f_10283.tif
│   │   ├── f_104569.tif
│   │   ├── f_11199.tif
│   │   ├── f_112807.tif
│   ├── z_53_5
│       ├── f_102921.tif
│       ├── f_10485.tif
│       ├── f_111159.tif
│       ├── f_11309.tif



Arguments:` // Usage should be always followed by PrintDefaults()

var gdalCommands = map[string]string{
	"vrt": "gdalbuildvrt",
	"tif": "gdalwarp",
	"cog": "gdalwarp",
	// not using gdal_merge since gdalwarp allow writing to COG while gdam_merge does not
	// also, gdalwarp is more powerful than gdal_merge
	// also it has consistent name across plateforms unlike gdal_merge vs gdal_merge.py
}

func writeFileList(fileList []string) (string, error) {
	tmpfile, err := os.CreateTemp("", "filelist-*.txt")
	if err != nil {
		return "", err
	}
	defer tmpfile.Close()

	for _, file := range fileList {
		if _, err := tmpfile.WriteString(file + "\n"); err != nil {
			return "", err
		}
	}

	slog.Debug("Created temporary file list", "path", tmpfile.Name(), "files_count", len(fileList))
	return tmpfile.Name(), nil
}

func Run(args []string) (gdalArgs []string, err error) {
	flags := flag.NewFlagSet("fim", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Println(usage)
		flags.PrintDefaults()
	}

	var controlsFile, fimLibDir, outputFile, outputFormat string

	// Define flags using flags.StringVar
	flags.StringVar(&fimLibDir, "lib", "", "Directory containing FIM Library. GDAL VSI paths can be used, given GDAL must have access to cloud creds")
	flags.StringVar(&controlsFile, "c", "", "Path to the controls CSV file")
	flags.StringVar(&outputFormat, "fmt", "vrt", "Output format: 'vrt', 'cog' or 'tif'")
	flags.StringVar(&outputFile, "o", "", "Output FIM file path")

	// Parse flags from the arguments
	if err := flags.Parse(args); err != nil {
		return []string{}, fmt.Errorf("error parsing flags: %v", err)
	}

	outputFormat = strings.ToLower(outputFormat) // COG, cog, VRT, vrt all okay

	// Validate required flags
	if controlsFile == "" || fimLibDir == "" || outputFile == "" {
		fmt.Println(controlsFile, fimLibDir, outputFile)
		fmt.Println("Missing required flags")
		flags.PrintDefaults()
		return []string{}, fmt.Errorf("missing required flags")
	}

	// Check if gdalbuildvrt or GDAL tool is available
	if !utils.CheckGDALToolAvailable(gdalCommands[outputFormat]) {
		slog.Error("GDAL tool missing", "tool", gdalCommands[outputFormat])
		return []string{}, fmt.Errorf("%[1]s is not available. Please install GDAL and ensure %[1]s is in your PATH", gdalCommands[outputFormat])
	}

	var absOutputPath, absFimLibPath string
	if strings.HasPrefix(outputFile, "/vsi") {
		absOutputPath = outputFile
	} else {
		absOutputPath, err = filepath.Abs(outputFile)
		if err != nil {
			return []string{}, fmt.Errorf("error getting absolute path for output file: %v", err)
		}
	}

	if strings.HasPrefix(fimLibDir, "/vsi") {
		absFimLibPath = fimLibDir
	} else {
		absFimLibPath, err = filepath.Abs(fimLibDir)
		if err != nil {
			return []string{}, fmt.Errorf("error getting absolute path for FIM library directory: %v", err)
		}
	}

	// Processing CSV and creating VRT or TIFF
	file, err := os.Open(controlsFile)
	if err != nil {
		return []string{}, fmt.Errorf("error opening controls file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return []string{}, fmt.Errorf("error reading CSV file: %v", err)
	}

	if len(records) < 2 {
		return []string{}, fmt.Errorf("no records in control file")
	}

	var tifFiles []string
	for _, record := range records[1:] { // Skip header row
		reachID := record[0]

		record[2] = strings.Replace(record[2], ".", "_", -1) // Replace '.' with '_'
		folderName := filepath.Join(absFimLibPath, reachID, fmt.Sprintf("z_%s", record[2]))
		fileName := fmt.Sprintf("f_%s.tif", record[1])
		absTifPath := filepath.Join(folderName, fileName)

		// join on windows may cause \vsi
		if strings.HasPrefix(absTifPath, `\vsi`) {
			absTifPath = strings.ReplaceAll(absTifPath, `\`, "/")
		}
		tifFiles = append(tifFiles, absTifPath)
	}

	// Write file paths to a temporary file
	tempFileName, err := writeFileList(tifFiles)
	if err != nil {
		return []string{}, fmt.Errorf("error writing file list to temporary file: %v", err)
	}
	defer os.Remove(tempFileName)

	switch outputFormat {
	case "vrt":
		gdalArgs = []string{"-input_file_list", tempFileName, absOutputPath}

	case "tif":
		gdalArgs = []string{
			"-co", "COMPRESS=DEFLATE", // we are not doing predictor because we are not sure what will be our input tifs format https://kokoalberti.com/articles/geotiff-compression-optimization-guide/?ref=feed.terramonitor.com
			"--optfile", tempFileName,
			"-overwrite", absOutputPath,
		}

	case "cog":
		// Simplified COG creation
		gdalArgs = []string{
			"-co", "COMPRESS=DEFLATE",
			"-of", "COG",
			"-overwrite",
			"--optfile", tempFileName,
			absOutputPath,
		}
	}

	cmd := exec.Command(gdalCommands[outputFormat], gdalArgs...)

	slog.Debug("Executing GDAL command",
		"command", fmt.Sprintf("%s %s", gdalCommands[outputFormat], strings.Join(gdalArgs, " ")),
		"workdir", cmd.Dir,
	)

	// Redirecting the output to the standard output of the Go program
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return []string{}, fmt.Errorf("error running %s: %v", gdalCommands[outputFormat], err)
	}

	fmt.Printf("Composite FIM created at %s\n", absOutputPath)

	return gdalArgs, nil
}
