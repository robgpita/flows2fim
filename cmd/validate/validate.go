package validate

import (
	"database/sql"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	_ "modernc.org/sqlite"
)

// fimRow represents a single record discovered in the FIM library
type fimRow struct {
	reachID           int
	usFlow            int
	dsWse             float64
	boundaryCondition string
}

// processReachDir enumerates subfolders z_XXX and f_*.tif files for a single top-level reach folder
func processReachDir(reachDir, absFimLibDir string, fimChan chan<- fimRow) {
	filepath.WalkDir(reachDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := d.Name()
		if strings.HasPrefix(name, "f_") && strings.HasSuffix(name, ".tif") {
			usFlowStr := strings.TrimSuffix(strings.TrimPrefix(name, "f_"), ".tif")
			usFlow, convErr := strconv.Atoi(usFlowStr)
			if convErr != nil {
				log.Printf("Warning: could not parse us_flow: %s", name)
				return nil
			}

			relPath, _ := filepath.Rel(absFimLibDir, path)
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) < 3 {
				return nil
			}
			reachIDStr := parts[0]
			dirName := parts[1] // z_XXX
			reachID, convErr := strconv.Atoi(reachIDStr)
			if convErr != nil {
				log.Printf("Warning: could not parse reach_id: %s", reachIDStr)
				return nil
			}

			if !strings.HasPrefix(dirName, "z_") {
				return nil
			}
			dirSuffix := strings.TrimPrefix(dirName, "z_")

			var bc string
			var dsWse float64
			if dirSuffix == "nd" {
				bc = "nd"
				dsWse = 0.0
			} else {
				bc = "kwse"
				dsWseStr := strings.ReplaceAll(dirSuffix, "_", ".")
				dsWseFloat, parseErr := strconv.ParseFloat(dsWseStr, 64)
				if parseErr != nil {
					return nil
				}
				dsWse = dsWseFloat
			}

			fimChan <- fimRow{
				reachID:           reachID,
				usFlow:            usFlow,
				dsWse:             dsWse,
				boundaryCondition: bc,
			}
		}
		return nil
	})
}

