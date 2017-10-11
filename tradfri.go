package main

import (
	"encoding/json"
	"flag"
	"fmt"
	// "github.com/davecgh/go-spew/spew"
	"os"
	"path"
	"time"

	"github.com/vharitonsky/iniflags"
	"github.com/zubairhamed/canopus"
)

// types to unmarshal json data from tradfri
// generated with help from https://mholt.github.io/json-to-go/
// struct names derived from
// - https://github.com/IPSO-Alliance/pub/blob/master/reg/README.md
// - https://github.com/hardillb/TRADFRI2MQTT/blob/master/src/main/java/uk/me/hardill/TRADFRI2MQTT/TradfriConstants.java
// - http://www.openmobilealliance.org/wp/OMNA/LwM2M/LwM2MRegistry.html#resources

type device_ids []int
type group_ids []int

type device_desc struct {
	Device struct {
		Manufacturer          string `json:"0"`
		DeviceDescription     string `json:"1"`
		SerialNumber          string `json:"2"`
		FirmwareVersion       string `json:"3"`
		AvailablePowerSources int    `json:"6"`
	} `json:"3"`
	LightControl []struct {
		Power   int `json:"5850"`
		Dim     int `json:"5851"`
		Num9003 int `json:"9003"`
	} `json:"3311"`
	ApplicationType int    `json:"5750"`
	DeviceName      string `json:"9001"`
	Num9002         int    `json:"9002"`
	DeviceID        int    `json:"9003"`
	Num9019         int    `json:"9019"`
	Num9020         int    `json:"9020"`
	Num9054         int    `json:"9054"`
}

type group_desc struct {
	Power         int    `json:"5850"`
	Dim           int    `json:"5851"`
	GroupName     string `json:"9001"`
	Num9002       int    `json:"9002"`
	GroupID       int    `json:"9003"`
	AccessoryLink struct {
		LinkedItems struct {
			DeviceIDs []int `json:"9003"`
		} `json:"15002"`
	} `json:"9018"`
	Num9039 int `json:"9039"`
}

// type to read the config file

type tradfri_cfg struct {
	Hubip string
	Key   string
}

// flags
var (
	gateway     = flag.String("gateway", "127.0.0.1", "Address of Tradfri gateway.")
	key         = flag.String("key", "deadbeef", "API key to access gateway.")
	status      = flag.Bool("status", false, "Show status of all Tradfri devices.")
	power       = flag.Bool("power", false, "Modify power state of a device or group.")
	dim         = flag.Bool("dim", false, "Dim a device or group.")
	color       = flag.Bool("color", false, "Set color for a device or group.")
	device      = flag.Bool("device", false, "Talk to a device")
	group       = flag.Bool("group", false, "Talk to a group")
	value       = flag.Int("value", -1, "Target value (0-100 for dim, 0 or 1 for power).")
	target_id   = flag.Int("id", -1, "Device or Group ID.")
	target_name = flag.String("name", "", "Device or Group name")
	period      = flag.Int("period", 0,
		"Dim period in seconds. Will dim immediately if set to 0.")
	steps = flag.Int("steps", 10, "Number of intermediate steps for dim action.")
)

// process flags
func init() {
	flag.Usage = usage
	iniflags.SetConfigFile(path.Join(os.Getenv("HOME"), ".tradfri.ini"))
	iniflags.Parse()
}

// usage info
func usage() {
	flag.PrintDefaults()
	os.Exit(1)
}

// deal with errors gracelessly
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func tradfri_conn(address string, key string) canopus.Connection {
	var tradfri_gw = fmt.Sprintf("%s:5684", address)
	conn, err := canopus.DialDTLS(tradfri_gw, "", key)
	check(err)
	return conn
}

func list_group_ids(conn canopus.Connection) group_ids {
	var group_id_list group_ids

	// setup request for device ids
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Get)
	req.SetStringPayload("")
	req.SetRequestURI("/15004")

	// request device ids
	fmt.Print("Looking for groups... ")
	resp, err := conn.Send(req)
	check(err)

	json.Unmarshal([]byte(resp.GetMessage().GetPayload().String()), &group_id_list)
	return group_id_list
}

func list_groups(group_id_list group_ids, conn canopus.Connection) {
	// enumerating group information
	fmt.Println("enumerating:")
	for _, group := range group_id_list {
		group_info(group, conn)
		// sleep for a while to avoid flood protection
		time.Sleep(500 * time.Millisecond)
	}
}

func group_description(group_id int, conn canopus.Connection) group_desc {
	var desc group_desc
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Get)
	req.SetStringPayload("")
	ru := fmt.Sprintf("/15004/%v", group_id)
	req.SetRequestURI(ru)
	dresp, err := conn.Send(req)
	check(err)
	// output basic device information
	json.Unmarshal([]byte(dresp.GetMessage().GetPayload().String()), &desc)
	return desc
}

