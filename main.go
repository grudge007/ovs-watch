package main

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var interval int
var bridgeName string
var watchIfaceState bool

var rootCmd = &cobra.Command{
	Use:   "ovs-watch",
	Short: "ovs-watch daemon",
}

var bridgeCmd = &cobra.Command{
	Use:   "bridge",
	Short: "watch bridge",
	Run: func(cmd *cobra.Command, args []string) {

		loadedBridges, err := LoadExistingBridges()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		watchBridge(loadedBridges, time.Duration(interval)*time.Second)
	},
}
var portCmd = &cobra.Command{
	Use:   "port",
	Short: "watch ports",
	Run: func(cmd *cobra.Command, args []string) {

		loadedBridges, err := LoadExistingBridges()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		loadedPorts := LoadExistingPorts(loadedBridges, bridgeName)
		if loadedPorts == nil {
			loadedPorts = make(map[string]string)
		}

		fmt.Println("Using bridge:", bridgeName)

		watchPort(loadedPorts, time.Duration(interval)*time.Second, bridgeName)
	},
}

var ifaceCmd = &cobra.Command{
	Use:   "iface",
	Short: "watch interface state",
	Run: func(cmd *cobra.Command, args []string) {
		watchInterfaceState(time.Duration(interval) * time.Second)
	},
}

func init() {
	rootCmd.AddCommand(bridgeCmd)
	rootCmd.AddCommand(portCmd)
	rootCmd.AddCommand(ifaceCmd)
}

func main() {
	fmt.Println("ovs-watch initialized....")

	bridgeCmd.Flags().IntVarP(&interval, "interval", "i", 5, "interval value")
	bridgeCmd.Flags().BoolVarP(&watchIfaceState, "watch-iface", "w", false, "Enable interface state watch")

	portCmd.Flags().IntVarP(&interval, "interval", "i", 5, "interval value")
	portCmd.Flags().StringVarP(&bridgeName, "bridge", "b", "", "bridge name")
	portCmd.Flags().BoolVarP(&watchIfaceState, "watch-iface", "w", false, "Enable interface state watch")

	ifaceCmd.Flags().IntVarP(&interval, "interval", "i", 5, "interval value")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// bridge watch
func watchBridge(loadedBridges map[string]bool, interval time.Duration) {

	for {
		CurrentBridges, err := LoadExistingBridges()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		var keysToDel []string
		// check for newly added bridges
		for bridge := range CurrentBridges {
			if loadedBridges[bridge] {
				continue
			} else {
				fmt.Printf("Created: %v\n", bridge)
				loadedBridges[bridge] = true
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
		for _, bridge := range keysToDel {
			delete(loadedBridges, bridge)
		}
		time.Sleep(interval)
	}

}

// port watch
func watchPort(loadedPorts map[string]string, interval time.Duration, bridgeName string) {
	var loadedBridges map[string]bool
	switch bridgeName {
	default:
		loadedPorts := LoadExistingPorts(loadedBridges, bridgeName)
		for {
			var portUpdates []string
			var portDegrades []string
			currentPorts := LoadExistingPorts(loadedBridges, bridgeName)
			for ports := range currentPorts {
				_, isExist := loadedPorts[ports]
				if !isExist {
					fmt.Printf("New Port Detected: %v  ->  %v \n", ports, bridgeName)
					portUpdates = append(portUpdates, ports)
					loadedPorts[ports] = bridgeName
				}
			}

			for _, key := range portUpdates {
				loadedPorts[key] = bridgeName
			}

			for ports := range loadedPorts {
				_, isExist := currentPorts[ports]
				if !isExist {
					fmt.Printf("Port Deletion Detected: %v\n", ports)
					portDegrades = append(portDegrades, ports)
				}
			}

			for _, key := range portDegrades {
				delete(loadedPorts, key)
			}
			time.Sleep(interval)
		}

	case "":
		for {
			portUpdates := make(map[string]string)
			var portDegrades []string
			loadedBridges, _ := LoadExistingBridges()
			currentPorts := LoadExistingPorts(loadedBridges, bridgeName)
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

}

// watch port state
func watchInterfaceState(interval time.Duration) {
	loadedIfaceState := LoadInterfaceStatus()
	for {
		portUpdates := make(map[string]string)
		var portDegrades []string
		currentIfaceState := LoadInterfaceStatus()
		for iface, state := range currentIfaceState {
			oldState, isExist := loadedIfaceState[iface]
			if !isExist {
				fmt.Printf("New Port Detected: %v\n", iface)
				portUpdates[iface] = state
				continue
			}

			if oldState != state {
				fmt.Printf("State of %v Changed To %v\n", iface, state)
				portUpdates[iface] = state
				continue
			}
		}
		for iface := range loadedIfaceState {
			_, isExist := currentIfaceState[iface]
			if !isExist {
				fmt.Printf("Port Deletion Detected: %v\n", iface)
				portDegrades = append(portDegrades, iface)
			}
		}

		for _, iface := range portDegrades {
			delete(loadedIfaceState, iface)
		}
		maps.Copy(loadedIfaceState, portUpdates)
		time.Sleep(interval)
	}

}

// utils
// load bridges
func LoadExistingBridges() (map[string]bool, error) {
	cmd := exec.Command("ovs-vsctl", "list-br")
	output, err := cmd.Output()

	if err != nil {
		return nil, err
	}
	existingBridges := SliceScannerSet(output)

	return existingBridges, err
}

// load ports
func LoadExistingPorts(loadedBridges map[string]bool, bridgeName string) map[string]string {
	loadedPorts := make(map[string]string)
	var loadedPortsSlice []string
	switch bridgeName {
	default:
		cmd := exec.Command("ovs-vsctl", "list-ports", bridgeName)
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Err: %v\n", err)
			os.Exit(1)
		}
		loadedPortsSlice = SliceScanner(output)
		for _, port := range loadedPortsSlice {
			loadedPorts[port] = bridgeName
		}
		return loadedPorts

	case "":
		for bridge := range loadedBridges {
			cmd := exec.Command("ovs-vsctl", "list-ports", bridge)
			output, err := cmd.Output()
			if err != nil {
				continue
			}
			loadedPortsSlice = SliceScanner(output)
			for _, port := range loadedPortsSlice {
				loadedPorts[port] = bridge
			}

		}
		return loadedPorts
	}

}

// load interface status
func LoadInterfaceStatus() map[string]string {
	loadedIfaceState := make(map[string]string)
	cmd := exec.Command("ovs-vsctl", "--format=csv", "--columns=name,link_state", "list", "Interface")
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("Err: %v", err)
		os.Exit(1)
	}
	outPut := SliceScanner(out)

	for i, rawLine := range outPut {
		if i == 0 {
			continue
		}
		ifaceRaw := strings.Split(rawLine, ",")
		loadedIfaceState[ifaceRaw[0]] = ifaceRaw[1]

	}
	return loadedIfaceState

}

func SliceScanner(input []byte) []string {
	var outPutSlice []string
	scanner := bufio.NewScanner(strings.NewReader(string(input)))
	for scanner.Scan() {
		line := scanner.Text()
		outPutSlice = append(outPutSlice, line)
	}
	return outPutSlice
}

func SliceScannerSet(input []byte) map[string]bool {
	outPutMap := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(input)))
	for scanner.Scan() {
		line := scanner.Text()
		outPutMap[line] = true
	}
	return outPutMap
}
