package main

import (
	"fmt"

	"github.com/jxskiss/gopkg/easy"
	"github.com/jxskiss/gopkg/mcli"
)

func main() {

	mcli.Add("browse", cmdBrowse, "Open the repository in the browser")

	mcli.AddGroup("codespace", "Connect to and manage your codespaces")

	/*
		Work with GitHub gists.

		USAGE
		  gh gist <command> [flags]

		CORE COMMANDS
		  clone:      Clone a gist locally
		  create:     Create a new gist
		  delete:     Delete a gist
		  edit:       Edit one of your gists
		  list:       List your gists
		  view:       View a gist

		INHERITED FLAGS
		  --help   Show help for command

		ARGUMENTS
		  A gist can be supplied as argument in either of the following formats:
		  - by ID, e.g. 5b0e0062eb8e9654adad7bb1d81cc75f
		  - by URL, e.g. "https://gist.github.com/OWNER/5b0e0062eb8e9654adad7bb1d81cc75f"

		LEARN MORE
		  Use 'gh <command> <subcommand> --help' for more information about a command.
		  Read the manual at https://cli.github.com/manual

	*/
	mcli.AddGroup("gist", "Manage gists")
	mcli.Add("gist clone", cmdGistClone, "Clone a gist locally")
	mcli.Add("gist create", cmdGistCreate, "Create a new gist")
	mcli.Add("gist delete", cmdGistDelete, "Delete a gist")
	mcli.Add("gist edit", cmdGistEdit, "Edit one of your gists")
	mcli.Add("gist list", cmdGistList, "List your gists")
	mcli.Add("gist view", cmdGistView, "View a gist")

	mcli.AddGroup("issue", "Manage issues")
	mcli.Add("issue close", cmdIssueClose, "Close issue")
	mcli.Add("issue comment", cmdIssueComment, "Create a new issue comment")
	mcli.Add("issue create", cmdIssueCreate, "Create a new issue")
	mcli.Add("issue delete", cmdIssueDelete, "Delete issue")
	mcli.Add("issue edit", cmdIssueEdit, "Edit an issue")
	mcli.Add("issue list", cmdIssueList, "List and filter issues in this repository")
	mcli.Add("issue repopen", cmdIssueReopen, "Reopen issue")
	mcli.Add("issue status", cmdIssueStatus, "Show status of relevant issues")
	mcli.Add("issue transfer", cmdIssueTransfer, "Transfer issue to another repository")
	mcli.Add("issue view", cmdIssueView, "View an issue")

	mcli.AddGroup("pr", "Manage pull requests")

	mcli.AddGroup("release", "Manage GitHub releases")

	mcli.AddGroup("repo", "Create, clone, fork, and view repositories")

	mcli.AddGroup("actions", "Learn about working with GitHub Actions")

	mcli.AddGroup("run", "View details about workflow runs")

	mcli.AddGroup("workflow", "View details about GitHub Actions workflows")

	mcli.AddGroup("alias", "Create command shortcuts")

	mcli.AddGroup("api", "Make an authenticated GitHub API request")

	mcli.AddGroup("auth", "Login, logout, and refresh your authentication")

	mcli.AddGroup("completion", "Generate shell completion scripts")

	mcli.AddGroup("config", "Manage configuration for gh")

	mcli.AddGroup("extension", "Manage gh extensions")

	mcli.AddGroup("gpg-key", "Manage GPG keys")

	mcli.AddGroup("help", "Help about any command")

	mcli.AddGroup("secret", "Manage GitHub secrets")

	mcli.AddGroup("ssh-key", "Manage SSH keys")

	mcli.Run()
}

