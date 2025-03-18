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
Given a control table and a fim library folder, create a composite flood inundation map for the control conditions.
GDAL VSI paths can be used (only for library and not for output), given GDAL must have access to cloud creds.

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

	// Check if required GDAL tools are available
	requiredTools := []string{"gdalbuildvrt"}
	if outputFormat != "vrt" {
		requiredTools = append(requiredTools, "gdal_translate")
	}

	for _, tool := range requiredTools {
		if !utils.CheckGDALToolAvailable(tool) {
			slog.Error("GDAL tool missing", "tool", tool)
			return []string{}, fmt.Errorf("%[1]s is not available. Please install GDAL and ensure %[1]s is in your PATH", tool)
		}
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

	// Create intermediate directories if they do not exist
	if err := os.MkdirAll(filepath.Dir(absOutputPath), 0755); err != nil {
		return []string{}, fmt.Errorf("could not create directories for %s: %v", absOutputPath, err)
	}

	// We don't really need os.CreateTemp, but we are using it to generate random file name
	tempVRTFile, err := os.CreateTemp(filepath.Dir(absOutputPath), "~f2f_*.tmp")
	if err != nil {
		return []string{}, fmt.Errorf("error creating temporary VRT file: %v", err)
	}
	tempVRTPath := tempVRTFile.Name()
	tempVRTFile.Close()          // Close now, gdalbuildvrt will write to it
	defer os.Remove(tempVRTPath) // Always attempt to remove tempfile even if file is already removed

	// Build temporary VRT file
	vrtArgs := []string{"-input_file_list", tempFileName, tempVRTPath}
	vrtCmd := exec.Command("gdalbuildvrt", vrtArgs...)
	vrtCmd.Stdout = os.Stdout
	vrtCmd.Stderr = os.Stderr

	slog.Debug("Creating temporary VRT file",
		"command", fmt.Sprintf("gdalbuildvrt %s", strings.Join(vrtArgs, " ")),
		"tempVRT", tempVRTPath,
	)

	if err := vrtCmd.Run(); err != nil {
		return []string{}, fmt.Errorf("error running gdalbuildvrt: %v", err)
	}

	if outputFormat == "vrt" {
		// For VRT, simply move the temporary file to the final destination for atomicity
		slog.Debug("Moving temporary VRT to final destination",
			"from", tempVRTPath,
			"to", absOutputPath)

		if err := os.Rename(tempVRTPath, absOutputPath); err != nil {
			return []string{}, fmt.Errorf("error renaming temp file %s to %s: %v", tempVRTPath, absOutputPath, err)
		}

	} else {
		// For TIF or COG, use gdal_translate to convert the VRT
		var translateArgs []string

		switch outputFormat {
		case "tif":
			translateArgs = []string{
				"-co", "COMPRESS=DEFLATE",
				"-of", "GTiff",
				tempVRTPath,
				absOutputPath,
			}
		case "cog":
			translateArgs = []string{
				"-co", "COMPRESS=DEFLATE",
				"-of", "COG",
				tempVRTPath,
				absOutputPath,
			}
		}

		translateCmd := exec.Command("gdal_translate", translateArgs...)
		translateCmd.Stdout = os.Stdout
		translateCmd.Stderr = os.Stderr

		slog.Debug("Converting VRT to final format",
			"command", fmt.Sprintf("gdal_translate %s", strings.Join(translateArgs, " ")),
			"format", outputFormat,
		)

		if err := translateCmd.Run(); err != nil {
			return []string{}, fmt.Errorf("error converting VRT to %s: %v", outputFormat, err)
		}
	}

	fmt.Printf("Composite FIM created at %s\n", absOutputPath)

	return gdalArgs, nil
}
