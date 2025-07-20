package parser

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ParsedDocument struct {
	Filename   string
	Text       string
	Categories []string
}

func ParseDocuments(root string) ([]ParsedDocument, error) {
	var docs []ParsedDocument

	fmt.Print("Parsing documents")

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		fmt.Print(".")
		if err != nil {
			fmt.Println("Failed accessing:", path, "error:", err)
			return nil
		}

		if d.IsDir() {
			return nil
		}

		categories := getCategories(path)

		ext := strings.ToLower(filepath.Ext(d.Name()))
		switch ext {
		case ".docx":
			text, err := parseDOCX(path)
			if err != nil {
				fmt.Println("Pandoc parse error:", err)
				return nil
			}
			docs = append(docs, ParsedDocument{Filename: path, Text: text, Categories: categories})

		case ".pdf":
			text, err := parsePDF(path)
			if err != nil {
				fmt.Println("Pandoc parse error:", err)
				return nil
			}
			docs = append(docs, ParsedDocument{Filename: path, Text: text, Categories: categories})

		case ".txt":
			content, err := os.ReadFile(path)
			if err != nil {
				fmt.Println("TXT read error:", err)
				return nil
			}
			docs = append(docs, ParsedDocument{Filename: path, Text: string(content), Categories: categories})
		}

		fmt.Print(".")

		return nil
	})

	fmt.Println()

	return docs, err
}

type Context struct {
	Categories []string           `yaml:"categories"`
	Files      map[string]Context `yaml:"overwrites"`
}

func getCategories(path string) []string {
	var context Context
	contextFilePath := filepath.Join(filepath.Dir(path), "CONTEXT.yml")
	contextFile, err := os.ReadFile(contextFilePath)

	if err != nil {
		return []string{}
	}

	if err := yaml.Unmarshal(contextFile, &context); err != nil {
		fmt.Println("Failed to parse ", contextFilePath, err)
		return []string{}
	}

	overwrite, ok := context.Files[filepath.Base(path)]

	if ok {
		return overwrite.Categories
	}

	return context.Categories
}

func parseDOCX(path string) (string, error) {
	out, err := exec.Command("pandoc", "-t", "plain", path).Output()
	if err != nil {
		return "", fmt.Errorf("pandoc error: %w", err)
	}
	return string(out), nil
}

func parsePDF(path string) (string, error) {
	out, err := exec.Command("pdftotext", path, "-").Output()
	if err != nil {
		return "", fmt.Errorf("pdftotext error: %w", err)
	}
	return string(out), nil
}
