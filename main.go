package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func main() {
	out, err := loadExistingBridges(Scanner)
	fmt.Println("ovs-watch initilized..")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	watchBridge(out, Scanner)

}

func loadExistingBridges(Scanner func([]byte, bool) map[string]bool) (map[string]bool, error) {
	cmd := exec.Command("ovs-vsctl", "list-br")
	output, err := cmd.Output()

	if err != nil {
		return nil, err
	}
	existingBridges := Scanner(output, true)

	return existingBridges, err
}

func watchBridge(loadedBridges map[string]bool, Scanner func([]byte, bool) map[string]bool) {
	for {
		cmd := exec.Command("ovs-vsctl", "list-br")
		out, _ := cmd.Output()
		CurrentBridges := Scanner(out, true)

		// check for newly added bridges
		for bridge := range CurrentBridges {

			if loadedBridges[bridge] {
				continue
			} else {
				loadedBridges[bridge] = true
				fmt.Printf("Created: %v\n", bridge)
			}

		}

		// check for deleted bridges
		for bridge := range loadedBridges {
			if CurrentBridges[bridge] {
				continue
			} else {
				delete(loadedBridges, bridge)
				fmt.Printf("Deleted: %v\n", bridge)

			}
		}
		time.Sleep(5 * time.Second)

	}
}

func Scanner(input []byte, value bool) map[string]bool {
	outPut := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(input)))
	for scanner.Scan() {
		line := scanner.Text()
		outPut[line] = value
	}
	return outPut
}
