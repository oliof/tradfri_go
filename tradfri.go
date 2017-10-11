package main

import (
	"encoding/json"
	"flag"
	"fmt"
	// "github.com/davecgh/go-spew/spew"
	"github.com/vharitonsky/iniflags"
	"github.com/zubairhamed/canopus"
	"os"
	"time"
)

// types to unmarshal json data from tradfri
// partially generated with https://mholt.github.io/json-to-go/
// struct names derived from https://github.com/IPSO-Alliance/pub/blob/master/reg/README.md
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
	Power     int    `json:"5850"`
	Dim       int    `json:"5851"`
	GroupName string `json:"9001"`
	Num9002   int    `json:"9002"`
	GroupID   int    `json:"9003"`
	Num9018   struct {
		Num15002 struct {
			Num9003 []int `json:"9003"`
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
	action      = flag.String("action", "status", "action to take [dim|status|power]).")
	target      = flag.Int("target", 0, "Target value (0-100 for dim, 0 or 1 for power).")
	target_id   = flag.Int("id", 65537, "Device or Group ID.")
	target_name = flag.String("name", "", "Device or Group name")
	steps       = flag.Int("steps", 10, "Number of intermediate steps for dim action.")
	period      = flag.Int("period", 60, "Time period in seconds to run dim action over.")
)

// process flags
func init() {
	flag.Usage = usage
	iniflags.SetConfigFile("tradfri.ini")
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
	fmt.Println("Connecting to tradfri gateway... ")
	conn, err := canopus.DialDTLS(tradfri_gw, "", key)
	check(err)
	fmt.Println("connected")
	return conn
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

func get_group_info(group_id int, conn canopus.Connection) {
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Get)
	req.SetStringPayload("")
	ru := fmt.Sprintf("/15004/%v", group_id)
	req.SetRequestURI(ru)
	dresp, err := conn.Send(req)
	check(err)

	// output basic device information
	var desc group_desc
	json.Unmarshal([]byte(dresp.GetMessage().GetPayload().String()), &desc)
	fmt.Printf("ID: %v, Name: %v\n", desc.GroupID, desc.GroupName)
	fmt.Printf("Power: %v, Dim: %v\n", desc.Power, desc.Dim)

}

func list_groups(group_id_list group_ids, conn canopus.Connection) {
	// enumerating group information
	fmt.Println("enumerating:")
	for _, group := range group_id_list {
		get_group_info(group, conn)
		// sleep for a while to avoid flood protection
		time.Sleep(500 * time.Millisecond)
	}
}

func get_device_info(device_id int, conn canopus.Connection) {
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Get)
	req.SetStringPayload("")
	ru := fmt.Sprintf("/15001/%v", device_id)
	req.SetRequestURI(ru)
	dresp, err := conn.Send(req)
	check(err)

	// output basic device information
	var desc device_desc
	json.Unmarshal([]byte(dresp.GetMessage().GetPayload().String()), &desc)
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

func list_devices(device_id_list device_ids, conn canopus.Connection) {
	fmt.Println("enumerating:")
	for _, device := range device_id_list {
		get_device_info(device, conn)

		// sleep for a while to avoid flood protection
		time.Sleep(500 * time.Millisecond)
	}
}

func power_device(device_id int, val int, conn canopus.Connection) {
	get_device_info(device_id, conn)
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Put)
	payload := fmt.Sprintf("{ \"3311\" : [{ \"5850\" : %v }] }", val)
	req.SetStringPayload(payload)
	ru := fmt.Sprintf("/15001/%v", device_id)
	req.SetRequestURI(ru)
	_, err := conn.Send(req)
	check(err)
	get_device_info(device_id, conn)
}

func dim_device(device_id int, val int, conn canopus.Connection) {
	get_device_info(device_id, conn)
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Put)
	payload := fmt.Sprintf("{ \"3311\" : [{ \"5851\" : %v }] }", val)
	req.SetStringPayload(payload)
	ru := fmt.Sprintf("/15001/%v", device_id)
	req.SetRequestURI(ru)
	_, err := conn.Send(req)
	check(err)
	get_device_info(device_id, conn)
}

func power_group(group_id int, val int, conn canopus.Connection) {
	get_group_info(group_id, conn)
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Put)
	payload := fmt.Sprintf("{ \"5850\": %d }", val)
	req.SetStringPayload(payload)
	ru := fmt.Sprintf("/15004/%v", group_id)
	req.SetRequestURI(ru)
	_, err := conn.Send(req)
	check(err)
	get_group_info(group_id, conn)
}

func dim_group(group_id int, val int, conn canopus.Connection) {
	get_group_info(group_id, conn)
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Put)
	payload := fmt.Sprintf("{ \"5851\": %d }", val)
	req.SetStringPayload(payload)
	ru := fmt.Sprintf("/15004/%v", group_id)
	req.SetRequestURI(ru)
	_, err := conn.Send(req)
	check(err)
	get_group_info(group_id, conn)
}

func main() {
	conn := tradfri_conn(*gateway, *key)
	if *action == "status" {
		list_devices(list_device_ids(conn), conn)
		list_groups(list_group_ids(conn), conn)
	}
	if *action == "power" {
		fmt.Printf("power target %v on id %v", *target, *target_id)
	}
	if *action == "dim" {
		fmt.Printf("dim target %v on id %v", *target, *target_id)
	}
}
