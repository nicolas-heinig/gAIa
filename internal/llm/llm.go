package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/philippgille/chromem-go"
	"github.com/sanity-io/litter"
)

type AnalyseResult struct {
	Keywords   []string `json:"keywords"`
	Categories []string `json:"categories"`
}

var analysePromt = `
	Du bekommst eine Frage und sollst die wichtigsten Begriffe und Kategorien daraus extrahieren.
	Versuche moeglichst viele relevante Kategorien zu extrahieren, die in der Frage vorkommen.
	Moegliche Kategorien sind:
	- character
	- nick
	- lore
	- alchemy
	- magic
	- song
	- faith

	Wenn du dir nicht sicher bist, nehme die Kategories "lore" und "character".

	Antworte AUF JEDEN FALL auf Deutsch!
	Erwaehne NIEMALS die Datenbank, den Kontext oder die Suchergebnisse in deiner Antwort.
	Die Antwort soll in der folgenden Form sein:
	{
		"keywords": ["Waldtempler", "Waldläufer"],
		"categories": ["lore", "character"]
	}
	Hier ein paar Beispiele:
	Frage: Was ist der Unterschied zwischen einem Waldtempler und einem Waldläufer?
	Die Antwort soll in der folgenden Form sein:
	{
		"keywords": ["Waldtempler", "Waldläufer"],
		"categories": ["lore", "character"]
	}
	Frage: Wer ist der Anführer der Waldtempler?
	{
		"keywords": ["Anführer", "Waldtempler"],
		"categories": ["character"]
	}
	Frage: Welche Zutaten braucht man fuer Wundalkohol?
	{
		"keywords": ["Zutaten", "Wundalkohol"],
		"categories": ["alchemy"]
	}
	Frage: Ist es sinnvoll, hier, jetzt, bei diesem Verletzten Magie anzuwenden?
	{
		"keywords": ["Verletzen", "Magie", "anwenden"],
		"categories": ["magic"]
	}
	Frage: Was sind die 5 Paladenstiere
	{
		"keywords": ["Paladenstiere", "5"],
		"categories": ["lore"]
	}
	Frage: Wer ist Hans?
	{
		"keywords": ["Hans"],
		"categories": ["character"]
	}
`

func AnalyseQuery(question string, verbose bool) (AnalyseResult, error) {
	var context strings.Builder
	context.WriteString(analysePromt + "\n\n")
	context.WriteString("Frage: " + question + "\n")

	cmd := exec.Command("ollama", "run", "--format", "json", "nuextract")
	cmd.Stdin = bytes.NewBufferString(context.String())

	output, err := cmd.Output()

	if err != nil {
		return AnalyseResult{}, fmt.Errorf("ollama failed: %w\n%s", err, string(output))
	}

	var result AnalyseResult
	err = json.Unmarshal(output, &result)

	if err != nil {
		return AnalyseResult{}, fmt.Errorf("ollama failed to parse output: %w\n%s", err, string(output))
	}

	if verbose {
		fmt.Println("\n\nAnalyse Result:")
		fmt.Println(litter.Sdump(result))
	}

	return result, nil
}

var questionPrompt = `
Du bist ein aldarischer Gelehrte der viel ueber Aldaria und den Orden der Waldtempler weiss.
Du bekommst eine Frage und sollst die Antwort auf die Frage geben.
Die Frage ist "%s".
Die Frage kann sich auf Lore, Charaktere, Alchemie, Magie, Lieder oder Glauben beziehen.
Wenn du dir nicht sicher bist, ob du die Frage beantworten kannst, dann antworte mit "Das weiss ich nicht".
Antworte nicht mit etwas, das nicht mit Aldaria oder den Waldtemplern zu tun hat.
Frage zu Personen drehen sich meist um Charaktere.

Antworte AUF JEDEN FALL auf Deutsch!

Erwaehne NIEMALS die Datenbank, den Kontext oder die Suchergebnisse in deiner Antwort.

Beispiele:
Frage: Was sind die 5 Paladenstiere?
Antwort: Funf Paladenstiere sind die heiligen Tiere des Ordens der Waldtempler. Sie sind: der Hirsch, der Fuchs, der Baer, die Schlange und die Eule.

Hier sind die Teile unserer Datenbank, den du als Kontext benutzen kannnst:
`

func AskQuestion(question string, results []chromem.Result, verbose bool) (string, error) {
	var context strings.Builder
	context.WriteString(fmt.Sprintf(questionPrompt, question) + "\n\n")

	for _, result := range results {
		context.WriteString("Dateiname: " + result.ID + ":\n")
		context.WriteString(result.Content + "\n---\n")
	}
	context.WriteString("\nFrage: " + question + "\nAntwort:")

	if verbose {
		fmt.Println("\n\nEnd Prompt:\n" + context.String())
	}

	cmd := exec.Command("ollama", "run", "cyberwald/llama-3.1-sauerkrautlm-8b-instruct")
	cmd.Stdin = bytes.NewBufferString(context.String())

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ollama failed: %w\n%s", err, string(output))
	}

	return string(output), nil
}
