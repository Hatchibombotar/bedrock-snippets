package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html"
	"html/template"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/chasefleming/elem-go"
	"github.com/chasefleming/elem-go/attrs"
	"github.com/chasefleming/elem-go/styles"

	"github.com/alecthomas/chroma/v2"
	chromaHtmlFormatter "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
)

type ContentPageData struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

func generatePagesForSnippet(snippetName string) {
	dirName := "./build/snippets/" + snippetName
	sourceHref := REPOSITORY_ROOT + "tree/main/snippets/" + snippetName

	err := os.Mkdir(dirName, os.ModePerm)
	if err != nil {
		panic(err)
	}

	tmpl, err := template.ParseFiles("./html/layout.html", "./html/snippet.html")
	if err != nil {
		fmt.Println("Error parsing template files")
		panic(err)
	}

	sidebar := generateSidebarElement(snippetName, "", 0)

	filepath.WalkDir(SNIPPET_DIRECTORY+"/"+snippetName+"/", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			panic(err)
		}

		rawPath, _ := filepath.Rel("../snippets", path)
		ext := filepath.Ext(path)
		outputPath := "build/snippets/" + rawPath

		if d.IsDir() {
			err := os.Mkdir(outputPath, os.ModePerm)
			if err != nil && !os.IsExist(err) {
				panic(err)
			}
			return nil
		}
		fileName := filepath.Base(path)

		if slices.Contains(ExcludeFiles, fileName) {
			return nil
		}

		outputFilePath := outputPath + ".html"
		outputFile, err := os.Create(outputFilePath)
		if err != nil {
			panic(err)
		}
		defer outputFile.Close()

		data := struct {
			Title          string
			SidebarHTML    template.HTML
			Breadcrumbs    string
			TextContent    template.HTML
			OverviewPage   bool
			SourceHref     string
			ROOT_DIRECTORY string
		}{
			Title:          snippetName,
			SidebarHTML:    template.HTML(sidebar.Render()),
			Breadcrumbs:    strings.Join(strings.Split(filepath.ToSlash(rawPath), "/"), " > "),
			OverviewPage:   false,
			SourceHref:     sourceHref,
			ROOT_DIRECTORY: ROOT_DIRECTORY,
		}

		switch ext {
		case ".json", ".material":
			data.TextContent = CreateJSONPreview(path)
		case ".png":
			data.TextContent = CreatePNGPreview(path)
		case ".md":
			data.TextContent = CreateMDPreview(path)
		case ".js", ".ts":
			data.TextContent = CreateJSPreview(path)
		default:
			log.Fatalf("No preview avaliable for %s", ext)
		}

		err = tmpl.ExecuteTemplate(outputFile, "layout.html", data)
		if err != nil {
			fmt.Println("Error executing template")
			panic(err)
		}

		return nil
	})

	indexFilePath := dirName + "/index.html"
	indexFile, err := os.Create(indexFilePath)
	if err != nil {
		panic(err)
	}
	defer indexFile.Close()

	indexFileContent, err := os.ReadFile(SNIPPET_DIRECTORY + "/" + snippetName + "/snippet.md")
	if err != nil {
		indexFileContent = []byte(`This snippet is missing a snippet.md file.`)
	}

	renderedIndexFileContent := template.HTML(mdToHTML([]byte(indexFileContent)))

	data := struct {
		Title          string
		SidebarHTML    template.HTML
		Breadcrumbs    string
		TextContent    template.HTML
		OverviewPage   bool
		SourceHref     string
		ROOT_DIRECTORY string
	}{
		Title:          snippetName,
		SidebarHTML:    template.HTML(sidebar.Render()),
		Breadcrumbs:    snippetName + " > overview",
		OverviewPage:   true,
		TextContent:    renderedIndexFileContent,
		SourceHref:     sourceHref,
		ROOT_DIRECTORY: ROOT_DIRECTORY,
	}

	err = tmpl.ExecuteTemplate(indexFile, "layout.html", data)
	if err != nil {
		fmt.Println("Error executing template")
		panic(err)
	}
}

