package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	HOME_ENV     = "HOME"
	CFG_FILENAME = ".shbconfig"
)

type Config struct {
	Name     string `json:"name"`
	Domain   string `json:"domain"`
	Username string `json:"username"`
}

func main() {
	homePath, exists := os.LookupEnv(HOME_ENV)
	if !exists {
		log.Fatal("Could not find environment variable HOME. Cannot proceed with setup. Exiting...")
	}

	chdirErr := os.Chdir(homePath)
	if chdirErr != nil {
		log.Fatal("A problem occurred when trying to access the config file. Exiting...")
	}

	configFile, err := os.OpenFile(CFG_FILENAME, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal("Error opening the config file. Exiting...")
	}
	configFile.Close()

	data, err := ioutil.ReadFile(CFG_FILENAME)
	if err != nil {
		log.Fatal("Could not find config file. Exiting...")
	}

	var configs []Config
	json.Unmarshal(data, &configs)

	args := os.Args[1:]

	if args[0] == "add" {
		if len(args) != 3 {
			fmt.Println("Invalid operation.\n\tFormat: shb add <name> <username>@<domain>\n\tType shb help for more information...")
			os.Exit(1)
		}

		var newEntry Config
		for _, conf := range configs {
			if args[1] == conf.Name {
				fmt.Printf("Name %s is already in use", conf.Name)
				os.Exit(1)
			}
		}

		domainAfterSplit := strings.Split(args[2], "@")
		if len(domainAfterSplit) == 1 {
			fmt.Println("Enter a valid domain in the format <username>@<domain>")
			os.Exit(1)
		}

		newEntry.Name = args[1]
		newEntry.Username = domainAfterSplit[0]
		newEntry.Domain = domainAfterSplit[1]
		configs = append(configs, newEntry)

		confs, err := json.Marshal(configs)
		if err != nil {
			log.Fatal("Error marshalling JSON object. Exiting...")
		}

		err = ioutil.WriteFile(CFG_FILENAME, confs, 0755)
		if err != nil {
			log.Fatal("Error writing to file. Exiting...")
		}
	} else if args[0] == "remove" {
		if len(args) != 2 {
			fmt.Println("Invalid operation.\n\tFormat: shb remove <name>\n\tType shb help for more information...")
			os.Exit(1)
		}

		for idx, conf := range configs {
			if args[1] == conf.Name {
				configs = append(configs[:idx], configs[idx+1:]...)
			}
		}

		confs, err := json.Marshal(configs)
		if err != nil {
			log.Fatal("Error marshalling JSON object. Exiting...")
		}

		err = ioutil.WriteFile(CFG_FILENAME, confs, 0755)
		if err != nil {
			log.Fatal("Error writing to file. Exiting...")
		}
	} else if args[0] == "edit" {
		// TODO
	} else {
		if len(args) != 1 {
			fmt.Println("Invalid operation.\n\tFormat: shb <name>\n\tType shb help for more information...")
			os.Exit(1)
		}

		found := false
		for _, conf := range configs {
			if args[0] == conf.Name {
				found = true
				cmdStr := conf.Username + "@" + conf.Domain
				cmd := exec.Command("ssh", cmdStr)
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin
				cmd.Stderr = os.Stderr
				cmd.Run()
			}
		}

		if found == false {
			fmt.Printf("Could not find %s. Exiting...", args[0])
			os.Exit(1)
		}
	}
}
