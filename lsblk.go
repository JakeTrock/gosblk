package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

type Partition struct {
	Name          string
	Type          string
	Identifier    string
	Size          string
	FSType        string
	Label         string
	UUID          string
	Mountpoint    string
	PartitionType string
}

type Disk struct {
	Name          string
	Size          string
	Type          string
	Identifier    string
	Partitions    []Partition
	FSType        string
	Label         string
	UUID          string
	Mountpoint    string
	PartitionType string
}

func main() {
	// Define flags
	bytesFlag := flag.Bool("b", false, "Print the SIZE column in bytes rather than in a human-readable format.")
	flag.BoolVar(bytesFlag, "bytes", false, "Print the SIZE column in bytes rather than in a human-readable format.")

	nodepsFlag := flag.Bool("d", false, "Do not print holder devices or replicas.")
	flag.BoolVar(nodepsFlag, "nodeps", false, "Do not print holder devices or replicas.")

	excludeList := flag.String("e", "", "Exclude the devices specified by the comma-separated list of names.")
	flag.StringVar(excludeList, "exclude", "", "Exclude the devices specified by the comma-separated list of names.")

	fsFlag := flag.Bool("f", false, "Output info about filesystems.")
	flag.BoolVar(fsFlag, "fs", false, "Output info about filesystems.")

	helpFlag := flag.Bool("h", false, "Display help text and exit.")
	flag.BoolVar(helpFlag, "help", false, "Display help text and exit.")

	includeList := flag.String("I", "", "Include devices specified by the comma-separated list of names.")
	flag.StringVar(includeList, "include", "", "Include devices specified by the comma-separated list of names.")

	asciiFlag := flag.Bool("i", false, "Use ASCII characters for tree formatting.")
	flag.BoolVar(asciiFlag, "ascii", false, "Use ASCII characters for tree formatting.")

	jsonFlag := flag.Bool("J", false, "Use JSON output format.")
	flag.BoolVar(jsonFlag, "json", false, "Use JSON output format.")

	listFlag := flag.Bool("l", false, "Produce output in the form of a list.")
	flag.BoolVar(listFlag, "list", false, "Produce output in the form of a list.")

	noheadingsFlag := flag.Bool("n", false, "Do not print a header line.")
	flag.BoolVar(noheadingsFlag, "noheadings", false, "Do not print a header line.")

	pathsFlag := flag.Bool("p", false, "Print full device paths.")
	flag.BoolVar(pathsFlag, "paths", false, "Print full device paths.")

	sortColumn := flag.String("x", "", "Sort by column (name, size, type, identifier)")
	flag.StringVar(sortColumn, "sort", "", "Sort by column (name, size, type, identifier)")

	versionFlag := flag.Bool("v", false, "Print version")

	flag.Parse()

	if *helpFlag {
		flag.Usage()
		return
	}

	if *versionFlag {
		fmt.Println("mac lsblk v0.1 made by jake.trock.com :^]")
		return
	}

	// Run diskutil list
	cmd := exec.Command("diskutil", "list")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error running diskutil list: %v\n", err)
		return
	}

	// Run df -h
	dfCmd := exec.Command("df")
	var dfOut bytes.Buffer
	dfCmd.Stdout = &dfOut
	err = dfCmd.Run()
	if err != nil {
		fmt.Printf("Error running df -h: %v\n", err)
		return
	}

	// Get mountpoints
	mounts := getMounts(dfOut.String())
	fmt.Println(mounts)
	disks := parseDiskutilOutput(out.String(), mounts)

	// Apply include and exclude filters
	if *includeList != "" {
		disks = includeDisks(disks, strings.Split(*includeList, ","))
	}
	if *excludeList != "" {
		disks = excludeDisks(disks, strings.Split(*excludeList, ","))
	}

	if *nodepsFlag {
		// Remove partitions
		for i := range disks {
			disks[i].Partitions = nil
		}
	}

	if *sortColumn != "" {
		sortDisks(disks, *sortColumn)
	}

	// Handle bytes flag
	if *bytesFlag {
		for i := range disks {
			disks[i].Size = sizeToBytes(disks[i].Size)
			for j := range disks[i].Partitions {
				disks[i].Partitions[j].Size = sizeToBytes(disks[i].Partitions[j].Size)
			}
		}
	}

	// Handle fs flag
	if *fsFlag {
		// Get filesystem info for disks and partitions
		for i := range disks {
			getFilesystemInfo(&disks[i])
			for j := range disks[i].Partitions {
				getPartitionFilesystemInfo(&disks[i].Partitions[j])
			}
		}
	}

	if *jsonFlag {
		jsonOutput(disks)
		return
	}

	// Prepare to print in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !*noheadingsFlag {
		if *fsFlag {
			fmt.Fprintln(w, "NAME\tSIZE\tFSTYPE\tLABEL\tUUID\tMOUNTPOINT")
		} else {
			fmt.Fprintln(w, "NAME\tSIZE\tTYPE\tIDENTIFIER")
		}
	}

	for _, disk := range disks {
		name := disk.Name
		if *pathsFlag {
			name = "/dev/" + name
		}
		fmt.Fprintf(w, "%s\t%s", name, disk.Size)

		if *fsFlag {
			fmt.Fprintf(w, "\t%s\t%s\t%s\t%s\n", disk.FSType, disk.Label, disk.UUID, disk.Mountpoint)
		} else {
			fmt.Fprintf(w, "\t%s\t%s\n", disk.Type, disk.Identifier)
		}

		if !*nodepsFlag {
			for ind, part := range disk.Partitions {
				partName := part.Name
				if *pathsFlag {
					partName = "/dev/" + part.Identifier
				}

				prefix := "├─"
				if ind == len(disk.Partitions)-1 {
					prefix = "└─"
				}
				if *asciiFlag {
					if prefix == "├─" {
						prefix = "|-"
					} else {
						prefix = "`-"
					}
				}

				fmt.Fprintf(w, "%s%s\t%s", prefix, partName, part.Size)

				if *fsFlag {
					fmt.Fprintf(w, "\t%s\t%s\t%s\t%s\n", part.FSType, part.Label, part.UUID, part.Mountpoint)
				} else {
					fmt.Fprintf(w, "\t%s\t%s\n", part.Type, part.Identifier)
				}
			}
		}
	}

	w.Flush()
}

