# csv-update

This is a companion project for [maketest](https://github.com/phpeterson-usf/maketest)
which generates the source CSV file described below. The purpose is to take maketest
output and allow it to be imported into the Canvas Gradebook

### The Problem
1. I have a CSV file where one column has unique IDs (GitHub User IDs)
1. I have a second CSV file which has different unique IDs (Canvas SIS Login IDs)
1. I want to update a column of values from the first CSV file to the second, without
relying on order of rows

### My Solution
1. Make up a mapping file which maps GitHub IDs to SIS Login IDs, we can merge the
data across CSV files, like a SQL join with foreign keys to different tables
1. The result of that merge is a CSV file which can be imported into Canvas,
reflecting the scoring results for Maketest without retyping them

### Usage
1. Set up your Assignment Groups in Canvas, with an assignment within the group
for test automation. This model of using Canvas ensures that there will be a column
for automated grading in the exported CSV file. Canvas rubrics do not get a column
in the exported file.
1. Export a CSV file from your Canvas Gradebook
1. Run maketest csv and copy that file into where you use `csv-update`
(I run `maketest` on a Raspberry Pi, but do Canvas import/export on a desktop computer)
1. Create a CSV file which maps GitHub ID to SIS Login ID
1. `csv-update` takes a `-C` option to set a working directory. I put my CSV files on my
desktop, so `/Users/phil/Desktop`
1. Run the merge
	<pre><code>csv-update -C ~/Desktop</code></pre>
	`csv-update` asks you to choose the source file (the one with scores), the destination
	file (the one you exported from Canvas), and the mapping file. You also choose the column
	in the destination file where the scores should go
1. Output goes to a new file called `<destination>-updated.csv` that you can import back to Canvas
