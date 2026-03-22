package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var interval int

var rootCmd = &cobra.Command{
	Use:   "ovs-watch",
	Short: "ovs-watch daemon",
}

var bridgeCmd = &cobra.Command{
	Use:   "bridge",
	Short: "watch bridge",
	Run: func(cmd *cobra.Command, args []string) {
		out, err := loadExistingBridges(Scanner)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		watchBridge(out, Scanner, time.Duration(interval)*time.Second)
	},
}

func init() {
	rootCmd.AddCommand(bridgeCmd)
}
func main() {
	fmt.Println("ovs-watch initilized....")

	bridgeCmd.Flags().IntVarP(&interval, "interval", "i", 5, "interval value")
	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func loadExistingBridges(Scanner func([]byte) map[string]bool) (map[string]bool, error) {
	cmd := exec.Command("ovs-vsctl", "list-br")
	output, err := cmd.Output()

	if err != nil {
		return nil, err
	}
	existingBridges := Scanner(output)

	return existingBridges, err
}

func watchBridge(loadedBridges map[string]bool, Scanner func([]byte) map[string]bool, interval time.Duration) {
	for {
		cmd := exec.Command("ovs-vsctl", "list-br")
		out, _ := cmd.Output()
		CurrentBridges := Scanner(out)
		var keysToDel []string
		var keysToAdd []string

		// check for newly added bridges
		for bridge := range CurrentBridges {
			if loadedBridges[bridge] {
				continue
			} else {
				keysToAdd = append(keysToAdd, bridge)
				fmt.Printf("Created: %v\n", bridge)
			}

		}
		// check for deleted bridges
		for bridge := range loadedBridges {
			if CurrentBridges[bridge] {
				continue
			} else {
				keysToDel = append(keysToDel, bridge)
				fmt.Printf("Deleted: %v\n", bridge)
			}
		}
		for _, bridge := range keysToAdd {
			loadedBridges[bridge] = true
		}
		for _, bridge := range keysToDel {
			delete(loadedBridges, bridge)
		}
		time.Sleep(interval)
	}

}

func Scanner(input []byte) map[string]bool {
	outPut := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(input)))
	for scanner.Scan() {
		line := scanner.Text()
		outPut[line] = true
	}
	return outPut
}
