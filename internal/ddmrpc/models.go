package ddmrpc

// DEL41D9,Dell S2721DGF,FDSJ093,2021 ISO week 37,1110
type AssetAttributes struct {
	ModelCode    string
	Model        string
	ServiceTag   string
	Manufactured string
	ActiveHours  int64
}

// (prot(monitor)type(lcd)model(s2721dgfa)cmds(01 02 03 07 0c e3 f3)vcp(02 04 05 08 10 12 14(05 08 0b 0c) 16 18 1a 52 60(0f 11 12 ) 62 ac ae b2 b6 c6 c8 c9 ca cc(02 0a 03 04 08 09 0d 06 ) d6(01 04 05) dc(00 05 ) df e0 e1 e2(00 20 21 22 2f 04 1e 1f 1d 0e 12 14 27 23 24 3a ) e3 ea(fe00 fe01) f0(0d 0e 0c 0f 10 11 13 31 32 34 36 ) f1 f2 fd)mswhql(1)asset_eep(40)mccs_ver(2.1)) (cached)
type Capabilities struct {
	AvailableInputs []byte
}

var KnownInputs = map[byte]string{
	0x01: "VGA-1",
	0x02: "VGA-2",
	0x03: "DVI-1",
	0x04: "DVI-2",
	0x05: "Composite video 1",
	0x06: "Composite video 2",
	0x07: "S-Video-1",
	0x08: "S-Video-2",
	0x09: "Tuner-1",
	0x0a: "Tuner-2",
	0x0b: "Tuner-3",
	0x0c: "Component video (YPrPb/YCrCb) 1",
	0x0d: "Component video (YPrPb/YCrCb) 2",
	0x0e: "Component video (YPrPb/YCrCb) 3",
	0x0f: "DisplayPort-1",
	0x10: "DisplayPort-2",
	0x11: "HDMI-1",
	0x12: "HDMI-2",
}