func group_info(group_id int, conn canopus.Connection) {
	desc := group_description(group_id, conn)
	fmt.Printf("ID: %v, Name: %v\n", desc.GroupID, desc.GroupName)
	fmt.Printf("Power: %v, Dim: %v\n", desc.Power, desc.Dim)
	fmt.Printf("Members: %v\n", desc.AccessoryLink.LinkedItems.DeviceIDs)
}

func group_power(group_id int, conn canopus.Connection) int {
	desc := group_description(group_id, conn)
	return desc.Power
}

func group_dim(group_id int, conn canopus.Connection) int {
	desc := group_description(group_id, conn)
	return desc.Dim
}

func list_device_ids(conn canopus.Connection) device_ids {
	var device_id_list device_ids

	// setup request for device ids
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Get)
	req.SetStringPayload("")
	req.SetRequestURI("/15001")

	// request device ids
	fmt.Print("Looking for devices... ")
	resp, err := conn.Send(req)
	check(err)

	json.Unmarshal([]byte(resp.GetMessage().GetPayload().String()), &device_id_list)
	return device_id_list
}

func list_devices(device_id_list device_ids, conn canopus.Connection) {
	fmt.Println("enumerating:")
	for _, device := range device_id_list {
		device_info(device, conn)

		// sleep for a while to avoid flood protection
		time.Sleep(500 * time.Millisecond)
	}
}

func device_description(device_id int, conn canopus.Connection) device_desc {
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Get)
	req.SetStringPayload("")
	ru := fmt.Sprintf("/15001/%v", device_id)
	req.SetRequestURI(ru)
	dresp, err := conn.Send(req)
	check(err)

	// output basic device information
	var desc device_desc
	json.Unmarshal([]byte(dresp.GetMessage().GetPayload().String()), &desc)
	return desc
}

func device_info(device_id int, conn canopus.Connection) {
	var desc device_desc
	desc = device_description(device_id, conn)
	fmt.Printf("ID: %v, Name; %v, Description: %v\n",
		desc.DeviceID, desc.DeviceName, desc.Device.DeviceDescription)

	// only output light control info if available
	if len(desc.LightControl) > 0 {
		for count, entry := range desc.LightControl {
			fmt.Printf("Light Control Set %v, Power: %v, Dim: %v\n",
				count, entry.Power, entry.Dim)
		}
	} else {
		fmt.Println("No light control values")
	}
}

func device_power(device_id int, conn canopus.Connection) int {
	desc := device_description(device_id, conn)
	// tradfri lamps only have a single light control
	if len(desc.LightControl) > 0 {
		return desc.LightControl[0].Power
	} else {
		panic("No light control info found1")
	}
}

func device_dim(device_id int, conn canopus.Connection) int {
	desc := device_description(device_id, conn)
	// tradfri lamps only have a single light control
	if len(desc.LightControl) > 0 {
		return desc.LightControl[0].Dim
	} else {
		panic("No light control info found1")
	}
}

func power_device(device_id int, val int, conn canopus.Connection) {
	device_info(device_id, conn)
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Put)
	payload := fmt.Sprintf("{ \"3311\" : [{ \"5850\" : %v }] }", val)
	req.SetStringPayload(payload)
	ru := fmt.Sprintf("/15001/%v", device_id)
	req.SetRequestURI(ru)
	_, err := conn.Send(req)
	check(err)
	device_info(device_id, conn)
}

func dim_device(device_id int, val int, conn canopus.Connection) {
	fmt.Printf("Dim level at start: %v, ", device_dim(device_id, conn))
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Put)
	payload := fmt.Sprintf("{ \"3311\" : [{ \"5851\" : %v }] }", val)
	req.SetStringPayload(payload)
	ru := fmt.Sprintf("/15001/%v", device_id)
	req.SetRequestURI(ru)
	_, err := conn.Send(req)
	check(err)
	fmt.Printf("dim level at end: %v\n", device_dim(device_id, conn))
}

func power_group(group_id int, val int, conn canopus.Connection) {
	group_info(group_id, conn)
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Put)
	payload := fmt.Sprintf("{ \"5850\": %d }", val)
	req.SetStringPayload(payload)
	ru := fmt.Sprintf("/15004/%v", group_id)
	req.SetRequestURI(ru)
	_, err := conn.Send(req)
	check(err)
	group_info(group_id, conn)
}