func generateSidebarElement(snippetName string, base string, level int) *elem.Element {
	contents, err := os.ReadDir(SNIPPET_DIRECTORY + "/" + snippetName + "/" + base)
	if err != nil {
		panic(err)
	}
	content := elem.Div(attrs.Props{
		attrs.Class: "flex flex-col",
	})

	if level == 0 {
		anchorElement := elem.A(
			attrs.Props{
				attrs.Class: "hover:bg-neutral-200 focus:bg-neutral-200 dark:hover:bg-neutral-700 dark:focus:bg-neutral-700 px-1 py-0.5",
				attrs.Href:  ROOT_DIRECTORY + "/snippets/" + snippetName + "/",
			},
			elem.Text("readme"),
		)

		content.Children = append(content.Children, anchorElement)
	}

	for _, e := range contents {
		if slices.Contains(ExcludeFiles, e.Name()) {
			continue
		}
		anchorElement := elem.A(
			attrs.Props{
				attrs.Class: "hover:bg-neutral-200 focus:bg-neutral-200 dark:hover:bg-neutral-700 dark:focus:bg-neutral-700 px-1 py-0.5",
				attrs.Style: styles.Props{
					styles.PaddingLeft: fmt.Sprint("calc(var(--spacing) * ", (2*level)+1, ")"),
				}.ToInline(),
			},
			elem.Text(e.Name()),
		)
		if !e.IsDir() {
			anchorElement.Attrs[attrs.Href] = ROOT_DIRECTORY + "/snippets/" + snippetName + "/" + base + e.Name()
		}
		content.Children = append(
			content.Children,
			anchorElement,
		)
		if e.IsDir() {
			content.Children = append(
				content.Children,
				generateSidebarElement(snippetName, base+e.Name()+"/", level+1),
			)
		}
	}
	return content
}

func CreateJSONPreview(filePath string) template.HTML {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalln("Unable to read file.", err)
	}

	htmlFormatter := chromaHtmlFormatter.New(
		chromaHtmlFormatter.LineNumbersInTable(true),
		chromaHtmlFormatter.WithClasses(true),
		chromaHtmlFormatter.ClassPrefix("chroma-"),
	)

	lexer := lexers.Get("json")

	iterator, err := lexer.Tokenise(nil, string(content))
	if err != nil {
		panic(err)
	}

	var result bytes.Buffer
	err = htmlFormatter.Format(&result, &chroma.Style{}, iterator)
	if err != nil {
		panic(err)
	}

	container := elem.Div(attrs.Props{
		attrs.ID:            "snippet-content",
		"data-content-type": "text",
		"data-content-text": html.EscapeString(string(content)),
	},
		elem.Raw(result.String()),
	)

	return template.HTML(container.Render())
}

func CreateJSPreview(filePath string) template.HTML {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalln("Unable to read file.", err)
	}

	htmlFormatter := chromaHtmlFormatter.New(
		chromaHtmlFormatter.LineNumbersInTable(true),
		chromaHtmlFormatter.WithClasses(true),
		chromaHtmlFormatter.ClassPrefix("chroma-"),
	)

	lexer := lexers.Get("ts")

	iterator, err := lexer.Tokenise(nil, string(content))
	if err != nil {
		panic(err)
	}

	var result bytes.Buffer
	err = htmlFormatter.Format(&result, &chroma.Style{}, iterator)
	if err != nil {
		panic(err)
	}

	container := elem.Div(attrs.Props{
		attrs.ID:            "snippet-content",
		"data-content-type": "text",
		"data-content-text": html.EscapeString(string(content)),
	},
		elem.Raw(result.String()),
	)

	return template.HTML(container.Render())
}

func CreatePNGPreview(filePath string) template.HTML {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalln("Unable to read file.", err)
	}

	imgBase64Str := base64.StdEncoding.EncodeToString(content)

	result := "<img id=\"snippet-content\" data-content-type=\"image\" class=\"preview-image\" src=\"data:image/png;base64," + imgBase64Str + "\" >"

	return template.HTML(result)
}

func CreateMDPreview(filePath string) template.HTML {
	indexFileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalln("Unable to read file.", err)
	}

	renderedIndexFileContent := mdToHTML([]byte(indexFileContent))
	return template.HTML(renderedIndexFileContent)
}