// batchInsertFIMs insert FIM rows in batches
func batchInsertFIMs(db *sql.DB, fimChan <-chan fimRow) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	const batchSize = 1000
	batch := make([]fimRow, 0, batchSize)

	commitBatch := func() error {
		if len(batch) == 0 {
			return nil
		}

		// Can't do prepared statement as the final batch would not be of same size
		// Build single statement for multi-VALUES insert
		// INSERT INTO memdb.fim_entries(reach_id, us_flow, ds_wse, boundary_condition) VALUES (?, ?, ?, ?), (?, ?, ?, ?) ...
		sqlStr := "INSERT INTO memdb.fim_entries(reach_id, us_flow, ds_wse, boundary_condition) VALUES "
		vals := make([]interface{}, 0, len(batch)*4)
		placeholders := make([]string, 0, len(batch))

		for _, r := range batch {
			placeholders = append(placeholders, "(?,?,?,?)")
			vals = append(vals, r.reachID, r.usFlow, r.dsWse, r.boundaryCondition)
		}
		sqlStr += strings.Join(placeholders, ",")

		if _, err := tx.Exec(sqlStr, vals...); err != nil {
			return err
		}
		batch = batch[:0]
		return nil
	}

	for row := range fimChan {
		batch = append(batch, row)
		if len(batch) >= batchSize {
			if err := commitBatch(); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
	}

	// Do final flush when fimChan is closed in main routine
	if err := commitBatch(); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// writeMissingFims queries rating_curves table minus fim_entries table
// and write csv file
func writeMissingFims(db *sql.DB, outFims string) error {
	query := `
	SELECT
		rc.reach_id,
		rc.us_flow,
		rc.ds_wse,
		rc.boundary_condition
	FROM
		rating_curves rc
	LEFT JOIN
		memdb.fim_entries f
		ON (rc.reach_id = f.reach_id
			AND rc.us_flow = f.us_flow
			AND (CASE
					WHEN rc.boundary_condition = 'nd' THEN 0
					ELSE rc.ds_wse
				END) = f.ds_wse
			AND rc.boundary_condition = f.boundary_condition)
	WHERE
		f.reach_id IS NULL
	ORDER BY
		rc.reach_id, rc.boundary_condition, rc.ds_wse, rc.us_flow;
	`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("error querying missing_fims: %v", err)
	}
	defer rows.Close()

	file, err := os.Create(outFims)
	if err != nil {
		return fmt.Errorf("error creating missing_fims file: %v", err)
	}
	defer file.Close()
	w := csv.NewWriter(file)
	defer w.Flush()

	if err := w.Write([]string{"reach_id", "us_flow", "ds_wse", "boundary_condition"}); err != nil {
		return err
	}

	for rows.Next() {
		var reachID, usFlow int
		var dsWse float64
		var bc string
		if err := rows.Scan(&reachID, &usFlow, &dsWse, &bc); err != nil {
			return err
		}
		record := []string{
			strconv.Itoa(reachID),
			strconv.Itoa(usFlow),
			fmt.Sprintf("%.1f", dsWse),
			bc,
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

// writeMissingRatingCurves queries fim_entries table minus rating_curves table
// and write csv file
func writeMissingRatingCurves(db *sql.DB, outRcs string) error {
	query := `
	SELECT
		f.reach_id,
		f.us_flow,
		f.ds_wse,
		f.boundary_condition
	FROM
		memdb.fim_entries f
	LEFT JOIN
		rating_curves rc
		ON (f.reach_id = rc.reach_id
			AND f.us_flow = rc.us_flow
			AND f.ds_wse = (CASE
					WHEN f.boundary_condition = 'nd' THEN 0
					ELSE rc.ds_wse
				END)
			AND f.boundary_condition = rc.boundary_condition)
	WHERE
		rc.reach_id IS NULL
	ORDER BY
		f.reach_id, f.boundary_condition, f.ds_wse, f.us_flow;
	`

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("error querying missing_rating_curves: %v", err)
	}
	defer rows.Close()

	file, err := os.Create(outRcs)
	if err != nil {
		return fmt.Errorf("error creating missing_rating_curves file: %v", err)
	}
	defer file.Close()
	w := csv.NewWriter(file)
	defer w.Flush()

	if err := w.Write([]string{"reach_id", "us_flow", "ds_wse", "boundary_condition"}); err != nil {
		return err
	}

	for rows.Next() {
		var reachID, usFlow int
		var dsWse float64
		var bc string
		if err := rows.Scan(&reachID, &usFlow, &dsWse, &bc); err != nil {
			return err
		}
		record := []string{
			strconv.Itoa(reachID),
			strconv.Itoa(usFlow),
			fmt.Sprintf("%.1f", dsWse),
			bc,
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

func Run(args []string) error {
	flags := flag.NewFlagSet("validate", flag.ExitOnError)

	var (
		dbPath     string
		fimLibDir  string
		outFims    string
		outRcs     string
		concurrent int
	)

	flags.StringVar(&dbPath, "db", "", "Path to the rating curves database file")
	flags.StringVar(&fimLibDir, "lib", "", "Path to the FIM library directory")
	flags.StringVar(&outFims, "o_fims", "missing_fims.csv", "Output CSV for rating curve entries missing corresponding FIM files")
	flags.StringVar(&outRcs, "o_rcs", "missing_rating_curves.csv", "Output CSV for FIM entries missing corresponding rating curve records")
	flags.IntVar(&concurrent, "cc", 25, "Concurrent Count, number of top-level reach directories to process concurrently (default 25)")

	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("error parsing flags: %v", err)
	}

	// Validate required flags
	if dbPath == "" || fimLibDir == "" {
		fmt.Println("Missing required flags")
		flags.PrintDefaults()
		return fmt.Errorf("missing required flags")
	}

	// 1) Open the input DB ( we won't modify it).
	// to do: check if db exists
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("error opening DB: %v", err)
	}
	defer db.Close()

	// 2) Attach an in-memory DB for fim_entries
	_, err = db.Exec(`ATTACH ':memory:' AS memdb;`)
	if err != nil {
		return fmt.Errorf("error attaching in-memory db: %v", err)
	}

	// Create table memdb.fim_entries
	// to do: move query to a constant at the top of the file
	_, err = db.Exec(`CREATE TABLE memdb.fim_entries (
		reach_id INTEGER,
		us_flow INTEGER,
		ds_wse REAL,
		boundary_condition TEXT
	);`)
	if err != nil {
		return fmt.Errorf("error creating memdb.fim_entries: %v", err)
	}

	absFimLibDir, err := filepath.Abs(fimLibDir)
	if err != nil {
		return fmt.Errorf("error getting absolute path for fim library: %v", err)
	}

	// 3) Setup concurrency
	fimChan := make(chan fimRow, 2000) // buffer for discovered rows
	doneChan := make(chan struct{})

	// Single writer goroutine that batch-inserts rows into memdb.fim_entries
	go func() {
		err := batchInsertFIMs(db, fimChan)
		if err != nil {
			log.Printf("Error inserting FIM rows: %v", err)
		}
		close(doneChan)
	}()

	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrent) // limit concurrency to 'cc'
	// sync/semaphore could also have been used here

	// 4) Find top-level directories (reach folders)
	dirEntries, err := os.ReadDir(absFimLibDir)
	if err != nil {
		return fmt.Errorf("error reading fim library directory: %v", err)
	}
	var reachDir string
	for _, de := range dirEntries {
		if de.IsDir() {
			reachDir = filepath.Join(absFimLibDir, de.Name())
			wg.Add(1)
			sem <- struct{}{} // Acquire concurrency token
			go func(reachDir string) {
				defer wg.Done()
				defer func() { <-sem }() // Release token
				processReachDir(reachDir, absFimLibDir, fimChan)
			}(reachDir)
		}
	}

	// Wait for all reach processing goroutines to finish
	wg.Wait()
	close(fimChan) // no more FIM rows
	<-doneChan     // wait for the DB writer goroutine

	// ) Generate two output CSV files:
	//    - missing_fims: rating_curves minus fim_entries
	//    - missing_rating_curves: fim_entries minus rating_curves
	if err := writeMissingFims(db, outFims); err != nil {
		return err
	}
	if err := writeMissingRatingCurves(db, outRcs); err != nil {
		return err
	}

	fmt.Println("Validation complete")
	fmt.Printf("Missing FIMs file created at: %s\nMissing Rating Curves file created at: %s\n", outFims, outRcs)
	return nil
}
