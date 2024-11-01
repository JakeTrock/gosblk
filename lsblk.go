package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/tabwriter"
)

type Partition struct {
	Name       string
	Type       string
	Identifier string
	Size       string
}

type Disk struct {
	Name       string
	Size       string
	Type       string
	Identifier string
	Partitions []Partition
}

func main() {
	cmd := exec.Command("diskutil", "list")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error running diskutil list: %v\n", err)
		return
	}

	disks := parseDiskutilOutput(out.String())

	// Prepare to print in table format
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSIZE\tTYPE\tIDENTIFIER")

	for _, disk := range disks {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", disk.Name, disk.Size, disk.Type, disk.Identifier)
		for ind, part := range disk.Partitions {
			if ind == len(disk.Partitions)-1 {
				fmt.Fprintf(w, "└─%s\t%s\t%s\t%s\n", part.Name, part.Size, part.Type, part.Identifier)
				continue
			}
			fmt.Fprintf(w, "├─%s\t%s\t%s\t%s\n", part.Name, part.Size, part.Type, part.Identifier)
		}
	}

	w.Flush()
}

func parseDiskutilOutput(output string) []Disk {
	lines := strings.Split(output, "\n")
	var disks []Disk
	var currentDisk *Disk
	diskHeaderRegex := regexp.MustCompile(`^\/dev\/(disk\d+)`)
	diskInfoRegex := regexp.MustCompile(`^\/dev\/(disk\d+).*?\*\s*([\d\.]+\s\w+).*`)
	partitionRegex := regexp.MustCompile(`^\s+(\d+):\s+(\S+)\s+(.*?)\s+([\d\.]+\s\w+)\s+(\S+)$`)

	for _, line := range lines {
		if diskInfoRegex.MatchString(line) {
			matches := diskInfoRegex.FindStringSubmatch(line)
			currentDisk = &Disk{
				Name:       matches[1],
				Size:       matches[2],
				Type:       "disk",
				Identifier: matches[1],
			}
			disks = append(disks, *currentDisk)
		} else if partitionRegex.MatchString(line) && currentDisk != nil {
			matches := partitionRegex.FindStringSubmatch(line)
			partition := Partition{
				Name:       matches[3],
				Type:       matches[2],
				Size:       matches[4],
				Identifier: matches[5],
			}
			// Add partition to the last disk in the slice
			diskIndex := len(disks) - 1
			disks[diskIndex].Partitions = append(disks[diskIndex].Partitions, partition)
		} else if diskHeaderRegex.MatchString(line) {
			matches := diskHeaderRegex.FindStringSubmatch(line)
			currentDisk = &Disk{
				Name:       matches[1],
				Type:       "disk",
				Identifier: matches[1],
			}
			disks = append(disks, *currentDisk)
		}
	}

	return disks
}
