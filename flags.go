package main

import "flag"

var (
	gateway = flag.String("gateway", "127.0.0.1",
		"Address of Tradfri gateway.")
	key = flag.String("key", "deadbeef",
		"API key to access gateway.")
	status = flag.Bool("status", false,
		"Show status of all Tradfri devices.")
	power = flag.Bool("power", false,
		"Modify power state of a device or group.")
	dim = flag.Bool("dim", false,
		"Dim a device or group.")
	device = flag.Bool("device", false,
		"Use a device")
	group = flag.Bool("group", false,
		"Use a group")
	value = flag.Int("value", -1,
		"Target value (0-100 for dim, 0 or 1 for power).")
	target_id = flag.Int("id", -1,
		"Device or Group ID.")
	period = flag.Int("period", 0,
		"Dim period in seconds. Will dim immediately if set to 0.")
	steps = flag.Int("steps",
		10, "Number of intermediate steps for dim action.")
)
