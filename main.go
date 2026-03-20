package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

func main() {
	out, err := loadExistingBridges(Scanner)
	fmt.Printf("%v\n", out)
	// watchBridge(out, Scanner)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

}

func loadExistingBridges(Scanner func([]byte) []string) (map[string]bool, error) {
	existingBridges := make(map[string]bool)
	cmd := exec.Command("ovs-vsctl", "list-br")
	output, err := cmd.Output()

	if err != nil {
		return nil, err
	}
	bridges := Scanner(output)
	for _, bridge := range bridges {
		existingBridges[bridge] = true
	}

	return existingBridges, err
}

func watchBridge(loadedBridges map[string]bool, Scanner func([]byte) []string) {
	count := len(loadedBridges)
	// found := true
	for {
		cmd := exec.Command("ovs-vsctl", "list-br")
		out, _ := cmd.CombinedOutput()
		bridges := Scanner(out)
		if len(bridges) == count {
			continue
		}
		if len(bridges) > count {
			fmt.Printf("Created: %v\n", bridges[count])
			count++
		}

	}
}

func Scanner(input []byte) []string {
	var outPut []string
	scanner := bufio.NewScanner(strings.NewReader(string(input)))
	for scanner.Scan() {
		line := scanner.Text()
		outPut = append(outPut, line)
	}
	return outPut
}
