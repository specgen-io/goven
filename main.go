package main

import (
	"flag"
	"fmt"
	"github.com/specgen-io/goven/git"
	"github.com/specgen-io/goven/goven"
	"os"
	"regexp"
	"strings"
)

func main() {
	execute(createCommands())
}

func createCommands() Commands {
	return Commands{
		createCmdVendor(),
		createCmdRelease(),
	}
}

type stringArray []string

func (arr *stringArray) String() string {
	return strings.Join(*arr, ", ")
}

func (arr *stringArray) Set(value string) error {
	*arr = append(*arr, value)
	return nil
}

type CommonOptions struct {
	gomodPath             string
	vendorRequired        bool
	outPath               string
	vendoredModulesFolder string
	newModuleName         string
	ignorePaths           stringArray
}

func (options *CommonOptions) Add(cmd *flag.FlagSet) {
	cmd.StringVar(&options.gomodPath, "module", "./go.mod", "location of go.mod to be vendored")
	cmd.BoolVar(&options.vendorRequired, "required", false, "vendor required modules (needs 'go mod vendor' prior goven, default: false)")
	cmd.StringVar(&options.outPath, "out", "./out", "path where to put vendored module")
	cmd.StringVar(&options.vendoredModulesFolder, "vendor", "goven", "internal path where vendored modules should be placed")
	cmd.StringVar(&options.newModuleName, "name", "", "name of the module after vendoring")
	cmd.Var(&options.ignorePaths, "ignore", "folders to ignore during vendoring")
}

func createCmdVendor() Command {
	cmdName := "vendor"
	cmd := flag.NewFlagSet(cmdName, flag.ExitOnError)

	common := CommonOptions{}
	common.Add(cmd)

	parse := func(arguments []string) error {
		err := cmd.Parse(arguments)
		if err != nil {
			return err
		}
		return nil
	}

	return Command{cmdName, func(arguments []string) error {
		err := parse(arguments)
		if err != nil {
			return err
		}

		err = goven.Vendor(common.gomodPath, common.outPath, common.newModuleName, common.vendoredModulesFolder, common.vendorRequired, common.ignorePaths)
		if err != nil {
			return fmt.Errorf(`vendoring failed: %s`, err.Error())
		}

		return nil
	}}
}

func splitVersion(moduleName string) (string, string) {
	parts := strings.Split(moduleName, "/")
	version := parts[len(parts)-1]
	matched, _ := regexp.Match("v[0-9]+", []byte(version))
	if matched {
		name := strings.TrimSuffix(moduleName, fmt.Sprintf(`/%s`, version))
		return name, version
	}
	return moduleName, ""
}

func createCmdRelease() Command {
	var version string
	var githubName string
	var githubEmail string
	var githubUser string
	var githubToken string

	cmdName := "release"
	cmd := flag.NewFlagSet(cmdName, flag.ExitOnError)

	common := CommonOptions{}
	common.Add(cmd)

	cmd.StringVar(&version, "version", "", `version to release (has to be in format: "vMAJOR.MINOR.BUILD")`)
	cmd.StringVar(&githubName, "github-name", os.Getenv("GITHUB_NAME"), "github commit author name (this is just a readable name - NOT a user name)")
	cmd.StringVar(&githubEmail, "github-email", os.Getenv("GITHUB_EMAIL"), "github commit author email")
	cmd.StringVar(&githubUser, "github-user", os.Getenv("GITHUB_USER"), "github user account to be used for push")
	cmd.StringVar(&githubToken, "github-token", os.Getenv("GITHUB_TOKEN"), "github token to be used for push")

	parse := func(arguments []string) error {
		err := cmd.Parse(arguments)
		if err != nil {
			return err
		}

		if githubName == "" {
			return fmt.Errorf(`argument "github-name" has to be provided or set via environment variable "GITHUB_NAME"`)
		}

		if githubEmail == "" {
			return fmt.Errorf(`argument "github-email" has to be provided or set via environment variable "GITHUB_EMAIL"`)
		}

		if githubUser == "" {
			return fmt.Errorf(`argument "github-user" has to be provided or set via environment variable "GITHUB_USER"`)
		}

		if githubToken == "" {
			return fmt.Errorf(`argument "github-token" has to be provided or set via environment variable "GITHUB_TOKEN"`)
		}

		return nil
	}

	return Command{cmdName, func(arguments []string) error {
		err := parse(arguments)
		if err != nil {
			return err
		}

		err = goven.Vendor(common.gomodPath, common.outPath, common.newModuleName, common.vendoredModulesFolder, common.vendorRequired, common.ignorePaths)
		if err != nil {
			return fmt.Errorf(`vendoring failed: %s`, err.Error())
		}

		moduleName := common.newModuleName
		if moduleName == "" {
			moduleName, err = goven.ModuleName(common.gomodPath)
			if err != nil {
				return fmt.Errorf(`can't get module name: %s`, err.Error())
			}
		}

		repo, majorVersion := splitVersion(moduleName)
		repoUrl := fmt.Sprintf(`https://%s.git`, repo)

		repoPath := majorVersion
		if repoPath == "" {
			repoPath = "."
		}

		credentials := git.Credentials{githubName, githubEmail, githubName, githubToken}
		err = git.PutFiles(common.outPath, repoUrl, repoPath, version, credentials)
		if err != nil {
			return fmt.Errorf(`saving to github failed: %s`, err.Error())
		}

		return nil
	}}
}

type Command struct {
	Name string
	Run  func(arguments []string) error
}

type Commands []Command

func execute(commands Commands) {
	if len(os.Args) < 2 {
		fmt.Println(`expected subcommand`)
		os.Exit(1)
	}

	name := os.Args[1]

	if name == "-help" {
		flag.Usage()
		return
	}

	for _, command := range commands {
		if command.Name == name {
			err := command.Run(os.Args[2:])
			if err != nil {
				fmt.Printf(`command "%s" failed: %s`, command.Name, err.Error())
				os.Exit(1)
			}
			return
		}
	}

	fmt.Println(`unknown command: %s`, name)
	os.Exit(1)
}
