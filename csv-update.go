package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)


// readRecords reads all of the CSV file's data into a 2D array of row/column data
func readRecords(fname string) ([][]string, error) {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	
	rdr := csv.NewReader(f)
	return rdr.ReadAll()
}


// writeRecords() writes all of the given records to a new/truncated CSV file
func writeRecords(recs [][]string, fname string, num_updated int) {
	f, err := os.Create(fname)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	err = writer.WriteAll(recs)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Updated %d values in %s\n", num_updated, fname)
}


// findColumnIndex() finds the column with the given name
func findColumnIndex(col_hdrs[]string, col_name string) (int) {
	for h := range(col_hdrs) {
		// Use Contains because Canvas appends a UID to the assignment name
		// e.g. "Project02-Automated (631594)" and we need it to match "Project02-Automated"
		if strings.Contains(col_hdrs[h], col_name) {
			return h
		}
	}
	log.Fatalf("Can't find column named: %s\n", col_name)
	return -1
}


// findRowIndex finds the row where col_idx contains the given value
func findRowIndex(recs [][]string, col_idx int, col_val string) (int) {
	for r := range(recs) {
		if recs[r][col_idx] == col_val {
			return r
		}
	}
	log.Fatalf("Can't find matching row for %s\n", col_val)
	return -1
}


// getValue() sets the value of a row and column idx to the given value
func getValue(recs [][]string, row_idx, col_idx int) (string) {
	return recs[row_idx][col_idx]
}


// setValue() takes the index of the row and column and sets the value
func setValue(recs [][]string, row_idx, col_idx int, val string) {
	recs[row_idx][col_idx] = val
}

// menuChoice() presents a list of options so the user doesn't have to type
// them as command line arguments
func menuChoice(choices []string, prompt string) (int) {
	fmt.Println(prompt)
	for c := range(choices) {
		fmt.Printf("[%d] %s\n", c, choices[c])
	}
	choice := -1
	for choice < 0 {
		fmt.Printf("Number? ")
		fmt.Scanf("%d", &choice)
	}
	return choice
}


func main() {
	var dir string
	flag.StringVar(&dir, "C", ".", "Working directory")
	flag.Parse()
	
	if dir[len(dir)-1] != '/' {
		dir += "/"
	}

	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if ext == ".csv" {
			files = append(files, info.Name())
		}
		return err
	})

	dst_fname := dir + files[menuChoice(files, "Which is the dest file from Canvas?")]
	src_fname := dir + files[menuChoice(files, "Which is the source file with scores?")]
	map_fname := dir + files[menuChoice(files, "Which is the mapping file?")]

	// Build the tables of row/column data from each file
	var dest_recs, src_recs, map_recs [][]string
	if dest_recs, err = readRecords(dst_fname); err != nil {
		log.Fatalln(err)
	}
	if src_recs, err = readRecords(src_fname); err != nil {
		log.Fatalln(err)
	}
	if map_recs, err = readRecords(map_fname); err != nil {
		log.Fatalln(err)
	}

	// Naming conventions: 
	// cix: column index
	// rix: row index
	// val: cell value at [rix][cix] 
	// e.g. score_github_cix is the column index of the Github column in the score table

	// I'm going to hard-code these for now. Maybe there will be a reason to
	// make the user choose them?
	src_key_cix := findColumnIndex(src_recs[0], "GitHub ID")
	src_val_cix := findColumnIndex(src_recs[0], "Score")
	map_dest_cix := findColumnIndex(map_recs[0], "SIS Login ID")
	map_src_cix := findColumnIndex(map_recs[0], "GitHub ID")
	dest_key_cix := findColumnIndex(dest_recs[0], "SIS Login ID")

	// Ask the user to specify the column ID to put scores into
	dest_val_cix := menuChoice(dest_recs[0], "Which column should we put scores in?")

	num_updated := 0
	// Walk the rows in the score table. Start at 1 to skip the header row
	for src_rix := 1; src_rix < len(src_recs); src_rix++ {
		// Find the value of the source key (GitHub ID) column in this row
		src_key_val := getValue(src_recs, src_rix, src_key_cix)

		// Find the row in the mapping table in which the source key (GitHub ID) column 
		// contains this value
		map_rix := findRowIndex(map_recs, map_src_cix, src_key_val)

		// Pull the value of dest key (SIS Login ID) column for that row index
		dest_key_val := getValue(map_recs, map_rix, map_dest_cix)

		if dest_key_val != "" {
			// Find the row in the dest table in which the dest key (SIS Login ID) column 
			// matches this dest key (SIS Login ID) value
			dest_key_rix := findRowIndex(dest_recs, dest_key_cix, dest_key_val)

			// Get the score from the source (score) table
			src_val := getValue(src_recs, src_rix, src_val_cix)

			// Set the (score) value into the project column of the SIS user's row
			setValue(dest_recs, dest_key_rix, dest_val_cix, src_val)
			num_updated++
		}
	}

	// Make up an output filename in the same directory as the original file
	without_ext := strings.TrimRight(dst_fname, filepath.Ext(dst_fname))
	writeRecords(dest_recs, without_ext + "-updated.csv", num_updated)
}
