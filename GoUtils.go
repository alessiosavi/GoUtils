package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/pierrec/lz4"
	log "github.com/sirupsen/logrus" // Pretty log library, not the fastest (zerolog/zap)
	"github.com/valyala/gozstd"
)

// Random initalizate a new seed using the UNIX Nano time and return an integer between the 2 input value
func Random(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

// IsDirectory Verify if the path provided is a directory
func IsDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
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
		/* debug.FreeOSMemory()
		log.Info("--- Memory freed! ---")
		log.Info(printMemUsage(&m, &g)) */
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
	// For info eon each, see: https://golang.org/pkg/runtime/#MemStats
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

// ReadFilePath is delegated to filter every (sub)file path from a given directory
func ReadFilePath(path string) []string {
	if !ValidateInjection(path, nil) {
		log.Error("readFilePath | Path is not valid :/ ", path)
		return nil
	}
	fileList := []string{}
	// Read all the file recursivly
	log.Debug("readFilePath | Reading file in ", path)
	err := filepath.Walk(path, func(file string, f os.FileInfo, err error) error {
		if IsFile(file) {
			log.Trace("Adding file -> ", file)
			fileList = append(fileList, file)
		} else {
			log.Debug("readFilePath | Directory found, skipping! -> ", file)
		}
		return nil
	})
	if err != nil {
		log.Error("readFilePath | Some trouble | Path: ", path, " | ERR: ", err)
		return nil
	}
	log.Warn("readFilePath | [", len(fileList), "] File readed: ", fileList)
	return fileList
}

// FilterFromFile return the text that containt "toFilter" from the file "filename" in a zipped format
func FilterFromFile(filename string, maxLinesToSearch int, toFilter string, reverse bool) []byte {
	log.Trace("FilterFromFile | START")
	reverseOn := ""
	toFilter = strings.Replace(toFilter, "\"", "\\\"", -1) // Replace the ' " ' in a "grep compliant" format
	if reverse {                                           // if the reverse enable add -v for the reverse search
		reverseOn = " -v"
	}
	app := "tail -n " + strconv.Itoa(maxLinesToSearch) + "  " + filename + "|egrep -i \"" + toFilter + "\"" + reverseOn
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
	app := "find -L -O3 " + dirName + " -type f" // Extract only the name of the file
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

// ValidateInjection provide commons methods for validate a given payload
func ValidateInjection(payload string, mustContain []string) bool {

	if len(payload) <= 4 {
		log.Debug("ValidateInjection | payload empty")
		return false
	}
	// Verify if the payload contains one of the list of string that we assume that have to have
	if mustContain != nil {
		log.Debug("ValidateInjection | verify word that must be contained")

		checkContains := false
		for i := 0; i < len(mustContain); i++ {
			if strings.Contains(payload, mustContain[i]) {
				log.Debug("ValidateInjection | Payload [", payload, "] contains [", mustContain[i], "]")
				// Ok, check satisfied, set a flag and exit the iteration
				checkContains = true
				break
			}
		}
		// If the flag is not true, no party
		if !checkContains {
			log.Error("ValidateInjection | Payload [", payload, "] does not contains any of our validation input [", mustContain, "]")
			return false
		}
	}
	evilword := [...]string{"../", "..", "/./", "/etc/", "/bin/", "/usr/", "/var/"}

	log.Debug("ValidateInjection | Trying to find evil word [", payload, "] ...")
	// Verify if the payload contains one of the evilword
	for i := 0; i < len(evilword); i++ {
		if strings.Contains(payload, evilword[i]) {
			log.Error("ValidateInjection | EvilWord [", payload, "] -- [", evilword[i], "]")
			return false
		}
	}
	return true
}

func Lz4CompressData(fileContent string) ([]byte, int) {
	toCompress := []byte(fileContent)
	compressed := make([]byte, len(toCompress))

	//compress
	var ht [1 << 16]int
	lenght, err := lz4.CompressBlock(toCompress, compressed, ht[:])
	if err != nil {
		panic(err)
	}
	log.Warning("compressed Data:", string(compressed[:lenght]))
	return compressed, lenght
}

func Lz4DecompressData(compressedData []byte, l int) {
	//decompress
	decompressed := make([]byte, len(compressedData)*3)
	lenght, err := lz4.UncompressBlock(compressedData[:l], decompressed)
	if err != nil {
		panic(err)
	}
	log.Warning("\ndecompressed Data:", string(decompressed[:lenght]))
}

//IsFile verify if a give filepath is a directory
func IsFile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		log.Fatal("isDir | Fatal on path ", path, " | ERR: ", err)
		return false
	}
	// fi.IsDir()
	return !fi.Mode().IsDir()
}

