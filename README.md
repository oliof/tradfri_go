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

## Known issues

 *   When using the `period` and `steps` flags, results are imprecise:  
   * Because we may lose messages due to flood protection on the gateway, the
     loop doesn't always end up at exactly n steps.
   * Because the way we compute the interval to dim by is imprecise, the loop
     doesn't always end up at exactly n steps.
 *   Running two instances ofg `tradfri`, one dimming down over a `period`, the
     other dimming up, will result in both running potentially for-ever.
 *   There is no protection against other clients (i.e. remotes) modifying the
     devices or group, which can also lead to changes in runtime or results on
     long-running dim actions that are rather unexpected
 *   There are no sanity checks on any of the flags (other than checking for
     non-sensical combinations). Proceed at your own risk!
 
