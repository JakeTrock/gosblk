lsblk
List block devices.

Syntax
lsblk [options] [device...]

Options:
-a, --all
Also list empty devices. (By default they are skipped.)

-b, --bytes
Print the SIZE column in bytes rather than in a human-readable format.

-D, --discard
Print information about the discarding capabilities (TRIM, UNMAP) for each device.

-d, --nodeps
Do not print holder devices or replicas. For example, lsblk
--nodeps /dev/sda prints information about the sda device only.

-e, --exclude list
Exclude the devices specified by the comma-separated list of major device
numbers. Note that RAM disks (major=1) are excluded by default.
The filter is applied to the top-level devices only.

-f, --fs
Output info about filesystems. This option is equivalent to
-o NAME,FSTYPE,LABEL,UUID,MOUNTPOINT. The authoritative
information about filesystems and raids is provided by the blkid(8) command.

-h, --help
Display help text and exit.

-I, --include list
Include devices specified by the comma-separated list of major
device numbers. The filter is applied to the top-level devices only.

-i, --ascii
Use ASCII characters for tree formatting.

-J, --json
Use JSON output format.

-l, --list
Produce output in the form of a list.

-m, --perms
Output info about device owner, group and mode. This option is
equivalent to -o NAME,SIZE,OWNER,GROUP,MODE.

-n, --noheadings
Do not print a header line.

-o, --output list
The output columns to print.
Use --help to get a list of all supported columns.

       The default list of columns may be extended if list is
       specified in the format +list (e.g. lsblk -o +UUID).

-O, --output-all
Output all available columns.

-P, --pairs
Produce output in the form of key="value" pairs. All
potentially unsafe characters are hex-escaped (\xcode).

-p, --paths
Print full device paths.

-r, --raw
Produce output in raw format. All potentially unsafe characters are
hex-escaped (\xcode) in the NAME, KNAME, LABEL, PARTLABEL and MOUNTPOINT columns.

-S, --scsi
Output info about SCSI devices only. All partitions, replicas and holder devices are ignored.

-s, --inverse
Print dependencies in inverse order.

-t, --topology
Output info about block-device topology.
This option is equivalent to -o NAME,ALIGNMENT,MIN-IO,OPT-IO,PHY-SEC,LOG-SEC,ROTA,SCHED,RQ-SIZE,RA,WSAME.

-V, --version
Display version information and exit.

-x, --sort column
Sort output lines by column.
lsblk lists information about all available or the specified block devices. The lsblk command reads the sysfs filesystem and udev db to gather information. The command prints all block devices (except RAM disks) in a tree-like format by default.

Use lsblk --help to get a list of all available columns.

The default output, as well as the default output from options like --fs and --topology, is subject to change. So whenever possible, you should avoid using default outputs in your scripts. Always explicitly define expected columns by using --output columns-list in environments where a stable output is required.

Note that lsblk might be executed in time when udev does not have all information about recently added or modified devices yet. In this case it is recommended to use udevadm settle before lsblk to synchronize with udev.

For partitions, some information (e.g. queue attributes) is inherited from the parent device. The lsblk command needs to be able to look up each block device by major:minor numbers, which is done by using /sys/dev/block.

The lsblk command is part of the util-linux package.

Return Codes
0 Success
1 Failure
32 Not found all specified devices
64 Some specified devices found, some not found.

Examples
List all block devices in a tree-like format:

$ lsblk

List all devices including empty ones:

$ lsblk -a

List the device owner, group and mode:

$ lsblk -m

List the size in bytes of the hard drive sda:

$ lsblk --bytes /dev/sda
