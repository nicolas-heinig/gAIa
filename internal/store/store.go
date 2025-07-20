package store

import (
	"context"
	"fmt"
	"gaia/internal/chunker"
	"gaia/internal/parser"
	"log"
	"sort"

	"github.com/philippgille/chromem-go"
)

type Store struct {
	smallChunks *chromem.Collection
	bigChunks   *chromem.Collection
	verbose     bool
}

func NewStore(path string, verbose bool) (*Store, error) {
	db, err := chromem.NewPersistentDB(path, true)
	if err != nil {
		return nil, err
	}

	embeddingFunc := chromem.NewEmbeddingFuncOllama("jina/jina-embeddings-v2-base-de", "")

	smallChunks, err := db.GetOrCreateCollection("small-chunks", nil, embeddingFunc)

	if err != nil {
		return nil, err
	}

	bigChunks, err := db.GetOrCreateCollection("big-chunks", nil, embeddingFunc)

	if err != nil {
		return nil, err
	}

	return &Store{
		smallChunks: smallChunks,
		bigChunks:   bigChunks,
		verbose:     verbose,
	}, nil
}

func (s *Store) StoreDocuments(ctx context.Context, docs []parser.ParsedDocument) error {
	if !s.verbose {
		fmt.Print("Storing documents")
	}

	for _, doc := range docs {
		err := s.storeChunks(ctx, doc, 512, s.smallChunks)

		if err != nil {
			return err
		}

		err = s.storeChunks(ctx, doc, 2048, s.bigChunks)

		if err != nil {
			return err
		}

		if !s.verbose {
			fmt.Print(".")
		}
	}

	if !s.verbose {
		fmt.Println()
	}

	return nil
}

func (s *Store) storeChunks(ctx context.Context, doc parser.ParsedDocument, chunkSize int, collection *chromem.Collection) error {
	metadata := map[string]string{
		"filename": doc.Filename,
	}

	for _, category := range doc.Categories {
		metadata[category] = "true"
	}

	if s.verbose {
		fmt.Printf("Storing chunks (size %d) of %s", chunkSize, doc.Filename)
	}

	chunks := chunker.ChunkText(doc.Text, chunkSize)
	for i, chunk := range chunks {
		id := fmt.Sprintf("%s_%d", doc.Filename, i)

		err := collection.AddDocument(ctx, chromem.Document{
			ID:       id,
			Content:  chunk,
			Metadata: metadata,
		})

		if err != nil {
			log.Printf("Failed to add document %s: %v", id, err)
		} else {
			if s.verbose {
				fmt.Print(".")
			}
		}
	}

	if s.verbose {
		fmt.Println(" Done")
	}

	return nil
}

func (s *Store) Query(ctx context.Context, question string, limit int) ([]chromem.Result, error) {
	smallResults, err := s.querySmallChunks(ctx, question)

	if err != nil {
		return nil, fmt.Errorf("failed to query small chunks: %w", err)
	}

	bigResults, err := s.queryBigChunks(ctx, question, smallResults, limit)
	if err != nil {
		return nil, err
	}

	return bigResults, nil
}

// query the samll chunks collection for results
// and filter them based on similarity score
func (s *Store) querySmallChunks(ctx context.Context, question string) ([]chromem.Result, error) {
	smallResults, err := s.smallChunks.Query(ctx, question, 25, nil, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to query small chunks: %w", err)
	}

	var filteredResults []chromem.Result
	var filteredOutResults []chromem.Result

	for _, res := range smallResults {
		if res.Similarity > 0.4 {
			filteredResults = append(filteredResults, res)
		} else {
			filteredOutResults = append(filteredOutResults, res)
		}
	}

	if s.verbose {
		logResults("Filtered small chunks results (similarity > 0.4):", filteredResults)
		logResults("Filtered out small chunks results (similarity <= 0.4):", filteredOutResults)
	}

	return filteredResults, nil
}

// query the big chunks collection for results only only for files that were found in the small chunks results
// sort all of them by similarity and pick the first limit results
func (s *Store) queryBigChunks(ctx context.Context, question string, smallResults []chromem.Result, limit int) ([]chromem.Result, error) {
	fileNames := make(map[string]struct{})

	for _, res := range smallResults {
		fileNames[res.Metadata["filename"]] = struct{}{}
	}

	var bigResults []chromem.Result

	for file := range fileNames {
		results, err := s.bigChunks.Query(ctx, question, 5, map[string]string{"filename": file}, nil)

		if err != nil {
			return []chromem.Result{}, fmt.Errorf("failed to query big chunks for file %s: %w", file, err)
		}

		bigResults = append(bigResults, results...)
	}

	sort.SliceStable(bigResults, func(i, j int) bool {
		return bigResults[i].Similarity > bigResults[j].Similarity
	})

	if s.verbose {
		logResults("Big chunks results:", bigResults)
	}

	if len(bigResults) > limit {
		bigResults = bigResults[:limit]
	}

	if s.verbose {
		logResults(fmt.Sprintf("Final %d results:", limit), bigResults)
	}

	return bigResults, nil
}

func logResults(msg string, results []chromem.Result) {
	fmt.Println(msg)
	for _, res := range results {
		fmt.Println("    " + formatResult(res))
	}
	fmt.Println()
}

func formatResult(result chromem.Result) string {
	categories := ""

	for key := range result.Metadata {
		if key != "filename" && key != "id" {
			categories += key + ", "
		}
	}

	return fmt.Sprintf("ID: %s; Categories: %s; Score: %.4f", result.ID, categories, result.Similarity)
}
