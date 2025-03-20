package fim

import (
	"encoding/csv"
	"encoding/xml"
	"flag"
	"flows2fim/pkg/utils"
	"fmt"
	"io"
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

func createTempVRT(inputFileListPath, absOutputPath string) (string, error) {

	// Create intermediate directories if they do not exist
	if err := os.MkdirAll(filepath.Dir(absOutputPath), 0755); err != nil {
		return "", fmt.Errorf("could not create directories for %s: %v", absOutputPath, err)
	}

	// We don't really need os.CreateTemp, but we are using it to generate random file name
	tempVRTFile, err := os.CreateTemp(filepath.Dir(absOutputPath), "~f2f_*.tmp")
	if err != nil {
		return "", fmt.Errorf("error creating temporary VRT file: %v", err)
	}
	tempVRTPath := tempVRTFile.Name()
	tempVRTFile.Close() // Close now, gdalbuildvrt will write to it

	// Build temporary VRT file
	vrtArgs := []string{"-input_file_list", inputFileListPath, tempVRTPath}
	vrtCmd := exec.Command("gdalbuildvrt", vrtArgs...)
	vrtCmd.Stdout = os.Stdout
	vrtCmd.Stderr = os.Stderr

	slog.Debug("Creating temporary VRT file",
		"command", fmt.Sprintf("gdalbuildvrt %s", strings.Join(vrtArgs, " ")),
		"tempVRT", tempVRTPath,
	)

	if err := vrtCmd.Run(); err != nil {
		return "", fmt.Errorf("error running gdalbuildvrt: %v", err)
	}

	return tempVRTPath, nil
}

func addVRTPixelFunc(vrtPath string) (string, error) {
	// Open the original VRT file for reading
	inFile, err := os.Open(vrtPath)
	if err != nil {
		return "", fmt.Errorf("error opening VRT file: %v", err)
	}
	defer inFile.Close()

	// Create a temporary file to write the modified XML
	modVRTFile, err := os.CreateTemp(filepath.Dir(vrtPath), "~f2f_*.tmp")
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %v", err)
	}

	modVRTPath := modVRTFile.Name()
	defer modVRTFile.Close()

	encoder := xml.NewEncoder(modVRTFile)
	decoder := xml.NewDecoder(inFile)

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", fmt.Errorf("error decoding token: %v", err)
		}

		switch se := tok.(type) {
		case xml.StartElement:
			if se.Name.Local == "VRTRasterBand" {

				se.Attr = append(se.Attr, xml.Attr{
					Name:  xml.Name{Local: "subClass"},
					Value: "VRTDerivedRasterBand",
				})

				// Encode the modified start element
				if err := encoder.EncodeToken(se); err != nil {
					return "", fmt.Errorf("error encoding VRTRasterBand: %v", err)
				}

				// Create and encode PixelFunctionType element
				pixelFunc := xml.StartElement{Name: xml.Name{Local: "PixelFunctionType"}}
				if err := encoder.EncodeToken(pixelFunc); err != nil {
					return "", fmt.Errorf("error encoding PixelFunctionType start: %v", err)
				}
				if err := encoder.EncodeToken(xml.CharData("max")); err != nil {
					return "", fmt.Errorf("error encoding PixelFunctionType value: %v", err)
				}
				if err := encoder.EncodeToken(pixelFunc.End()); err != nil {
					return "", fmt.Errorf("error encoding PixelFunctionType end: %v", err)
				}
			} else {
				// Encode other elements as-is
				if err := encoder.EncodeToken(se); err != nil {
					return "", fmt.Errorf("error encoding token: %v", err)
				}
			}
		default:
			// Encode all other tokens as-is
			if err := encoder.EncodeToken(tok); err != nil {
				return "", fmt.Errorf("error encoding token: %v", err)
			}
		}
	}

	// Flush encoder to ensure all data is written
	if err := encoder.Flush(); err != nil {
		return "", fmt.Errorf("error flushing encoder: %v", err)
	}

	return modVRTPath, nil
}

