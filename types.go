package main

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
                BatteryLevel          int    `json:"9"`
	} `json:"3"`
	LightControl []struct {
                Color   int `json:"5706"` 
                ColorX  int `json:"5709"`
                ColorY  int `json:"5710"`
		Power   int `json:"5850"`
		Dim     int `json:"5851"`
		Num9003 int `json:"9003"`
	} `json:"3311"`
	ApplicationType    int    `json:"5750"`
	DeviceName         string `json:"9001"`
	CreatedAt          int    `json:"9002"`
	DeviceID           int    `json:"9003"`
	Reachability_state int    `json:"9019"`
	LastSeen           int    `json:"9020"`
	OTAUpdateState     int    `json:"9054"`
}

type group_desc struct {
	Power         int    `json:"5850"`
	Dim           int    `json:"5851"`
	GroupName     string `json:"9001"`
	CreatedAt     int    `json:"9002"`
	GroupID       int    `json:"9003"`
	AccessoryLink struct {
		LinkedItems struct {
			DeviceIDs []int `json:"9003"`
		} `json:"15002"`
	} `json:"9018"`
	Num9039 int `json:"9039"`
}

