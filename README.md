# Go-SIET

This is Cisco Smart Install Exploitation tool written in golang, inspired by [SIETpy3](https://github.com/Sab0tag3d/SIETpy3)

## Modes
you must specify a mode to run this program:

  **test**:  Tests if devices are vulnerable to Smart Install exploitation

  **get-config**: Retrieves configuration files from vulnerable devices

  **change-config**: Installs new configuration files

  **execute**: Executes commands on devices

you can use **help** command with modes to print help messages for modes. for example:
```
  ./Go-SIET test --help
```

## Usage
```
Usage:
  ./Go-SIET [command]

Available Commands:
  change-config Change device configurations
  execute       Execute commands on vulnerable devices
  get-config    Retrieve device configurations
  help          Help about any command
  test          Test devices for Smart Install vulnerabilities

Flags:
  -h, --help               help for Go-SIET
      --ip string          Target IP address
  -l, --list string        File containing list of target IPs
      --password string    Password for authentication (default "cisco")
      --public-ip string   Public IP address (for public targets, suitable for NAT scenarios)
  -u, --username string    Username for authentication (default "cisco")
      --reload-time        Device reload time in HH:MM format (default is 00:01)
      --reload-time        Device reload time in HH:MM format (default is 00:01)

```

Example:
```
  ./Go-SIET test --ip [ip address]
```
