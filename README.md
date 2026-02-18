## Dell Display Manager integration to Home Assistant via MQTT

### Features
- brightness/contrast
- presets (like 20%, 40%, 50%, 100%)
- active hours sensor (just for fun)
- power on/off
- reset
- input selection (after switching the input DDM connection may be lost - it is technical limitation of HDMI (displayPort probably works fine))

### Multiple monitors (beta)
Multiple monitors support is experimental
- Hot reload is supported (monitors connection/disconnection doesnt require restarting the app), but rescan takes some time (~1 min) and during rescan even connected monitors may temporarily be reported as unavailable in mqtt 

### TODO
- presets (as in DDM ui)

### Setup
- Configure
  - copy config.example.json into config.json
  - fill `registry_user` with current user SID (run `whoami /user` in command line)
  - fill `broker_addr` with mqtt broker address
- Install as windows service
  - I suggest [NSSM](https://nssm.cc/)
  - or if you are brave enough [sc.exe](https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/sc-create) is a good option

### Preview
![image](https://github.com/user-attachments/assets/8dc3baf5-b736-4ce0-9f90-3a963ea4e868)