func getMounts(dfOutput string) map[string]string {
	mounts := make(map[string]string)
	lines := strings.Split(dfOutput, "\n")

	// Skip the header line
	if len(lines) < 2 {
		return mounts
	}

	// Regex pattern to match the filesystem and mounted on columns
	pattern := `^(\S+)\s+\d+\s+\d+\s+\d+\s+\d+%\s+\d+\s+\d+\s+\d+%\s+(.+)$`
	re := regexp.MustCompile(pattern)
	for _, line := range lines[1:] {
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			filesystem := strings.TrimSpace(matches[1])
			mountedOn := strings.TrimSpace(matches[2])
			if filesystem != "" && mountedOn != "" {
				mounts[filesystem] = mountedOn
			}
		}
	}

	return mounts
}

func parseDiskutilOutput(output string, mounts map[string]string) []Disk {
	lines := strings.Split(output, "\n")
	var disks []Disk
	var currentDisk *Disk
	diskHeaderRegex := regexp.MustCompile(`^\/dev\/(disk\d+)`)
	diskInfoRegex := regexp.MustCompile(`^\/dev\/(disk\d+).*?\*\s*([\d\.]+\s\w+).*`)
	partitionRegex := regexp.MustCompile(`^\s+(\d+):\s+(\S+)\s+(.*?)\s+([\d\.]+\s\w+)\s+(\S+)$`)

	for _, line := range lines {
		if diskHeaderRegex.MatchString(line) {
			matches := diskHeaderRegex.FindStringSubmatch(line)
			currentDisk = &Disk{
				Name:       matches[1],
				Type:       "disk",
				Identifier: matches[1],
			}
			// Call parseInfo to fill additional info
			parseInfo(currentDisk.Identifier, currentDisk, mounts)
			disks = append(disks, *currentDisk)
		} else if diskInfoRegex.MatchString(line) {
			matches := diskInfoRegex.FindStringSubmatch(line)
			currentDisk = &Disk{
				Name:       matches[1],
				Size:       matches[2],
				Type:       "disk",
				Identifier: matches[1],
			}
			parseInfo(currentDisk.Identifier, currentDisk, mounts)
			disks = append(disks, *currentDisk)
		} else if partitionRegex.MatchString(line) && currentDisk != nil {
			matches := partitionRegex.FindStringSubmatch(line)
			partition := Partition{
				Name:       matches[3],
				Type:       matches[2],
				Size:       matches[4],
				Identifier: matches[5],
			}
			parseInfo(matches[5], currentDisk, mounts)
			// Add partition to the last disk in the slice
			diskIndex := len(disks) - 1
			disks[diskIndex].Partitions = append(disks[diskIndex].Partitions, partition)
		}
	}

	return disks
}

