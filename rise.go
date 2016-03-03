package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/nitrous-io/rise-cli-go/cli/deploy"
	"github.com/nitrous-io/rise-cli-go/cli/login"
	"github.com/nitrous-io/rise-cli-go/cli/logout"
	"github.com/nitrous-io/rise-cli-go/cli/signup"
)

func main() {
	app := cli.NewApp()
	app.Name = "rise"
	app.Usage = "Command line interface for Rise.sh"

	app.Commands = []cli.Command{
		{
			Name:   "signup",
			Usage:  "Create a new Rise account",
			Action: signup.Signup,
		},
		{
			Name:   "login",
			Usage:  "Log in to a Rise account",
			Action: login.Login,
		},
		{
			Name:   "logout",
			Usage:  "Log out from current session",
			Action: logout.Logout,
		},
		{
			Name:   "deploy",
			Usage:  "Deploy your project",
			Action: deploy.Deploy,
		},
	}

	app.Run(os.Args)
}
