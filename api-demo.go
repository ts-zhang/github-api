package main

import (
	"bufio"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"strings"
	"syscall"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/google/go-github/v29/github"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/oauth2"
)

var ctx = context.Background()

func showUserPage(client *github.Client) {
	ui.Clear()
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	resp, err := http.Get(*user.AvatarURL)
	image, _, err := image.Decode(resp.Body)
	userAvatar := widgets.NewImage(image)
	userAvatar.SetRect(0, 0, 20, 10)

	userInfo := widgets.NewParagraph()
	userInfo.Title = "UserInfo"
	userInfo.Text = fmt.Sprintf("Name: [%s](fg:yellow,mod:bold)\nLogin: %s\nLocation: %s\nBlog: %s\n%s([C](fg:red))\n%s([U](fg:green))", *user.Name, *user.Login, *user.Location, *user.Blog, *user.CreatedAt, *user.UpdatedAt)
	userInfo.SetRect(21, 0, 61, 10)

	userPlan := widgets.NewGauge()
	userPlan.Title = fmt.Sprintf("UserPlan(%s)", *user.Plan.Name)
	userPlan.SetRect(0, 15, 61, 12)
	userPlan.BarColor = ui.ColorYellow
	userPlan.LabelStyle = ui.NewStyle(ui.ColorBlue)
	userPlan.BorderStyle.Fg = ui.ColorWhite
	userPlan.Percent = (*user.TotalPrivateRepos * 100) / *user.Plan.PrivateRepos
	userPlan.Label = fmt.Sprintf("%d/%d", *user.TotalPrivateRepos, *user.Plan.PrivateRepos)

	usage := widgets.NewParagraph()
	usage.Title = "Usage"
	usage.Text = "Press [h](fg:green) to show help\nPress [u](fg:yellow) to show user info\nPress [p](fg:blue) to show repo list\nPress [q](fg:red) to exit."
	usage.SetRect(0, 17, 61, 23)

	ui.Render(userAvatar, userInfo, userPlan, usage)
}

func showRepoPage(client *github.Client) {
	ui.Clear()

	var repoList []string

	opts := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{
			PerPage: 15,
		},
	}

	repos, resp, err := client.Repositories.List(ctx, "", opts)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, repo := range repos {
		repoList = append(repoList, *repo.Name)
	}

	repoListUI := widgets.NewList()
	repoListUI.Title = "Repositories"
	repoListUI.Rows = repoList
	repoListUI.TextStyle = ui.NewStyle(ui.ColorYellow)
	repoListUI.WrapText = false
	repoListUI.SetRect(0, 0, 50, 25)

	pageInfo := widgets.NewParagraph()
	pageInfo.Title = ""
	pageInfo.Text = fmt.Sprintf("[First:%d (%d/%d) Last:%d]", resp.FirstPage, resp.PrevPage, resp.NextPage, resp.LastPage)
	pageInfo.SetRect(0, 27, 50, 32)

	ui.Render(repoListUI, pageInfo)
}

func showHelpPage(client *github.Client) {
	ui.Clear()
	helpInfo := widgets.NewParagraph()
	helpInfo.Title = "Usage"
	helpInfo.Text = "Press [h](fg:green) to show help\nPress [u](fg:yellow) to show user info\nPress [p](fg:blue) to show repo list\nPress [q](fg:red) to exit."
	helpInfo.SetRect(0, 0, 50, 25)
	ui.Render(helpInfo)
}

func main() {
	var httpClient *http.Client

	githubToken := os.Getenv("GITHUB_AUTH_TOKEN")

	if len(githubToken) == 0 {
		//use username and password to auth
		r := bufio.NewReader(os.Stdin)
		print("Username: ")
		username, _ := r.ReadString('\n')

		print("Password: ")
		bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
		password := string(bytePassword)

		authTp := github.BasicAuthTransport{
			Username:  strings.TrimSpace(username),
			Password:  strings.TrimSpace(password),
			OTP:       "",
			Transport: nil,
		}
		httpClient = authTp.Client()
	} else {
		//use github token to auth
		token := oauth2.Token{AccessToken: githubToken}

		src := oauth2.StaticTokenSource(&token)

		httpClient = oauth2.NewClient(ctx, src)
	}

	client := github.NewClient(httpClient)

	ui.Init()

	defer ui.Close()

	showRepoPage(client)

	uiEvents := ui.PollEvents()

	for {
		e := <-uiEvents
		if e.Type == ui.KeyboardEvent {
			switch e.ID {
			case "h":
				//help
				showHelpPage(client)
			case "u":
				//user info
				showUserPage(client)
			case "p":
				//repo list
				showRepoPage(client)
			case "q", "<C-c>":
				os.Exit(0)
			}
		} else if e.Type == ui.ResizeEvent {

		}
	}
}
