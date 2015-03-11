Arguments:

 * Source file
 * Update index file (k,v)
 * Output file

Flags:

 * Number of worker threads
 * Verbose
 * Read / write progress
 
Functionality:

 * Read the entire input file line by line into a slice

 * Read the Update Index into a map

 * Create a pool of worker goroutines
 * Worker gets a copy of one line of the input file:
   * For each entry in the Update Index map (has a copy of the map?) ...
     * Replace all old with new usernames (strings/bytes.Replace())
   * Returns updated line to the boss
 * Once workers complete their job they're given the next line

* Write slice back to outputfile
