package main

import (
	"flag"
	"fmt"
	"github.com/specgen-io/goven/git"
	"github.com/specgen-io/goven/goven"
	"os"
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

func createCmdVendor() Command {
	var gomodPath string
	var vendorRequired bool
	var outPath string
	var vendoredModulesFolder string
	var newModuleName string

	cmdName := "vendor"
	cmd := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmd.StringVar(&gomodPath, "module", "./go.mod", "location of go.mod to be vendored")
	cmd.BoolVar(&vendorRequired, "required", false, "vendor required modules (needs 'go mod vendor' prior goven, deafult false)")
	cmd.StringVar(&outPath, "out", "./out", "path where to put vendored module")
	cmd.StringVar(&vendoredModulesFolder, "vendor", "goven", "internal path where vendored modules should be placed")
	cmd.StringVar(&newModuleName, "name", "", "name of the module after vendoring")

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

		err = goven.Vendor(gomodPath, outPath, newModuleName, vendoredModulesFolder, vendorRequired)
		if err != nil {
			return fmt.Errorf(`vendoring failed: %s`, err.Error())
		}

		return nil
	}}
}

func createCmdRelease() Command {
	var gomodPath string
	var vendorRequired bool
	var outPath string
	var vendoredModulesFolder string
	var repoSlug string
	var majorVersion string
	var version string
	var githubName string
	var githubEmail string
	var githubUser string
	var githubToken string

	cmdName := "release"
	cmd := flag.NewFlagSet(cmdName, flag.ExitOnError)
	cmd.StringVar(&gomodPath, "module", "./go.mod", "location of go.mod to be vendored")
	cmd.BoolVar(&vendorRequired, "required", false, "vendor required modules (needs 'go mod vendor' prior goven, deafult false)")
	cmd.StringVar(&outPath, "out", "./out", "path where to put vendored module")
	cmd.StringVar(&vendoredModulesFolder, "vendor", "goven", "internal path where vendored modules should be placed")
	cmd.StringVar(&repoSlug, "repo", "", "github repo slug to release vendored module to")
	cmd.StringVar(&majorVersion, "major", "", `major version name (has to start with "v" if provided)`)
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

		if repoSlug == "" {
			return fmt.Errorf(`argument "repo" has to be provided`)
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

		newModuleName := fmt.Sprintf(`github.com/%s`, repoSlug)
		if majorVersion != "" {
			newModuleName = fmt.Sprintf(`%s/%s`, newModuleName, majorVersion)
		}

		repoUrl := fmt.Sprintf(`https://github.com/%s.git`, repoSlug)

		err = goven.Vendor(gomodPath, outPath, newModuleName, vendoredModulesFolder, vendorRequired)
		if err != nil {
			return fmt.Errorf(`vendoring failed: %s`, err.Error())
		}

		repoPath := majorVersion
		if repoPath == "" {
			repoPath = "."
		}

		credentials := git.Credentials{githubName, githubEmail, githubName, githubToken}
		err = git.PutFiles(outPath, repoUrl, repoPath, version, credentials)
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
