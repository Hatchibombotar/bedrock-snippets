package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"slices"

	"github.com/chasefleming/elem-go"
	"github.com/chasefleming/elem-go/attrs"
)

func generateHomepage() {
	homepageLinks := generateHomepageLinks()

	snippetContributors := generateContributors(SNIPPET_REPO_OWNER, SNIPPET_REPO_NAME)
	websiteContributors := generateContributors("hatchibombotar", "bedrock-snippets")

	tmpl, err := template.ParseFiles("./html/layout.html", "./html/home.html")
	if err != nil {
		fmt.Println("Error parsing template files")
		panic(err)
	}
	outputFile, err := os.Create("build/index.html")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	data := struct {
		Links               template.HTML
		SnippetContributors template.HTML
		WebsiteContributors template.HTML
		Title               string
		ROOT_DIRECTORY      string
	}{
		Links:               homepageLinks,
		SnippetContributors: snippetContributors,
		WebsiteContributors: websiteContributors,
		Title:               "Home",
		ROOT_DIRECTORY:      ROOT_DIRECTORY,
	}

	err = tmpl.ExecuteTemplate(
		outputFile, "layout.html",
		data,
	)

	if err != nil {
		fmt.Println("Error executing template")
		panic(err)
	}
}

func generateHomepageLinks() template.HTML {
	content := elem.Div(attrs.Props{
		attrs.Class: "flex flex-row flex-wrap gap-3 p-2",
	})

	snippets, err := os.ReadDir(SNIPPET_DIRECTORY)
	if err != nil {
		panic(err)
	}

	for _, category := range Categories {
		categoryDiv := elem.Div(
			attrs.Props{
				attrs.Class: "flex flex-col",
			},
			elem.Div(
				attrs.Props{
					attrs.Class: "flex flex-row",
				},
				elem.Img(
					attrs.Props{
						attrs.Src:   ROOT_DIRECTORY + category.Icon,
						attrs.Class: "w-6 h-6 inline mr-1",
						attrs.Alt:   "",
					},
				),
				elem.H2(attrs.Props{
					attrs.Class: "font-semibold",
				},
					elem.Text(category.DisplayName),
				),
			),
		)

		snippetsIncluded := 0

		for _, e := range snippets {
			if !e.IsDir() {
				continue
			}
			snippetId := e.Name()

			var snippetData SnippetData
			snippetDataJson, err := os.ReadFile(SNIPPET_DIRECTORY + "/" + snippetId + "/meta.json")
			if err != nil {
				fmt.Println("Snippet", snippetId, "is missing a meta.json file.")
				continue
			}

			err = json.Unmarshal([]byte(snippetDataJson), &snippetData)
			if err != nil {
				panic("Error decoding JSON.")
			}

			categoryIncluded := slices.Contains(snippetData.Tags, category.Id)
			if categoryIncluded {
				snippetsIncluded += 1
				categoryDiv.Children = append(categoryDiv.Children,
					elem.A(
						attrs.Props{
							attrs.Href: ROOT_DIRECTORY + "/snippets/" + snippetId,
						},
						elem.Span(attrs.Props{
							attrs.Class: "text-blue-600 dark:text-blue-500",
						},
							elem.Text(snippetData.Name),
						),
					),
				)
			}
		}

		if snippetsIncluded > 0 {
			content.Children = append(content.Children, categoryDiv)
		}
	}

	return template.HTML(content.Render())
}

func generateContributors(owner, repo string) template.HTML {
	contributors, err := GetGitHubContributorsAPI(owner, repo, "")
	if err != nil {
		fmt.Println(err)
		return ""
	}

	content := elem.Div(attrs.Props{
		attrs.Class: "flex flex-row flex-wrap gap-3 p-2",
	})

	for _, contributor := range contributors {
		contributorDiv := elem.A(
			attrs.Props{
				attrs.Href:  contributor.HTMLURL,
				attrs.Title: contributor.Login,
				attrs.Class: "h-8 w-8",
			},
			elem.Img(
				attrs.Props{
					attrs.Src:   contributor.AvatarURL,
					attrs.Alt:   contributor.Login,
					attrs.Class: "rounded-full",
				},
			),
		)
		content.Children = append(content.Children, contributorDiv)
	}

	return template.HTML(content.Render())
}
