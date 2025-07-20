package main

import (
	"context"
	"fmt"
	"gaia/internal/llm"
	"gaia/internal/parser"
	"gaia/internal/store"
	"log"
	"time"

	"github.com/briandowns/spinner"
	_ "github.com/sanity-io/litter"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "gAIa"}
	var verbose bool

	var indexCmd = &cobra.Command{
		Use:   "index [path]",
		Short: "Index documents in the given folder",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Indexing documents in:", args[0])
			docs, err := parser.ParseDocuments(args[0])

			if err != nil {
				fmt.Println("Error parsing documents:", err)
				return
			}

			store, err := store.NewStore("./vectors", verbose)

			ctx := context.Background()

			store.StoreDocuments(ctx, docs)
		},
	}

	indexCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Be verbose with the logging")

	var askCmd = &cobra.Command{
		Use:   "ask [question]",
		Short: "Ask a question about your LARP lore",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			question := args[0]

			store, err := store.NewStore("./vectors", verbose)

			if err != nil {
				log.Fatal("Error creating store:", err)
			}

			ctx := context.Background()

			results, err := store.Query(ctx, question, 5)

			if err != nil {
				log.Fatal("Error querying store:", err)
			}

			s2 := spinner.New(spinner.CharSets[31], 100*time.Millisecond)
			s2.Start()

			answer, err := llm.AskQuestion(question, results, verbose)

			if err != nil {
				log.Fatal("Error asking question:", err)
			}

			s2.Stop()

			fmt.Println("\n" + answer)
		},
	}

	askCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Be verbose with the logging")

	rootCmd.AddCommand(indexCmd, askCmd)
	rootCmd.Execute()
}
