package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

type FileSource struct {
	Path string `json:"path"`
}

type TestCase struct {
	Name          string          `json:"name"`
	Points        int             `json:"points"`
	Args          []string        `json:"args"`
	Stdin         json.RawMessage `json:"stdin"`
	Stdout        json.RawMessage `json:"stdout"`
	CheckStderr   bool            `json:"check_stderr"`
	CheckCommand  string          `json:"check_command"`
}

func getInput(data json.RawMessage) []byte {
	if len(data) == 0 {
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		return []byte(s)
	}

	var fileSource FileSource
	if err := json.Unmarshal(data, &fileSource); err == nil {
		if _, err := os.Stat(fileSource.Path); os.IsNotExist(err) {
			fmt.Printf("Error: file not found for stdin/stdout: %s\n", fileSource.Path)
			os.Exit(1)
		}
		content, err := ioutil.ReadFile(fileSource.Path)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		return content
	}

	fmt.Printf("Invalid format for stdin/stdout: %s\n", string(data))
	os.Exit(1)
	return nil
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: ./checker <path_to_executable> <path_to_json_file>")
		os.Exit(1)
	}

	executablePath := os.Args[1]
	jsonPath := os.Args[2]

	jsonFile, err := os.Open(jsonPath)
	if err != nil {
		fmt.Printf("Error opening JSON file: %v\n", err)
		os.Exit(1)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var testCases []TestCase
	json.Unmarshal(byteValue, &testCases)

	totalPoints := 0
	earnedPoints := 0

	for _, testCase := range testCases {
		totalPoints += testCase.Points

		cmd := exec.Command(executablePath, testCase.Args...)

		stdinBytes := getInput(testCase.Stdin)
		stdinPipe, _ := cmd.StdinPipe()
		go func() {
			defer stdinPipe.Close()
			stdinPipe.Write(stdinBytes)
		}()

		output, err := cmd.CombinedOutput()

		if err != nil && testCase.CheckStderr {
			fmt.Printf("%s: fail\n", testCase.Name)
			continue
		}

		if testCase.Stdout != nil {
			expectedStdout := getInput(testCase.Stdout)
			if string(output) != string(expectedStdout) {
				fmt.Printf("%s: fail\n", testCase.Name)
				continue
			}
		}

		if testCase.CheckCommand != "" {
			cmd := exec.Command("sh", "-c", testCase.CheckCommand)
			if err := cmd.Run(); err != nil {
				fmt.Printf("%s: fail\n", testCase.Name)
				continue
			}
		}

		fmt.Printf("%s: ok\n", testCase.Name)
		earnedPoints += testCase.Points
	}

	fmt.Printf("\nTotal points: %d/%d\n", earnedPoints, totalPoints)
}
