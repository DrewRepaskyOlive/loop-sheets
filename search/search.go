package search

import (
	"fmt"
	"time"

	"github.com/DrewRepaskyOlive/loop-sheets/sheets"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	ldk "github.com/open-olive/loop-development-kit/ldk/go"
)

// Searcher wrapper for Bleve search
type Searcher struct {
	BleveIndex bleve.Index
	logger     *ldk.Logger
}

const (
	batchSize = 100
)

// each row, regardless of sheet, gets a unique auto-incrementing ID
var currentRowID = 0

func buildIndexMapping() (mapping.IndexMapping, error) {
	indexMapping := bleve.NewIndexMapping()

	// FIXME needs tuned
	// a generic reusable mapping for English text
	//englishTextFieldMapping := bleve.NewTextFieldMapping()
	//englishTextFieldMapping.Analyzer = en.AnalyzerName
	//englishTextFieldMapping.Analyzer = simple.Name

	// a generic reusable mapping for keyword text
	//keywordFieldMapping := bleve.NewTextFieldMapping()
	//keywordFieldMapping.Analyzer = keyword.Name

	indexMapping.DefaultAnalyzer = "en"
	//indexMapping.DefaultAnalyzer = keyword.Name
	//indexMapping.DefaultAnalyzer = simple.Name

	return indexMapping, nil
}

func New(logger *ldk.Logger, indexName string, fileContent *sheets.SheetContent) (*Searcher, error) {
	start := time.Now()
	indexMapping, err := buildIndexMapping()
	if err != nil {
		return &Searcher{}, err
	}

	index, err := bleve.NewMemOnly(indexMapping)
	// persist the index to disk
	//index, err := bleve.New(indexName, indexMapping)
	if err != nil {
		return &Searcher{}, err
	}
	index.SetName(indexName)

	count := 0
	batch := index.NewBatch()

	for i, r := range fileContent.Content {
		currentRowID++
		id := fmt.Sprintf("%d", currentRowID)

		if batch.Size() > batchSize {
			err := index.Batch(batch)
			if err != nil {
				return nil, fmt.Errorf("could not flush batch: %w", err)
			}
			count += batch.Size()
			batch = index.NewBatch()
			logger.Info(fmt.Sprintf("bumped bleve index batch size to %d", count))
		}

		err := batch.Index(id, r)
		if err != nil {
			return nil, fmt.Errorf("could not create search index for row %d (overall row %d): %w", i, currentRowID, err)
		}
	}

	if count > 0 {
		err = index.Batch(batch)
		if err != nil {
			return nil, fmt.Errorf("could not flush final batch: %w", err)
		}
	}

	duration := time.Since(start)
	logger.Info(fmt.Sprintf("Bleve search index took %v to build", duration))
	return &Searcher{
		BleveIndex: index,
		logger:     logger,
	}, nil
}

func (s *Searcher) DoSearch(criteria string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// string "contains" (exact) match
	query := bleve.NewMatchPhraseQuery(criteria)

	// other options:
	//query := bleve.NewQueryStringQuery(criteria)
	//query := bleve.NewMatchQuery(criteria)
	//query := bleve.NewTermQuery(criteria)

	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.Fields = []string{"*"} // return all fields

	// unlimited search result count
	size, err := s.BleveIndex.DocCount()
	if err != nil {
		return nil, fmt.Errorf("could not get search index count: %w", err)
	}
	searchRequest.Size = int(size)

	searchResult, err := s.BleveIndex.Search(searchRequest)

	if err != nil {
		return nil, err
	}

	for _, hit := range searchResult.Hits {
		match := hit.Fields
		results = append(results, match)
	}

	return results, nil

}