/*
Open the GitHub repository in the web browser.

USAGE
  gh browse [<number> | <path>] [flags]

FLAGS
  -b, --branch string            Select another branch by passing in the branch name
  -c, --commit                   Open the last commit
  -n, --no-browser               Print destination URL instead of opening the browser
  -p, --projects                 Open repository projects
  -R, --repo [HOST/]OWNER/REPO   Select another repository using the [HOST/]OWNER/REPO format
  -s, --settings                 Open repository settings
  -w, --wiki                     Open repository wiki

INHERITED FLAGS
  --help   Show help for command

ARGUMENTS
  A browser location can be specified using arguments in the following format:
  - by number for issue or pull request, e.g. "123"; or
  - by path for opening folders and files, e.g. "cmd/gh/main.go"

EXAMPLES
  $ gh browse
  #=> Open the home page of the current repository

  $ gh browse 217
  #=> Open issue or pull request 217

  $ gh browse --settings
  #=> Open repository settings

  $ gh browse main.go:312
  #=> Open main.go at line 312

  $ gh browse main.go --branch main
  #=> Open main.go in the main branch

ENVIRONMENT VARIABLES
  To configure a web browser other than the default, use the BROWSER environment variable.

LEARN MORE
  Use 'gh <command> <subcommand> --help' for more information about a command.
  Read the manual at https://cli.github.com/manual

*/
func cmdBrowse() {
	args := struct {
		Branch    string `cli:"-b, --branch, Select another branch by passing in the branch name"`
		Commit    bool   `cli:"-c, --commit, Open the last commit"`
		NoBrowser bool   `cli:"-n, --no-browser, Print destination URL instead of opening the browser"`
		Projects  bool   `cli:"-p, --projects, Open repository projects"`
		Repo      string `cli:"-R, --repo, Select another repository using the '[HOST/]OWNER/REPO' format"`
		Settings  bool   `cli:"-s, --settings, Open repository settings"`
		Wiki      bool   `cli:"-w, --wiki, Open repository wiki"`

		Location string `cli:"location"`
	}{}
	mcli.Parse(&args)
	mcli.PrintHelp()
}

func cmdGistClone() {
	mcli.PrintHelp()
}

/*
Create a new GitHub gist with given contents.

Gists can be created from one or multiple files. Alternatively, pass "-" as
file name to read from standard input.

By default, gists are secret; use '--public' to make publicly listed ones.


USAGE
  gh gist create [<filename>... | -] [flags]

FLAGS
  -d, --desc string       A description for this gist
  -f, --filename string   Provide a filename to be used when reading from standard input
  -p, --public            List the gist publicly (default: secret)
  -w, --web               Open the web browser with created gist

INHERITED FLAGS
  --help   Show help for command

EXAMPLES
  # publish file 'hello.py' as a public gist
  $ gh gist create --public hello.py

  # create a gist with a description
  $ gh gist create hello.py -d "my Hello-World program in Python"

  # create a gist containing several files
  $ gh gist create hello.py world.py cool.txt

  # read from standard input to create a gist
  $ gh gist create -

  # create a gist from output piped from another command
  $ cat cool.txt | gh gist create

LEARN MORE
  Use 'gh <command> <subcommand> --help' for more information about a command.
  Read the manual at https://cli.github.com/manual

*/
func cmdGistCreate() {

	// Between flag name and description, we allow splitting by spaces,
	// semantically there is no ambiguity.

	var args struct {
		Desc     string `cli:"-d, --desc       A description for this gist"`
		Filename string `cli:"-f, --filename       Provide a filename to be used when reading from standard input"`
		Public   bool   `cli:"-p, --public     List the gist publicly (default: secret)"`
		Web      bool   `cli:"-w, --web            Open the web browser with created gist, some comma, 1, 2, 3"`
	}
	mcli.Parse(&args)
	mcli.PrintHelp()
	fmt.Println(easy.Pretty(&args))
}

func cmdGistDelete() {
	mcli.PrintHelp()
}

func cmdGistEdit() {
	mcli.PrintHelp()
}

func cmdGistList() {
	mcli.PrintHelp()
}

func cmdGistView() {
	mcli.PrintHelp()
}

type commonIssueArgs struct {
	Repo string `cli:"-R, --repo, Select another repository using the '[HOST/]OWNER/REPO' format"`
}

