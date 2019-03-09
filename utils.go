package utils

import (
	"encoding/json"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus" // Pretty log library, not the fastest (zerolog/zap)
	"github.com/valyala/gozstd"
)

// Random initalizate a new seed using the UNIX Nano time and return an integer between the 2 input value
func Random(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

// IsDirectory Verify if the path provided is a directory
func IsDirectory(path *string) bool {
	fileInfo, err := os.Stat(*path)
	if err != nil {
		log.Error("IsDirectory | Some error occours! | Path ", path, " IS NOT A DIRECTORY :/")
		return false
	}
	return fileInfo.IsDir()
}

// FreeSystemMemory Call the garbage collector every "gcSleep" minutes
func FreeSystemMemory(gcSleep *int) {
	time.Sleep(1 * time.Minute) // Just relax and load the data
	var m runtime.MemStats
	var g debug.GCStats
	for i := 0; i > -1; {
		log.Info(printMemUsage(&m, &g))
		debug.FreeOSMemory()
		log.Info("--- Memory freed! ---")
		log.Info(printMemUsage(&m, &g))
		time.Sleep(time.Duration(*gcSleep) * time.Minute)
	}
}

//Monitor is the data structure for store the metrics data
type Monitor struct {
	Alloc,
	TotalAlloc,
	Sys,
	Mallocs,
	Frees,
	LiveObjects,
	PauseTotalNs,
	MCacheInuse uint64
	GCCPUFraction float64
	NumGC         uint32
	NumGoroutine  int
}

// ExportMetrics  is in charge to retrive the information related to the resources used
func ExportMetrics() []byte {
	var m Monitor
	var rtm runtime.MemStats

	// Read full mem stats
	runtime.ReadMemStats(&rtm)

	// Number of goroutines
	m.NumGoroutine = runtime.NumGoroutine()

	// Misc memory stats
	m.Alloc = rtm.Alloc
	m.TotalAlloc = rtm.TotalAlloc
	m.Sys = rtm.Sys
	m.Mallocs = rtm.Mallocs
	m.Frees = rtm.Frees

	// Live objects = Mallocs - Frees
	m.LiveObjects = m.Mallocs - m.Frees

	// GC Stats
	m.PauseTotalNs = rtm.PauseTotalNs
	m.MCacheInuse = rtm.MCacheInuse
	m.NumGC = rtm.NumGC
	m.GCCPUFraction = rtm.GCCPUFraction

	// Just encode to json and print
	b, _ := json.Marshal(m)
	return b

}

/* func ExportMetrics() {
	memstatsFunc := expvar.Get("memstats").(expvar.Func)
	memstats := memstatsFunc().(runtime.MemStats)
	log.Debug(memstats)
}
*/
// printMemUsage outputs the current, total and OS memory being used.
// As well as the number of garage collection cycles completed.
func printMemUsage(m *runtime.MemStats, g *debug.GCStats) string {
	runtime.ReadMemStats(m)
	debug.ReadGCStats(g)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	return "Alloc = " + bToMb(m.Alloc) + "MiB\tTotalAlloc = " + bToMb(m.TotalAlloc) +
		"MiB\tSys = " + bToMb(m.Sys) + "MiB\tNumGC = " + strconv.FormatInt(int64(m.NumGC), 10) + "\tPausGCTotal = " + g.PauseTotal.String()
}

// lazy
func bToMb(b uint64) string {
	return strconv.FormatUint(b/2024, 10)
}

// ReadFile is in charge to read the last number of lines for a given file and return the content in a compressed format
func ReadFile(filename string, lines int) []byte {
	app := "tail -n " + strconv.Itoa(lines) + " \"" + filename + "\""
	cmd := exec.Command("/bin/sh", "-c", app)
	stdout, err := cmd.Output()
	if len(stdout) == 0 || err != nil { // Empty file or error
		log.Error("ReadFile | ERROR, empty file | ", filename)
		return nil
	}
	return gozstd.Compress(nil, stdout)
}

// FilterFromFile return the text that containt "toFilter" from the file "filename" in a zipped format
func FilterFromFile(filename string, maxLinesToSearch int, toFilter string, reverse bool) []byte {
	log.Trace("FilterFromFile | START")
	var reverse_on string
	reverse_on = ""
	toFilter = strings.Replace(toFilter, "\"", "\\\"", -1) // Replace the ' " ' in a "grep compliant" format
	if reverse == true {                                   // if the reverse enable add -v for the reverse search
		reverse_on = " -v"
	}
	app := "tail -n " + strconv.Itoa(maxLinesToSearch) + "  " + filename + "|egrep -i \"" + toFilter + "\"" + reverse_on
	cmd := exec.Command("/bin/sh", "-c", app)
	stdout, err := cmd.Output()
	if err != nil { // No file found
		log.Warn("FilterFromFile | No text found [", toFilter, "] in file [", filename, "]")
		return gozstd.Compress(nil, []byte("No text found :("))
	}
	log.Trace("FilterFromFile | STOP | Ok, compressing the data ...")
	return gozstd.Compress(nil, stdout)
}

// ListFiles return the list of files in every subdirectory given in input.
// stdout return a new line at the end, so, after the split, the last element (-1) is discarded.
// TODO: Verify WTF i've thinked when i've discarded the last element
func ListFiles(dirName string) []string {
	log.Trace("ListFiles | START")
	app := "find -L " + dirName + " -type f" // Extract only the name of the file
	cmd := exec.Command("/bin/sh", "-c", app)
	stdout, err := cmd.Output()
	if err != nil { // Nothing in folder ?
		log.Error(err.Error() + ": " + string(stdout))
		return nil
	}
	tmp := strings.Split(string(stdout), "\n")
	log.Info("ListFiles | STOP")
	return tmp[:len(tmp)-1]
}

// GetFileModification return the last modification time of the file in input in a UNIX time format
func GetFileModification(filepath string) int64 {
	f, err := os.Open(filepath)
	if err != nil {
		log.Error("No file :/", filepath)
		//f.Close()
		return -1
	}
	defer f.Close()
	statinfo, err := f.Stat()
	if err != nil {
		log.Error("Error getting stats of file -> :/", filepath)
		//f.Close()
		return -1
	}
	return statinfo.ModTime().Unix()

}

// CountLine return the number of line for a given file using the "wc" shell utils (CPU hungry)
func CountLine(filename string) int {
	app := "wc -l \"" + filename + "\""
	cmd := exec.Command("/bin/sh", "-c", app)
	stdout, err := cmd.Output()
	if err != nil { // File deleted ?
		log.Error(err.Error() + ": " + string(stdout))
		return -1
	}
	n, err := strconv.Atoi(strings.Split(string(stdout), " ")[0]) //Extract files number
	if err != nil {
		log.Error(err)
		return -1
	}
	return n
}
