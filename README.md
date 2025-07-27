# Go-SIET

This is Cisco Smart Install Exploitation tool written in golang, inspired by [SIETpy3](https://github.com/Sab0tag3d/SIETpy3)

## Modes
you must specify a mode to run this program:

-**m test**:  Tests if devices are vulnerable to Smart Install exploitation

-**m get_config**: Retrieves configuration files from vulnerable devices

-**m change_config**: Installs new configuration files

-**m execute**: Executes commands on devices

## Usage
```
-c string
        Configuration file path
  -i string
        Target IP address
  -l string
        File with list of target IPs
  -m string
        Mode: test, get_config, change_config, execute
  -p string
        Public IP address (for public targets)
  -pass string
        Password for default config (default "cisco")
  -r string
        Reload time (HH:MM) (default "00:01")
  -u string
        Username for default config (default "cisco")
```