/*
Edit an issue

USAGE
  gh issue edit {<number> | <url>} [flags]

FLAGS
      --add-assignee login      Add assigned users by their login. Use "@me" to assign yourself.
      --add-label name          Add labels by name
      --add-project name        Add the issue to projects by name
  -b, --body string             Set the new body.
  -F, --body-file file          Read body text from file (use "-" to read from standard input)
  -m, --milestone name          Edit the milestone the issue belongs to by name
      --remove-assignee login   Remove assigned users by their login. Use "@me" to unassign yourself.
      --remove-label name       Remove labels by name
      --remove-project name     Remove the issue from projects by name
  -t, --title string            Set the new title.

INHERITED FLAGS
      --help                     Show help for command
  -R, --repo [HOST/]OWNER/REPO   Select another repository using the [HOST/]OWNER/REPO format

EXAMPLES
  $ gh issue edit 23 --title "I found a bug" --body "Nothing works"
  $ gh issue edit 23 --add-label "bug,help wanted" --remove-label "core"
  $ gh issue edit 23 --add-assignee "@me" --remove-assignee monalisa,hubot
  $ gh issue edit 23 --add-project "Roadmap" --remove-project v1,v2
  $ gh issue edit 23 --milestone "Version 1"
  $ gh issue edit 23 --body-file body.txt

LEARN MORE
  Use 'gh <command> <subcommand> --help' for more information about a command.
  Read the manual at https://cli.github.com/manual

*/
func cmdIssueEdit() {
	var args struct {
		AddAssignee    bool   `cli:"--add-assignee      Add assigned users by their 'login'. Use \"@me\" to assign yourself."`
		AddLabel       string `cli:"--add-label          Add labels by 'name'"`
		AddProject     string `cli:"--add-project        Add the issue to projects by 'name'"`
		Body           string `cli:"-b, --body             Set the new body."`
		BodyFile       string `cli:"-F, --body-file          Read body text from 'file' (use \"-\" to read from standard input)"`
		Milestone      string `cli:"-m, --milestone          Edit the milestone the issue belongs to by 'name'"`
		RemoveAssignee string `cli:"--remove-assignee   Remove assigned users by their 'login'. Use \"@me\" to unassign yourself."`
		RemoveLabel    string `cli:"--remove-label       Remove labels by 'name'"`
		RemoveProject  string `cli:"--remove-project     Remove the issue from projects by 'name'"`
		Title          string `cli:"-t, --title            Set the new title."`
		commonIssueArgs
	}
	mcli.Parse(&args)
	mcli.PrintHelp()
	fmt.Println(easy.Pretty(&args))
}

/*
Close issue

USAGE
  gh issue close {<number> | <url>} [flags]

INHERITED FLAGS
      --help                     Show help for command
  -R, --repo [HOST/]OWNER/REPO   Select another repository using the [HOST/]OWNER/REPO format

LEARN MORE
  Use 'gh <command> <subcommand> --help' for more information about a command.
  Read the manual at https://cli.github.com/manual

*/
func cmdIssueClose() {
	var args struct {
		commonIssueArgs
	}
	mcli.Parse(&args)
	mcli.PrintHelp()
	fmt.Println(easy.Pretty(&args))
}

/*
Create a new issue comment

USAGE
  gh issue comment {<number> | <url>} [flags]

FLAGS
  -b, --body string      Supply a body. Will prompt for one otherwise.
  -F, --body-file file   Read body text from file (use "-" to read from standard input)
  -e, --editor           Add body using editor
  -w, --web              Add body in browser

INHERITED FLAGS
      --help                     Show help for command
  -R, --repo [HOST/]OWNER/REPO   Select another repository using the [HOST/]OWNER/REPO format

EXAMPLES
  $ gh issue comment 22 --body "I was able to reproduce this issue, lets fix it."

LEARN MORE
  Use 'gh <command> <subcommand> --help' for more information about a command.
  Read the manual at https://cli.github.com/manual

*/
func cmdIssueComment() {
	var args struct {
		Body     string `cli:"-b, --body      Supply a body. Will prompt for one otherwise."`
		BodyFile string `cli:"-F, --body-file   Read body text from 'file' (use \"-\" to read from standard input)"`
		Editor   bool   `cli:"-e, --editor, Add body using editor"`
		Web      bool   `cli:"-w, --web, Add body in browser"`
		commonIssueArgs
	}
	mcli.Parse(&args)
	mcli.PrintHelp()
	fmt.Println(easy.Pretty(&args))
}

func cmdIssueCreate() {
	mcli.PrintHelp()
}

func cmdIssueDelete() {
	mcli.PrintHelp()
}

func cmdIssueList() {
	mcli.PrintHelp()
}

func cmdIssueReopen() {
	mcli.PrintHelp()
}

func cmdIssueStatus() {
	mcli.PrintHelp()
}

func cmdIssueTransfer() {
	mcli.PrintHelp()
}

func cmdIssueView() {
	mcli.PrintHelp()
}
