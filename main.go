package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

const (
	HOME_ENV     = "HOME"
	CFG_FILENAME = ".wwconfig"
	DEFAULT_PORT = "22"
	DEFAULT_IDEN = "id_rsa"
)

type Config struct {
	Name     string `json:"name"`
	Domain   string `json:"domain"`
	Username string `json:"username"`
	Identity string `json:"identity"`
	Port     string `json:"port"`
}

func main() {
	homePath, exists := os.LookupEnv(HOME_ENV)
	if !exists {
		color.Red("Could not find environment variable HOME. Cannot proceed with setup. Exiting...")
		os.Exit(1)
	}

	chdirErr := os.Chdir(homePath)
	if chdirErr != nil {
		color.Red("A problem occurred when trying to access the config file. Exiting...")
		os.Exit(1)
	}

	configFile, err := os.OpenFile(CFG_FILENAME, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		color.Red("Error opening the config file. Exiting...")
		os.Exit(1)
	}
	configFile.Close()

	data, err := ioutil.ReadFile(CFG_FILENAME)
	if err != nil {
		color.Red("Could not find config file. Exiting...")
		os.Exit(1)
	}

	var configs []Config
	json.Unmarshal(data, &configs)

	if len(os.Args) == 1 {
		color.Yellow("Wadsworth: Your friendly neighborhood SSH butler.\nFor information on how to use Wadsworth, type `ww help`")
		os.Exit(1)
	}

	args := os.Args[1:]

	if args[0] == "add" {
		if len(args) < 3 {
			color.Red("Invalid operation.\n\tFormat: ww add <name> <username>@<domain>[<port>] [optional: <identity_file>]\n\tType ww help for more information...")
			os.Exit(1)
		}

		var newEntry Config
		for _, conf := range configs {
			if args[1] == conf.Name {
				color.Red("Name %s is already in use", conf.Name)
				os.Exit(1)
			}
		}

		domainAfterSplit := strings.Split(args[2], "@")
		if len(domainAfterSplit) == 1 {
			color.Red("Enter a valid domain in the format <username>@<domain>[<port>]")
			os.Exit(1)
		}

		newEntry.Name = args[1]
		newEntry.Username = domainAfterSplit[0]

		if strings.Contains(domainAfterSplit[1], ":") {
			spl := strings.Split(domainAfterSplit[1], ":")
			newEntry.Domain = spl[0]
			newEntry.Port = spl[1]
		} else {
			newEntry.Domain = domainAfterSplit[1]
			newEntry.Port = DEFAULT_PORT
		}

		if len(args) == 4 {
			newEntry.Identity = args[3]
		} else {
			newEntry.Identity = DEFAULT_IDEN
		}

		configs = append(configs, newEntry)

		confs, err := json.Marshal(configs)
		if err != nil {
			color.Red("Error marshalling JSON object. Exiting...")
			os.Exit(1)
		}

		err = ioutil.WriteFile(CFG_FILENAME, confs, 0755)
		if err != nil {
			color.Red("Error writing to file. Exiting...")
			os.Exit(1)
		}
	} else if args[0] == "remove" {
		if len(args) != 2 {
			color.Red("Invalid operation.\n\tFormat: ww remove <name>\n\tType ww help for more information...")
			os.Exit(1)
		}

		for idx, conf := range configs {
			if args[1] == conf.Name {
				configs = append(configs[:idx], configs[idx+1:]...)
			}
		}

		confs, err := json.Marshal(configs)
		if err != nil {
			color.Red("Error marshalling JSON object. Exiting...")
			os.Exit(1)
		}

		err = ioutil.WriteFile(CFG_FILENAME, confs, 0755)
		if err != nil {
			color.Red("Error writing to file. Exiting...")
			os.Exit(1)
		}
	} else if args[0] == "edit" {
		// TODO
		if len(args) != 3 {
			color.Red("Invalid operation.\n\tFormat: ww edit <name> <new_username>@<new_domain>:[<port>]\n\tType ww help for more information...")
			os.Exit(1)
		}

		newDomainAfterSplit := strings.Split(args[2], "@")
		if len(newDomainAfterSplit) == 1 {
			color.Red("Enter a valid domain in the format <username>@<domain>:[<port>]")
			os.Exit(1)
		}

		for idx, _ := range configs {
			if args[1] == configs[idx].Name {
				configs[idx].Username = newDomainAfterSplit[0]

				if strings.Contains(newDomainAfterSplit[1], ":") {
					spl := strings.Split(newDomainAfterSplit[1], ":")
					configs[idx].Domain = spl[0]
					configs[idx].Port = spl[1]
				} else {
					configs[idx].Domain = newDomainAfterSplit[1]
					configs[idx].Port = DEFAULT_PORT
				}
			}
		}

		confs, err := json.Marshal(configs)
		if err != nil {
			color.Red("Error marshalling JSON object. Exiting...")
			os.Exit(1)
		}

		err = ioutil.WriteFile(CFG_FILENAME, confs, 0755)
		if err != nil {
			color.Red("Error writing to file. Exiting...")
			os.Exit(1)
		}
	} else if args[0] == "help" {
		color.Yellow("Say hello to Wadsworth, your friendly neighborhood SSH butler.")
		color.Set(color.FgYellow, color.Bold)
		fmt.Println("\tTo use Wadsworth, type: ww <command> [<extra_arguments>]\n")
		color.Unset()
		color.Red("\tThe list of available commands are ([] indicates optional parameters):\n")
		color.Set(color.FgGreen, color.Bold)
		fmt.Print("\tww add <short_name> <username>@<domain>[<port>] [<identity_file>]: ")
		color.Unset()
		fmt.Println("Adds a new entry for quick access")

		color.Set(color.FgGreen, color.Bold)
		fmt.Print("\tww remove <short_name>: ")
		color.Unset()
		fmt.Println("Removes entry with name <short_name>")

		color.Set(color.FgGreen, color.Bold)
		fmt.Print("\tww edit <short_name> <new_username>@<new_domain>[<port>]: ")
		color.Unset()
		fmt.Println("Edits <short_name> with new domain and username")

		color.Set(color.FgGreen, color.Bold)
		fmt.Print("\tww <short_name>: ")
		color.Unset()
		fmt.Println("Launches SSH process with configuration associated with <short_name>")

		color.Set(color.FgGreen, color.Bold)
		fmt.Print("\tww ls [<short_name>]: ")
		color.Unset()
		fmt.Println("Lists either all or one particular configuration")
	} else if args[0] == "ls" {
		if len(args) == 2 {
			for _, conf := range configs {
				if args[1] == conf.Name {
					color.Set(color.FgYellow, color.Bold)
					fmt.Println("[" + conf.Name + "]")
					color.Unset()
					color.Set(color.FgGreen)
					fmt.Print("\tUsername: ")
					color.Unset()
					fmt.Println("\t" + conf.Username)
					color.Set(color.FgGreen)
					fmt.Print("\tDomain: ")
					color.Unset()
					fmt.Println("\t" + conf.Domain)
					if conf.Port != DEFAULT_PORT {
						color.Set(color.FgGreen)
						fmt.Print("\tPort: ")
						color.Unset()
						fmt.Println("\t" + conf.Port)
					}
					if conf.Identity != DEFAULT_IDEN {
						color.Set(color.FgGreen)
						fmt.Print("\tIdentity File: ")
						color.Unset()
						fmt.Println("\t" + filepath.Join(os.Getenv("HOME"), ".ssh", conf.Identity))
					}
				}
			}
		} else {
			for _, conf := range configs {
				color.Set(color.FgYellow, color.Bold)
				fmt.Println("[" + conf.Name + "]")
				color.Unset()
				color.Set(color.FgGreen)
				fmt.Print("\tUsername: ")
				color.Unset()
				fmt.Println("\t" + conf.Username)
				color.Set(color.FgGreen)
				fmt.Print("\tDomain: ")
				color.Unset()
				fmt.Println("\t" + conf.Domain)
				if conf.Port != DEFAULT_PORT {
					color.Set(color.FgGreen)
					fmt.Print("\tPort: ")
					color.Unset()
					fmt.Println("\t" + conf.Port)
				}
				if conf.Identity != DEFAULT_IDEN {
					color.Set(color.FgGreen)
					fmt.Print("\tIdentity File: ")
					color.Unset()
					fmt.Println("\t" + filepath.Join(os.Getenv("HOME"), ".ssh", conf.Identity))
				}
			}
		}
	} else {
		if len(args) < 1 {
			color.Red("Invalid operation.\n\tFormat: ww <name>\n\tType ww help for more information...")
			os.Exit(1)
		}

		found := false
		for _, conf := range configs {
			if args[0] == conf.Name {
				found = true
				cmdStr := conf.Username + "@" + conf.Domain
				arguments := []string{"-i", filepath.Join(os.Getenv("HOME"), ".ssh", conf.Identity), "-p", conf.Port, cmdStr}
				if len(os.Args) > 2 {
					arguments = append(arguments, os.Args[2:]...)
				}
				cmd := exec.Command("ssh", arguments...)
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin
				cmd.Stderr = os.Stderr
				cmd.Run()
			}
		}

		if found == false {
			color.Red("Could not find %s. Exiting...", args[0])
			os.Exit(1)
		}
	}
}
