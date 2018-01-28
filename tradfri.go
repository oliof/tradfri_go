package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"time"

        "github.com/davecgh/go-spew/spew"
	"github.com/vharitonsky/iniflags"
	"github.com/zubairhamed/canopus"

)

// process flags
func init() {
	flag.Usage = usage
	inifile := path.Join(os.Getenv("HOME"), ".tradfri.ini")
	iniflags.SetConfigFile(inifile)
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

func group_get_description(group_id int, conn canopus.Connection) group_desc {
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
	desc := group_get_description(group_id, conn)
	fmt.Printf("ID: %v, Name: %v\n", desc.GroupID, desc.GroupName)
	fmt.Printf("Power: %v, Dim: %v\n", desc.Power, desc.Dim)
	fmt.Printf("Members: %v\n", desc.AccessoryLink.LinkedItems.DeviceIDs)
}

func group_get_power(group_id int, conn canopus.Connection) int {
	desc := group_get_description(group_id, conn)
	return desc.Power
}

func group_get_dim(group_id int, conn canopus.Connection) int {
	desc := group_get_description(group_id, conn)
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

func device_get_description(device_id int, conn canopus.Connection) device_desc {
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
	desc = device_get_description(device_id, conn)
	fmt.Printf("ID: %v, Name; %v, Description: %v, Type: %v\n",
		desc.DeviceID, desc.DeviceName, desc.Device.DeviceDescription, desc.ApplicationType)
        if desc.ApplicationType == Remote {
                fmt.Printf( "Battery Level: %v\n", desc.Device.BatteryLevel)
        }
        if desc.ApplicationType == Lamp {
		for count, entry := range desc.LightControl {
			fmt.Printf("Light Control Set %v, Power: %v, Dim: %v, 9003: %v\n",
				count, entry.Power, entry.Dim, entry.Num9003)
		}
	}
        if *verbose {
               spew.Dump(desc)
        }
}

func device_get_power(device_id int, conn canopus.Connection) int {
	desc := device_get_description(device_id, conn)
	// tradfri lamps only have a single light control
	if len(desc.LightControl) > 0 {
		return desc.LightControl[0].Power
	} else {
		panic("No light control info found!")
	}
}

func device_get_dim(device_id int, conn canopus.Connection) int {
	desc := device_get_description(device_id, conn)
	// tradfri lamps only have a single light control
	if len(desc.LightControl) > 0 {
		return desc.LightControl[0].Dim
	} else {
		panic("No light control info found!")
	}
}

func device_set_power(device_id int, val int, conn canopus.Connection) {
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

func device_set_dim(device_id int, val int, conn canopus.Connection) {
	fmt.Printf("Dim level at start: %v, ", device_get_dim(device_id, conn))
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Put)
	payload := fmt.Sprintf("{ \"3311\" : [{ \"5851\" : %v }] }", val)
	req.SetStringPayload(payload)
	ru := fmt.Sprintf("/15001/%v", device_id)
	req.SetRequestURI(ru)
	_, err := conn.Send(req)
	check(err)
	fmt.Printf("dim level at end: %v\n", device_get_dim(device_id, conn))
}

func group_set_power(group_id int, val int, conn canopus.Connection) {
	group_get_power(group_id, conn)
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Put)
	payload := fmt.Sprintf("{ \"5850\": %d }", val)
	req.SetStringPayload(payload)
	ru := fmt.Sprintf("/15004/%v", group_id)
	req.SetRequestURI(ru)
	_, err := conn.Send(req)
	check(err)
	group_get_power(group_id, conn)
}

func group_set_dim(group_id int, val int, conn canopus.Connection) {
	fmt.Printf("Dim level at start: %v, ", group_get_dim(group_id, conn))
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Put)
	payload := fmt.Sprintf("{ \"5851\": %d }", val)
	req.SetStringPayload(payload)
	ru := fmt.Sprintf("/15004/%v", group_id)
	req.SetRequestURI(ru)
	_, err := conn.Send(req)
	check(err)
	fmt.Printf("dim level at end: %v\n", group_get_dim(group_id, conn))
}

func validate_flags() {
	if !*status && !*power && !*dim {
		usage()
	}
	if *device && *group {
		fmt.Println("Only one of -device and -group should be set.")
		usage()
	}
	if *device && *target_id == -1 {
		fmt.Println("Need device id to run.")
		usage()
	}
	if *group && *target_id == -1 {
		fmt.Println("Need group id or name to run.")
		usage()
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
			device_set_power(*target_id, *value, conn)
		}
		if *device && *value == -1 {
			device_get_power(*target_id, conn)
		}

		if *group && *value != -1 {
			group_set_power(*target_id, *value, conn)
		}
		if *group && *value == -1 {
			group_get_power(*target_id, conn)
		}
	}

	if *dim && *period == 0 {

		if *device && *value != -1 {
			device_set_dim(*target_id, *value, conn)
		}
		if *device && *value == -1 {
			device_get_dim(*target_id, conn)
		}

		// if device_get_dim(*target_id, conn) < 13 {
		//	fmt.Printf("Minimum brightness reached, turning off device.")
		//	device_set_power(*target_id, 0, conn)
		// }

		if *group && *value != -1 {
			group_set_dim(*target_id, *value, conn)
		}
		if *group && *value == -1 {
			group_get_dim(*target_id, conn)
		}

		// if group_get_dim(*target_id, conn) < 13 {
		//	fmt.Printf("Minimum brightness reached, turning off device.")
		//	group_set_power(*target_id, 0, conn)
		// }

	}

	if *dim && *period > 0 {
		interval := int(*period / *steps)
		fmt.Printf("dimming in %v %v second intervals, ", *steps, interval)
		if *device {
			current_brightness := device_get_dim(*target_id, conn)
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
				if new_dim > 12 && device_get_power(*target_id, conn) == 0 {
					fmt.Println(
						"Turning up dimmer on device that is powered down, powering up ...")
					device_set_power(*target_id, 1, conn)
				}
				fmt.Printf(" new dim level %v\n", new_dim)
				device_set_dim(*target_id, new_dim, conn)
				time.Sleep(time.Duration(interval) * time.Second)
				current_brightness = device_get_dim(*target_id, conn)
			}
			if current_brightness < 12 {
				fmt.Println("Minimum brightness reached, turning off device.")
				device_set_power(*target_id, 0, conn)
			}
		}

		if *group {
			current_brightness := group_get_dim(*target_id, conn)
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
				if new_dim > 12 && group_get_power(*target_id, conn) == 0 {
					fmt.Println("Turning up dimmer on group that is powered down, powering up ...")
					group_set_power(*target_id, 1, conn)
				}
				fmt.Printf(" new dim level %v\n", new_dim)
				group_set_dim(*target_id, new_dim, conn)
				time.Sleep(time.Duration(interval) * time.Second)
				current_brightness = group_get_dim(*target_id, conn)
			}
			if current_brightness < 12 {
				fmt.Println("Minimum brightness reached, turning off group.")
				group_set_power(*target_id, 0, conn)
			}
		}
	}
}
