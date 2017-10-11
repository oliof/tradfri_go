# tradfri.go 

### tradfri.go is a golang implementation of a coaps client to control IKEA Tradfri lamps.

# Usage

Flags can be passed on the command line or in an ini file. The default file name is `tradfri.ini`

## Flags

### Mandatory Flags

 * `-gateway` is the address of the Tradfri gateway
 * `-key` is the API key that is printed on the bottom of your gateway

### Simple Flags

 * `-status` finds and lists all Tradfri objects: Remotes, lamps, groups.
 * `-h` prints usage information.
 
## Examples

### Get all info

```bash
$ cat tradfri.ini
gateway = 192.168.172.2
key = abcdefgh12345678
$ ./tradfri -status
Looking for devices... enumerating:
ID: 65536, Name; Remote, Description: TRADFRI remote control
No light control values
ID: 65538, Name; Lamp1, Description: TRADFRI bulb E27 opal 1000lm
Light Control Set 0, Power: 1, Dim: 254
ID: 65537, Name; Lamp2, Description: TRADFRI bulb E27 opal 1000lm
Light Control Set 0, Power: 1, Dim: 254
Looking for groups... enumerating:
ID: 135490, Name: MyGroup
Power: 1, Dim: 255
``` 

### List a group's information

```bash
$ ./tradfri -status -device -id 65538
ID: 65538, Name; Lamp1, Description: TRADFRI bulb E27 opal 1000lm
Light Control Set 0, Power: 1, Dim: 254
./tradfri -status -group -id 135490
ID: 135490, Name: MyGroup
Power: 1, Dim: 255
```

### Dim a lamp to 50% brightness

```bash
$ ./tradfri -dim -device -id 65537 -value 128
Dim level at start: 202, dim level at end: 128
```

### Dim a lamp to 80ish% brightness in 5 steps over 10 seconds
```bash
$ ./tradfri -dim -device -id 65537 -value 224 -steps 5 -period 10
dimming in 5 2 second intervals, difference per interval 19
 new dim level 147
Dim level at start: 128, dim level at end: 147
 new dim level 166
Dim level at start: 147, dim level at end: 166
 new dim level 185
Dim level at start: 166, dim level at end: 185
 new dim level 204
Dim level at start: 185, dim level at end: 204
 new dim level 223
Dim level at start: 204, dim level at end: 223
 new dim level 224
Dim level at start: 223, dim level at end: 224
```





allowMissingConfig = false  # Don't terminate the app if the ini file cannot be read.
allowUnknownFlags = false  # Don't terminate the app if ini file contains unknown flags.
configUpdateInterval = 0s  # Update interval for re-reading config file set via -config flag. Zero disables config file re-reading.
color = false  # Set color for a device or group. 
device = false  # Talk to a device
dim = false  # Dim a device or group.
gateway = 127.0.0.1  # Address of Tradfri gateway.
group = false  # Talk to a group
id = -1  # Device or Group ID.
key = deadbeef  # API key to access gateway.
name =   # Device or Group name
period = 0  # Dim period in seconds. Will dim immediately if set to 0.
power = false  # Modify power state of a device or group.
status = false  # Show status of all Tradfri devices.
steps = 10  # Number of intermediate steps for dim action.
value = -1  # Target value (0-100 for dim, 0 or 1 for power).