func dim_group(group_id int, val int, conn canopus.Connection) {
	fmt.Printf("Dim level at start: %v, ", group_dim(group_id, conn))
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Put)
	payload := fmt.Sprintf("{ \"5851\": %d }", val)
	req.SetStringPayload(payload)
	ru := fmt.Sprintf("/15004/%v", group_id)
	req.SetRequestURI(ru)
	_, err := conn.Send(req)
	check(err)
	fmt.Printf("dim level at end: %v\n", group_dim(group_id, conn))
}

func validate_flags() {
	if !*status && !*power && !*dim && !*color {
		usage()
	}
	if *device && *group {
		panic("Only one of -device and -group should be set.")
	}
	if *device && *target_id == -1 {
		if len(*target_name) < 1 {
			panic("Need device id or name to run.")
		}
	}
	if *group && *target_id == -1 {
		if len(*target_name) < 1 {
			panic("Need group id or name to run.")
		}
	}
}

func main() {
	validate_flags()
	conn := tradfri_conn(*gateway, *key)
	if *status && *target_id == -1 {
		list_devices(list_device_ids(conn), conn)
		list_groups(list_group_ids(conn), conn)
	}
	if *status && *device {
		device_info(*target_id, conn)
	}
	if *status && *group {
		group_info(*target_id, conn)
	}

	if *power {
		if *device && *value != -1 {
			power_device(*target_id, *value, conn)
		}
		if *device && *value == -1 {
			device_power(*target_id, conn)
		}

		if *group && *value != -1 {
			power_group(*target_id, *value, conn)
		}
		if *group && *value == -1 {
			group_power(*target_id, conn)
		}
	}

	if *dim && *period == 0 {

		if *device && *value != -1 {
			dim_device(*target_id, *value, conn)
		}
		if *device && *value == -1 {
			device_dim(*target_id, conn)
		}

		// if device_dim(*target_id, conn) < 13 {
		//	fmt.Printf("Minimum brightness reached, turning off device.")
		//	power_device(*target_id, 0, conn)
		// }

		if *group && *value != -1 {
			dim_group(*target_id, *value, conn)
		}
		if *group && *value == -1 {
			group_dim(*target_id, conn)
		}

		// if group_dim(*target_id, conn) < 13 {
		//	fmt.Printf("Minimum brightness reached, turning off device.")
		//	power_group(*target_id, 0, conn)
		// }

	}

	if *dim && *period > 0 {
		interval := int(*period / *steps)
		fmt.Printf("dimming in %v %v second intervals, ", *steps, interval)
		if *device {
			current_brightness := device_dim(*target_id, conn)
			difference := int(*value - current_brightness)
			difference_per_interval := int(difference / *steps)
			fmt.Printf("difference per interval %v\n",
				difference_per_interval)
			new_dim := current_brightness
			for current_brightness != *value {
				new_dim += difference_per_interval
				if new_dim < *value && difference_per_interval < 0 {
					new_dim = *value
				}
				if new_dim > *value && difference_per_interval > 0 {
					new_dim = *value
				}
				if new_dim > 12 && device_power(*target_id, conn) == 0 {
					fmt.Printf("Turning up dimmer on device that is powered down, powering up ...")
					power_device(*target_id, 1, conn)
				}
				fmt.Printf(" new dim level %v\n", new_dim)
				dim_device(*target_id, new_dim, conn)
				time.Sleep(time.Duration(interval) * time.Second)
				current_brightness = device_dim(*target_id, conn)
			}
			if current_brightness < 12 {
				fmt.Printf("Minimum brightness reached, turning off device.")
				power_device(*target_id, 0, conn)
			}
		}

		if *group {
			current_brightness := group_dim(*target_id, conn)
			difference := int(*value - current_brightness)
			difference_per_interval := int(difference / *steps)
			fmt.Printf("difference per interval %v\n",
				difference_per_interval)
			new_dim := current_brightness
			for current_brightness != *value {
				new_dim += difference_per_interval
				if new_dim < *value && difference_per_interval < 0 {
					new_dim = *value
				}
				if new_dim > *value && difference_per_interval > 0 {
					new_dim = *value
				}
				if new_dim > 12 && group_power(*target_id, conn) == 0 {
					fmt.Printf("Turning up dimmer on group that is powered down, powering up ...")
					power_group(*target_id, 1, conn)
				}
				fmt.Printf(" new dim level %v\n", new_dim)
				dim_group(*target_id, new_dim, conn)
				time.Sleep(time.Duration(interval) * time.Second)
				current_brightness = group_dim(*target_id, conn)
			}
			if current_brightness < 12 {
				fmt.Printf("Minimum brightness reached, turning off group.")
				power_group(*target_id, 0, conn)
			}
		}
	}

	if *color {
		// TODO(hwa): Add color support when we have lamps that can change color.
		fmt.Printf("set color to value %v on id %v\n",
			*value, *target_id)
	}
}
