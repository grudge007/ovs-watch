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
var loadedBridges map[string]bool
var loadedPorts map[string]string

var rootCmd = &cobra.Command{
	Use:   "ovs-watch",
	Short: "ovs-watch daemon",
}

var bridgeCmd = &cobra.Command{
	Use:   "bridge",
	Short: "watch bridge",
	Run: func(cmd *cobra.Command, args []string) {
		watchBridge(loadedBridges, time.Duration(interval)*time.Second)
	},
}

var portCmd = &cobra.Command{
	Use:   "port",
	Short: "watch ports",
	Run: func(cmd *cobra.Command, args []string) {
		watchPort(loadedPorts, time.Duration(interval)*time.Second)
	},
}

func init() {
	rootCmd.AddCommand(bridgeCmd)
	rootCmd.AddCommand(portCmd)
}

func main() {
	var err error
	fmt.Println("ovs-watch initilized....")

	loadedBridges, err = LoadExistingBridges()
	loadedPorts = LoadExistingPorts(loadedBridges)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	bridgeCmd.Flags().IntVarP(&interval, "interval", "i", 5, "interval value")
	portCmd.Flags().IntVarP(&interval, "interval", "i", 5, "interval value")
	err = rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

// bridge watch
func watchBridge(loadedBridges map[string]bool, interval time.Duration) {
	for {
		cmd := exec.Command("ovs-vsctl", "list-br")
		out, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		CurrentBridges, _ := SliceScanner(out)
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

// port watch
func watchPort(loadedPorts map[string]string, interval time.Duration) {

	for {
		portUpdates := make(map[string]string)
		var portDegrades []string
		loadedBridges, _ := LoadExistingBridges()
		currentPorts := LoadExistingPorts(loadedBridges)
		for ports, bridge := range currentPorts {
			oldBridge, exists := loadedPorts[ports]
			if exists && oldBridge == bridge {
				continue
			}
			if !exists {
				fmt.Printf("New Port Detected: %v  ->  %v \n", ports, bridge)
				portUpdates[ports] = bridge
			} else {
				fmt.Printf("Updated Detected: %v from %v to %v\n", ports, oldBridge, bridge)
				portUpdates[ports] = bridge
			}
		}
		for key, value := range portUpdates {
			loadedPorts[key] = value
		}

		for port := range loadedPorts {
			_, isExist := currentPorts[port]
			if !isExist {
				portDegrades = append(portDegrades, port)
				fmt.Printf("Port Deletion Detected: %v\n", port)
			}
		}
		for _, bridge := range portDegrades {
			delete(loadedPorts, bridge)
		}
		time.Sleep(interval)

	}

}

// utils
func LoadExistingBridges() (map[string]bool, error) {
	cmd := exec.Command("ovs-vsctl", "list-br")
	output, err := cmd.Output()

	if err != nil {
		return nil, err
	}
	existingBridges, _ := SliceScanner(output)

	return existingBridges, err
}

func LoadExistingPorts(loadedBridges map[string]bool) map[string]string {
	loadedPorts := make(map[string]string)
	var loadedPortsSlice []string
	for bridge := range loadedBridges {
		cmd := exec.Command("ovs-vsctl", "list-ports", bridge)
		output, err := cmd.Output()
		if err != nil {
			continue
		}
		_, loadedPortsSlice = SliceScanner(output)
		for _, port := range loadedPortsSlice {
			loadedPorts[port] = bridge
		}

	}
	return loadedPorts

}

func SliceScanner(input []byte) (map[string]bool, []string) {
	outPutMap := make(map[string]bool)
	var outPutSlice []string
	scanner := bufio.NewScanner(strings.NewReader(string(input)))
	for scanner.Scan() {
		line := scanner.Text()
		outPutMap[line] = true
		outPutSlice = append(outPutSlice, line)
	}
	return outPutMap, outPutSlice
}