func Run(args []string) (gdalArgs []string, err error) {
	flags := flag.NewFlagSet("fim", flag.ExitOnError)
	flags.Usage = func() {
		fmt.Println(usage)
		flags.PrintDefaults()
	}

	var controlsFile, fimLibDir, libType, outputFormat, outputFile string

	// Define flags using flags.StringVar
	flags.StringVar(&fimLibDir, "lib", "", "Directory containing FIM Library. GDAL VSI paths can be used, given GDAL must have access to cloud creds")
	flags.StringVar(&controlsFile, "c", "", "Path to the controls CSV file")
	flags.StringVar(&outputFormat, "fmt", "vrt", "Output format: 'vrt', 'cog' or 'tif'")
	flags.StringVar(&libType, "type", "", "Library type: 'depth' or 'extent'")
	flags.StringVar(&outputFile, "o", "", "Output FIM file path")

	// Parse flags from the arguments
	if err := flags.Parse(args); err != nil {
		return []string{}, fmt.Errorf("error parsing flags: %v", err)
	}

	outputFormat = strings.ToLower(outputFormat) // COG, cog, VRT, vrt all okay

	// Validate required flags
	if controlsFile == "" || fimLibDir == "" || libType == "" || outputFile == "" {
		fmt.Println(controlsFile, fimLibDir, outputFile)
		fmt.Println("Missing required flags")
		flags.PrintDefaults()
		return []string{}, fmt.Errorf("missing required flags")
	}

	if libType != "depth" && libType != "extent" {
		return []string{}, fmt.Errorf("library type must be either 'depth' or 'extent'")
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
	inputFileListPath, err := writeFileList(tifFiles)
	if err != nil {
		return []string{}, fmt.Errorf("error writing file list to temporary file: %v", err)
	}
	defer os.Remove(inputFileListPath)

	tempVRTPath, err := createTempVRT(inputFileListPath, absOutputPath)
	if err != nil {
		return []string{}, fmt.Errorf("error creating temp vrt: %v", err)
	}
	defer os.Remove(tempVRTPath)

	var modVRTPath string
	// gdal pixel functions require minimum of 2 bands
	// gdal_translate is painfully slow when pixel function is there,
	// hence we only add pixel func when it is necessary which is only for extent library
	// pixel function max is available starting gdal 3.8.0
	if len(tifFiles) > 1 && libType == "extent" {
		modVRTPath, err = addVRTPixelFunc(tempVRTPath)
		if err != nil {
			return []string{}, fmt.Errorf("error modifying temp vrt: %v", err)
		}
		defer os.Remove(modVRTPath)
	} else {
		modVRTPath = tempVRTPath
	}

	if outputFormat == "vrt" {
		// For VRT, simply move the temporary file to the final destination for atomicity
		slog.Debug("Moving temporary VRT to final destination",
			"from", modVRTPath,
			"to", absOutputPath)

		if err := os.Rename(modVRTPath, absOutputPath); err != nil {
			return []string{}, fmt.Errorf("error renaming temp file %s to %s: %v", tempVRTPath, absOutputPath, err)
		}

	} else {
		// For TIF or COG, use gdal_translate to convert the VRT
		translateArgs := []string{
			"-co", "COMPRESS=DEFLATE",
			"-co", "NUM_THREADS=ALL_CPUS",
			"-if", "VRT",
			"-of", "GTiff",
			modVRTPath,
			"",
		}

		// Directly creating COG with Pixel Function is hours magnitude slower than first creating GTiff and then converting to COG
		switch outputFormat {
		case "tif":
			translateArgs[7] = absOutputPath
		case "cog":
			tempTIFFile, err := os.CreateTemp(filepath.Dir(absOutputPath), "~f2f_*.tmp")
			if err != nil {
				return []string{}, fmt.Errorf("error creating temporary TIF file: %v", err)
			}
			tempTIFFile.Close()
			tempTIFPath := tempTIFFile.Name()
			translateArgs[7] = tempTIFPath
			defer os.Remove(tempTIFPath)
		}

		translateCmd := exec.Command("gdal_translate", translateArgs...)
		translateCmd.Stdout = os.Stdout
		translateCmd.Stderr = os.Stderr

		slog.Debug("Converting VRT to TIFF",
			"command", fmt.Sprintf("gdal_translate %s", strings.Join(translateArgs, " ")),
			"format", outputFormat,
		)

		if err := translateCmd.Run(); err != nil {
			return []string{}, fmt.Errorf("error converting VRT to GTIFF: %v", err)
		}

		if outputFormat == "cog" {
			// Convert to COG
			cogArgs := []string{
				"-co", "COMPRESS=DEFLATE",
				"-co", "NUM_THREADS=ALL_CPUS",
				"-if", "GTiff",
				"-of", "COG",
				translateArgs[7], // Use the temporary TIF file
				absOutputPath,
			}

			cogCmd := exec.Command("gdal_translate", cogArgs...)
			cogCmd.Stdout = os.Stdout
			cogCmd.Stderr = os.Stderr

			slog.Debug("Converting TIF to COG",
				"command", fmt.Sprintf("gdal_translate %s", strings.Join(cogArgs, " ")),
			)

			if err := cogCmd.Run(); err != nil {
				return []string{}, fmt.Errorf("error converting TIF to COG: %v", err)
			}
		}
	}

	fmt.Printf("Composite FIM created at %s\n", absOutputPath)

	return gdalArgs, nil
}