func parseInfo(identifier string, disk *Disk, mounts map[string]string) {
	cmd := exec.Command("diskutil", "info", identifier)
	out, err := cmd.Output()
	if err != nil {
		return
	}
	info := string(out)
	disk.Label = getValueForKey(info, "Volume Name:")
	disk.Mountpoint = getValueForKey(info, "Mount Point:")
	disk.FSType = getValueForKey(info, "File System:")
	disk.UUID = getValueForKey(info, "Disk / Partition UUID:")
	disk.PartitionType = getValueForKey(info, "Partition Type:")
	if disk.Mountpoint == "" {
		fmt.Println(disk.Name, identifier, mounts["/dev/"+disk.Name])
		for partition, mountpoint := range mounts {
			if strings.Contains(partition, disk.Name) {
				fmt.Println("jjjjj ", partition, mountpoint)
			}
		}
		disk.Mountpoint = mounts["/dev/"+disk.Name]
	}
	// Add any other fields as needed
}

func sortDisks(disks []Disk, sortColumn string) {
	switch sortColumn {
	case "name":
		sort.Slice(disks, func(i, j int) bool {
			return disks[i].Name < disks[j].Name
		})
	case "size":
		sort.Slice(disks, func(i, j int) bool {
			return compareSizeStrings(disks[i].Size, disks[j].Size)
		})
	case "type":
		sort.Slice(disks, func(i, j int) bool {
			return disks[i].Type < disks[j].Type
		})
	case "identifier":
		sort.Slice(disks, func(i, j int) bool {
			return disks[i].Identifier < disks[j].Identifier
		})
	}

	for i := range disks {
		sortPartitions(disks[i].Partitions, sortColumn)
	}
}

func sortPartitions(partitions []Partition, sortColumn string) {
	switch sortColumn {
	case "name":
		sort.Slice(partitions, func(i, j int) bool {
			return partitions[i].Name < partitions[j].Name
		})
	case "size":
		sort.Slice(partitions, func(i, j int) bool {
			return compareSizeStrings(partitions[i].Size, partitions[j].Size)
		})
	case "type":
		sort.Slice(partitions, func(i, j int) bool {
			return partitions[i].Type < partitions[j].Type
		})
	case "identifier":
		sort.Slice(partitions, func(i, j int) bool {
			return partitions[i].Identifier < partitions[j].Identifier
		})
	}
}

func compareSizeStrings(size1, size2 string) bool {
	bytes1 := parseSizeToBytes(size1)
	bytes2 := parseSizeToBytes(size2)
	return bytes1 < bytes2
}

func parseSizeToBytes(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(sizeStr)
	re := regexp.MustCompile(`([\d\.]+)\s*(\w+)`)
	matches := re.FindStringSubmatch(sizeStr)
	if len(matches) != 3 {
		return 0
	}
	sizeValue, _ := strconv.ParseFloat(matches[1], 64)
	unit := strings.ToUpper(matches[2])

	multiplier := map[string]float64{
		"B":  1,
		"KB": 1 << 10,
		"MB": 1 << 20,
		"GB": 1 << 30,
		"TB": 1 << 40,
	}
	return int64(sizeValue * multiplier[unit])
}

