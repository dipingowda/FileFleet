//
// Generate word count of a set of files. Uses goroutines
// to allow concurrent processing.
//

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"
)

type Count struct {
	fname string
	count int
}
type FileDetails struct {
	Name    string
	Size    int64
	ModTIme time.Time
}

///Counting WORDS in files

func countFile_chan(fname string, c chan map[string]int) {
	count := map[string]int{}
	f, err := os.Open(fname)
	if err == nil {
		scanner := bufio.NewScanner(f)
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			word := scanner.Text()
			count[word]++
		}
	} else {
		fmt.Println("countFile: error reading", fname)
	}
	f.Close() // important: close the file
	c <- count
}

// COUNTING ALL THE FILES

func countAllFiles() {
	// countChan is a buffered channel to avoid "too many files open" error
	var countChan = make(chan map[string]int, 200)

	count := map[string]int{}

	fileNames := getPwdFiles()

	fmt.Printf("%d files in pwd ...\n", len(fileNames))

	// launch all the goroutines
	for _, file := range fileNames {
		go countFile_chan(file.Name, countChan)
	}

	// drain the channel
	for i := 0; i < len(fileNames); i++ {
		m := <-countChan // wait for a result on the channel
		for w, c := range m {
			count[w] += c
		}
	}

	fmt.Println(len(count), " different words")

	wc := 0
	for _, val := range count {
		wc += val
	}

	fmt.Println(wc, " total words")
}

func getPwdFiles() []FileDetails {
	files, err := ioutil.ReadDir("./")
	result := []FileDetails{}
	if err == nil {
		for _, f := range files {
			result = append(result, FileDetails{
				Name:    f.Name(),
				Size:    f.Size(),
				ModTIme: f.ModTime(),
			})
		}
	} else {
		panic("ReadDir error")
	}
	return result
}

func printFiles(files []FileDetails) {
	w := tabwriter.NewWriter(os.Stdout, 20, 0, 1, ' ', 0)

	// Print top border
	fmt.Fprintln(w, "\n+----------------------------------------------------------------------------+")

	// Print header
	fmt.Fprintln(w, "| File Name\t| Size\t| Last Modified\t|")

	// Print separator
	fmt.Fprintln(w, "+----------------------------------------------------------------------------+")

	// Print each file
	for _, f := range files {
		fmt.Fprintf(w, "| %s\t| %d bytes\t| %s\t|\n", f.Name, f.Size, f.ModTIme.Format(time.RFC3339))
	}

	// Print bottom border
	fmt.Fprintln(w, "+----------------------------------------------------------------------------+")

	w.Flush()
}

func CreateFile(fileName string) {
	_, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("Failed to create file: %s, error : %s\n", fileName, err)
	}
}

func openFile(fileName string) {
	// Try to open the file
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		// If the file doesn't exist, create it
		_, err := os.Create(fileName)
		if err != nil {
			fmt.Printf("Failed to create file: %s, error : %s\n", fileName, err)
			return
		}
	} else if err != nil {
		// If there was an error checking if the file exists
		fmt.Printf("Failed to open file: %s, error : %s\n", fileName, err)
		return
	}

	// Open the file
	cmd := exec.Command("cmd", "/C", "start", fileName)
	err = cmd.Start()
	if err != nil {
		fmt.Printf("Failed to open file: %s, error : %s\n", fileName, err)
	}
}

func deleteFile(filename string) {
	err := os.Remove(filename)
	if err != nil {
		fmt.Printf("Failed to delete file: %s, error : %s\n", filename, err)
	}
}

func runCommand(commandStr string) (err error) {
	commandStr = strings.TrimSuffix(commandStr, "\n")
	arrCommandStr := strings.Fields(commandStr)
	switch arrCommandStr[0] {
	case "exit":
		os.Exit(0)
	case "list":
		go func() {
			files := getPwdFiles()
			printFiles(files)
		}()
	case "open":
		if len(arrCommandStr) < 2 {
			fmt.Println("Please provide at least one file name")
		} else {
			for _, fileName := range arrCommandStr[1:] {
				go openFile(fileName)
			}
		}
	case "delete":
		if len(arrCommandStr) < 2 {
			fmt.Println("Please provide at least one file name")
		} else {
			for _, fileName := range arrCommandStr[1:] {
				go deleteFile(fileName)
			}
		}
	case "create":
		if len(arrCommandStr) < 2 {
			fmt.Println("Please provide at least one file name")
		} else {
			for _, fileName := range arrCommandStr[1:] {
				go CreateFile(fileName)
			}
		}
	default:
		err = fmt.Errorf("unknown command")
	}
	return
}
func runCommands(input string) error {
	// Split the input into separate commands
	commands := strings.Split(input, ";")

	// Run each command
	for _, commandStr := range commands {
		commandStr = strings.TrimSpace(commandStr) // Remove leading and trailing whitespace
		err := runCommand(commandStr)
		if err != nil {
			return err
		}
	}

	return nil
}
func main() {
	nCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(nCPU)

	// countChan is a buffered channel to avoid "too many files open" error
	var countChan = make(chan map[string]int, 200)

	count := map[string]int{}

	fileNames := getPwdFiles()

	fmt.Printf("%d files in pwd \n", len(fileNames))

	// launch all the goroutines
	for _, fname := range fileNames {
		go countFile_chan(fname.Name, countChan)
	}

	// drain the channel
	for i := 0; i < len(fileNames); i++ {
		m := <-countChan // wait for a result on the channel
		for w, c := range m {
			count[w] += c
		}
	}

	// Start the command-line interface
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter command: ")
		cmdString, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		err = runCommands(cmdString)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