/* Remove a given element from a string*/
func RemoveFromString(s []byte, i int) []byte {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}

func ReadAllFileInArray(filePath string) []string {
	// Open a file.
	var file *os.File // File to read
	var err error
	var linesList []string
	//var result string
	log.Trace("ReadAllFile | START")
	file, err = os.Open(filePath)
	if err != nil {
		log.Error("ReadAllFile | Error opening file [", filePath, "]")
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Clean the string and remove the double space
		linesList = append(linesList, strings.Replace(strings.TrimSpace(scanner.Text()), "  ", "", -1))
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return linesList
}

func ReadAllFile(filePath string) string {
	// Open a file.
	var file *os.File // File to read
	var err error
	var reader *bufio.Reader
	var content []byte
	//var result string
	log.Trace("ReadAllFile | START")
	file, err = os.Open(filePath)
	if err != nil {
		log.Error("ReadAllFile | Error opening file [", filePath, "]")
		return ""
	}
	defer file.Close()

	// Use bufio.NewReader to get a Reader.
	// ... Then use ioutil.ReadAll to read the entire content.
	reader = bufio.NewReader(file)
	content, err = ioutil.ReadAll(reader)
	if err != nil {
		log.Error("ReadAllFile | Error reading file [", filePath, "]")
		return ""
	}
	log.Trace("ReadAllFile | STOP")
	return string(content)
}

// IsASCII is delegated to verify if a given string is ASCII compliant
func IsASCII(s string) bool {
	for _, c := range s {
		if c > 127 {
			return false
		}
	}
	return true
}

//RemoveElement delete the element of the indexes contained in j of the data in input
func RemoveElement(data []string, j []int) []string {
	for i := 0; i < len(j); i++ {
		data[j[i]] = data[len(data)-1] // Copy last element to index i.
		data[len(data)-1] = ""         // Erase last element (write zero value).
		data = data[:len(data)-1]      // Truncate slice.
	}
	return data
}

// SetDebugLevel return the LogRus object by the given string
func SetDebugLevel(level string) log.Level {
	if strings.Compare(strings.ToLower(level), "debug") == 0 {
		return log.DebugLevel
	} else if strings.Compare(strings.ToLower(level), "trace") == 0 {
		return log.TraceLevel
	} else if strings.Compare(strings.ToLower(level), "info") == 0 {
		return log.InfoLevel
	} else if strings.Compare(strings.ToLower(level), "error") == 0 {
		return log.ErrorLevel
	} else if strings.Compare(strings.ToLower(level), "fatal") == 0 {
		return log.FatalLevel
	} else if strings.Compare(strings.ToLower(level), "panic") == 0 {
		return log.PanicLevel
	} else if strings.Contains(strings.ToLower(level), "warn") {
		return log.WarnLevel
	}
	return log.DebugLevel
}

// ByteCountSI convert the byte in input to MB/KB/TB ecc
func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

// ByteCountIEC convert the byte in input to MB/KB/TB ecc
func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

// RetrieveLines return the number of lines in the given string
func RetrieveLines(fileContet string) int {
	scanner := bufio.NewScanner(strings.NewReader(fileContet)) // Create a scanner for iterate the string
	counter := 0
	for scanner.Scan() {
		counter++
	}
	return counter
}

// ParseDate2 is an hardcoded parser for date like 31/01/2019 13:29:37,932. He reurn the nanoseceond since Unix epoch
func ParseDate2(strdate string) int64 {
	day, _ := strconv.Atoi(strdate[:2])
	month, _ := strconv.Atoi(strdate[3:5])
	year, _ := strconv.Atoi(strdate[6:10])
	hour, _ := strconv.Atoi(strdate[11:13])
	minute, _ := strconv.Atoi(strdate[14:16])
	second, _ := strconv.Atoi(strdate[17:19])
	milli, _ := strconv.Atoi(strdate[20:23])
	return time.Date(year, time.Month(month), day, hour, minute, second, milli*1000000, time.UTC).UnixNano() / 1000000
}

func RemoveWhiteSpaceArray(data []string) []string {
	var toDelete []int
	// Iterate the string in the list
	for i := 0; i < len(data); i++ {
		// Iterate the char in the string
		for _, c := range data[i] {
			if c == 32 { // if whitespace
				toDelete = append(toDelete, i)
			}
		}
	}
	return RemoveElement(data, toDelete)
}

// Join is a quite efficient string concatenator
func Join(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}

// JoinArray concatenate every data in the array and return the string content
func JoinArray(data []string) string {
	var sb strings.Builder
	for i := 0; i < len(data); i++ {
		sb.WriteString(data[i] + " ")
	}
	return sb.String()

}

// VerifyIfPresent Verify if a given string is present in the list
func VerifyIfPresent(content string, entryList []string) bool {
	for i := 0; i < len(entryList); i++ {
		if strings.Contains(content, entryList[i]) {
			return true
		}
	}
	return false
}

// StartCPUProfiler Save the cpu profile information into the given file
// NOTE: Remember to defer pprof.StopCPUProfile() after the function
func StartCPUProfiler(file *os.File) {
	// Start the cpu profiler
	if err := pprof.StartCPUProfile(file); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
}

//ExtractString is delegated to filter the content of the given data delimited by 'first' and 'last' string
func ExtractString(data *string, first, last string) string {
	// Find the first instance of 'start' in the give string data
	startHeder := strings.Index(*data, first)
	// Found !
	if startHeder != -1 {
		// Remove the first word
		startHeder += len(first)
		// Check the first occurrence of 'last' that delimit the string to return
		endHeader := strings.Index((*data)[startHeder:], last)
		// Ok, seems good, return the content of the string delimited by 'first' and 'last'
		if endHeader != -1 {
			return (*data)[startHeder : startHeder+endHeader]
		}
	}
	return ""
}

// recognizeFormat is delegated to valutate the extension and return the properly Mimetype by a given format type
// reurn: (Mimetype http compliant,Content-Disposition header value)
func RecognizeFormat(input string) (string, string) {
	// Find the last occurrence of the dot
	// extract only the extension of the file by slicing the string
	var contentDisposition string
	var mimeType string

	contentDisposition = `inline; filename="` + input + `"`
	switch input[strings.LastIndex(input, ".")+1:] {
	case "doc":
		mimeType = "application/msword"
		//mimeType = "application/octet-stream"
	case "docx":
		mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		//mimeType = "application/octet-stream"
	case "pdf":
		mimeType = "application/pdf"
		//contentDisposition = `inline; filename="` + input + `"`
	default:
		mimeType = "application/octet-stream"
		contentDisposition = `inline; filename="` + input + `"`
	}
	return mimeType, contentDisposition
}

func RemoveWhiteSpaceString(str string) string {
	var b strings.Builder
	b.Grow(len(str))
	for i, ch := range str {
		//fmt.Println("I: ", i)
		if !(str[i] == 32 && (i+1 < len(str) && str[i+1] == 32)) {
			b.WriteRune(ch)
		}
	}
	return b.String()
}
