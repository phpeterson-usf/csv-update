/*
	maketest-merge generates a Canvas-ready CSV file with grade values inserted
	for the scores calculated by maketest: https://github.com/phpeterson-usf/maketest

	The Problem
	1. Maketest knows students' Github IDs and the scores generated by maketest
	2. Canvas knows SIS Login IDs, which it uses to uniquely identify students
	3. In order to get grades from the automated test tool, we need a mapping table

	My Solution
	1. If we make up (by hand) a CSV file which maps GitHub IDs to SIS Login IDs, we can 
	   merge the data across CSV files, like a SQL join with foreign keys to different tables
	2. The result of that merge is a CSV file which can be imported into Canvas, 
	   reflecting the scoring results for Maketest without retyping them
	   
	Usage
	1. Set up your Assignment Groups in Canvas, with an assignment within the group
	   for test automation. This model of using Canvas ensures that there will be a column
	   for automated grading in the exported CSV file. Canvas rubrics do not get a column
	   in the exported file.
	2. Export a CSV file from your Canvas Gradebook
	3. Run maketest csv and copy that file into where you use maketest-merge
	   (I run maketest on a Raspberry Pi, but do Canvas import/export on a desktop computer)
	4. Create a CSV file which maps GitHub ID to SIS Login ID
	5. Run the merge: maketest-merge -canvas foo.csv -maketest project02.csv -map map.csv
	6. The tool generates foo-merged.csv which is ready to import into Canvas
*/

package main

import (
	"encoding/csv"
	"flag"
	"log"
	"os"
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
func writeRecords(recs [][]string, fname string) {
	f, err := os.Create(fname)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	err = writer.WriteAll(recs)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
}


// findColumnIndex() finds the column with the given name
func findColumnIndex(col_hdrs[]string, col_name string) (int) {
	for h := range(col_hdrs) {
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

/*
const (
	kSrc int = iota
	kDst
	kMap
	kMax
)

type struct {
	fname, prompt string
} CsvSpec


func chooseFiles(exclude string[3]) (string)
*/
func main() {
/*
	files [kMax]CsvSpec
	files[kSrc] = {
		fname: "",  
		prompt:"Choose a SOURCE CSV file", colPrompt: "Choose the value column"
	}
	for i := kSrc; i < kMax; i++ {
		csv_files[i] = chooseCsvFile(csvFiles)
	}
	
	var choice int
	fmt.Printf("1. foo\n2. bar\n")
	fmt.Scanf("%d", &choice)
	fmt.Printf("Choice %d\n", choice)
*/	
	dst_fname := flag.String("dst", "", "CSV file exported from Canvas")
	src_fname := flag.String("src", "", "CSV file containing scores")
	map_fname := flag.String("map", "", "CSV file containing mappings from GitHub profile to SIS ID")
	
	flag.Parse()
	
	if *dst_fname == "" || *src_fname == "" || *map_fname == "" {
		flag.Usage()
		return
	}

	// Build the tables of row/column data from each file
	var (
		dest_recs, src_recs, map_recs [][]string
		err error
	)
	if dest_recs, err = readRecords(*dst_fname); err != nil {
		log.Fatalln(err)
	}
	if src_recs, err = readRecords(*src_fname); err != nil {
		log.Fatalln(err)
	}
	if map_recs, err = readRecords(*map_fname); err != nil {
		log.Fatalln(err)
	}


	// Naming conventions: 
	// cix: column index
	// rix: row index
	// val: cell value at [rix][cix] 
	// e.g. score_github_cix is the column index of the Github column in the score table
	
	src_key_cix := findColumnIndex(src_recs[0], "GitHub ID")
	src_val_cix := findColumnIndex(src_recs[0], "Score")
	map_dest_cix := findColumnIndex(map_recs[0], "SIS Login ID")
	map_src_cix := findColumnIndex(map_recs[0], "GitHub ID")
	dest_key_cix := findColumnIndex(dest_recs[0], "SIS Login ID")
	dest_val_cix := findColumnIndex(dest_recs[0], "Project01-Automated")

	// Walk the rows in the score table. Start at 1 to skip the header row
	for src_rix := 1; src_rix < len(src_recs); src_rix++ {
		// Find the value of the GitHub column in this row
		src_key_val := src_recs[src_rix][src_key_cix]

		// Find the row in the mapping table in which the github column contains this github value
		map_rix := findRowIndex(map_recs, map_src_cix, src_key_val)

		// Pull the value of the SIS Login ID column in that row
		dest_key_val := getValue(map_recs, map_rix, map_dest_cix)

		if dest_key_val != "" {
			// Find the row in the input table in which the SIS Login ID column matches this SIS Login ID value
			dest_key_rix := findRowIndex(dest_recs, dest_key_cix, dest_key_val)

			// Get the score from the score table
			src_val := getValue(src_recs, src_rix, src_val_cix)

			// Set the score into the project column of the SIS user's row
			setValue(dest_recs, dest_key_rix, dest_val_cix, src_val)
		}
	}

	writeRecords(dest_recs, "out.csv")
}