func sizeToBytes(sizeStr string) string {
	bytes := parseSizeToBytes(sizeStr)
	return fmt.Sprintf("%d", bytes)
}

func getFilesystemInfo(disk *Disk) {
	cmd := exec.Command("diskutil", "info", disk.Identifier)
	out, err := cmd.Output()
	if err != nil {
		return
	}
	// Parse the output to extract filesystem info
	info := string(out)
	disk.FSType = getValueForKey(info, "Type (Bundle):")
	disk.Label = getValueForKey(info, "Volume Name:")
	disk.UUID = getValueForKey(info, "Volume UUID:")
	disk.Mountpoint = getValueForKey(info, "Mount Point:")
}

func getPartitionFilesystemInfo(part *Partition) {
	cmd := exec.Command("diskutil", "info", part.Identifier)
	out, err := cmd.Output()
	if err != nil {
		return
	}
	// Parse the output to extract filesystem info
	info := string(out)
	part.FSType = getValueForKey(info, "Type (Bundle):")
	part.Label = getValueForKey(info, "Volume Name:")
	part.UUID = getValueForKey(info, "Volume UUID:")
	part.Mountpoint = getValueForKey(info, "Mount Point:")
}

func getValueForKey(info string, key string) string {
	re := regexp.MustCompile(key + `\s*(.*)`)
	matches := re.FindStringSubmatch(info)
	if len(matches) == 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func includeDisks(disks []Disk, include []string) []Disk {
	var result []Disk
	includeMap := make(map[string]bool)
	for _, name := range include {
		includeMap[name] = true
	}
	for _, disk := range disks {
		if includeMap[disk.Name] {
			result = append(result, disk)
		}
	}
	return result
}

func excludeDisks(disks []Disk, exclude []string) []Disk {
	var result []Disk
	excludeMap := make(map[string]bool)
	for _, name := range exclude {
		excludeMap[name] = true
	}
	for _, disk := range disks {
		if !excludeMap[disk.Name] {
			result = append(result, disk)
		}
	}
	return result
}

func jsonOutput(disks []Disk) {
	genericJSON := make(map[string]interface{})
	for _, disk := range disks {
		diskJSON := make(map[string]interface{})
		if disk.Size != "" {
			diskJSON["size"] = disk.Size
		}
		diskJSON["type"] = disk.Type
		diskJSON["identifier"] = disk.Identifier
		if disk.FSType != "" {
			diskJSON["fstype"] = disk.FSType
		}
		if disk.Label != "" {
			diskJSON["label"] = disk.Label
		}
		if disk.UUID != "" {
			diskJSON["uuid"] = disk.UUID
		}
		if disk.Mountpoint != "" {
			diskJSON["mountpoint"] = disk.Mountpoint
		}
		if disk.PartitionType != "" {
			diskJSON["partitiontype"] = disk.PartitionType
		}

		partitions := make([]map[string]interface{}, len(disk.Partitions))
		for i, part := range disk.Partitions {
			partJSON := make(map[string]interface{})
			partJSON["name"] = part.Name
			partJSON["type"] = part.Type
			if part.Identifier != "" {
				partJSON["identifier"] = part.Identifier
			}
			if part.Size != "" {
				partJSON["size"] = part.Size
			}
			if part.FSType != "" {
				partJSON["fstype"] = part.FSType
			}
			if part.Label != "" {
				partJSON["label"] = part.Label
			}
			if part.UUID != "" {
				partJSON["uuid"] = part.UUID
			}
			if part.Mountpoint != "" {
				partJSON["mountpoint"] = part.Mountpoint
			}
			if part.PartitionType != "" {
				partJSON["partitiontype"] = part.PartitionType
			}
			partitions[i] = partJSON
		}
		diskJSON["partitions"] = partitions
		genericJSON[disk.Name] = diskJSON
	}

	// Remove blank fields
	data, err := json.MarshalIndent(genericJSON, "", "  ")
	if err != nil {
		fmt.Printf("Error generating JSON output: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
