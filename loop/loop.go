package loop

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/DrewRepaskyOlive/loop-sheets/search"
	"github.com/DrewRepaskyOlive/loop-sheets/sheets"

	ldk "github.com/open-olive/loop-development-kit/ldk/go"
)

const (
	loopName        = "loop-sheets"
	searchIndexName = "loop-sheets-index"
	csvFilePath     = "/Users/drewrepasky/Documents/JIRA.csv"
)

type Loop struct {
	ctx      context.Context
	cancel   context.CancelFunc
	logger   *ldk.Logger
	sidekick ldk.Sidekick
	searcher *search.Searcher
}

func Serve() error {
	log := ldk.NewLogger(loopName)
	loop, err := NewLoop(log)
	if err != nil {
		return err
	}
	ldk.ServeLoopPlugin(log, loop)
	return nil
}

func NewLoop(logger *ldk.Logger) (*Loop, error) {
	logger.Trace("NewLoop called: " + loopName)
	return &Loop{
		logger: logger,
	}, nil
}

func (l *Loop) SendWhisper(label string, markdown string) {
	whisper := ldk.WhisperContentMarkdown{
		Label:    label,
		Markdown: markdown,
	}
	go func() {
		err := l.sidekick.Whisper().Markdown(l.ctx, &whisper)
		if err != nil {
			l.logger.Error("failed to emit whisper", "error", err)
		}
	}()
}

func (l *Loop) listener(searchCriteria string, err error) {
	if err != nil {
		l.logger.Error("received error in callback", "error", err)
		return
	}

	// double quotes cause blevesearch to error, which are sent in when copying from Excel fields
	searchCriteria = strings.ReplaceAll(searchCriteria, "\"", "")

	results, err := l.searcher.DoSearch(searchCriteria)
	if err != nil {
		l.logger.Error("bleve search failed", "error", err)
		return
	}

	l.logger.Info("searchCriteria", searchCriteria)
	l.logger.Info("results", results)

	for _, result := range results {
		label := fmt.Sprintf("matched %q", searchCriteria)
		markdown := fmt.Sprintf("%+v", result)
		l.SendWhisper(label, markdown)
	}
}

func (l *Loop) LoopStart(sidekick ldk.Sidekick) error {
	l.logger.Trace("starting " + loopName)
	l.ctx, l.cancel = context.WithCancel(context.Background())
	l.sidekick = sidekick

	l.buildSearchIndex(csvFilePath)

	err := sidekick.UI().ListenGlobalSearch(l.ctx, l.listener)
	if err != nil {
		return err
	}

	err = sidekick.UI().ListenSearchbar(l.ctx, l.listener)
	if err != nil {
		return err
	}
	return nil
}

func (l *Loop) buildSearchIndex(csvFile string) error {
	sheetContent, err := sheets.ReadCSV(csvFile)
	if err != nil {
		return err
	}

	log.Printf("%+v", sheetContent.Content[0])
	log.Printf("headers %+v", sheetContent.Headers)
	log.Println(len(sheetContent.Headers))
	log.Println(len(sheetContent.Content))

	searcher, err := search.New(l.logger, searchIndexName, sheetContent)
	if err != nil {
		return err
	}
	l.searcher = searcher
	return nil
}

func (l *Loop) LoopStop() error {
	l.logger.Trace("stopping " + loopName)
	l.cancel()
	return nil
}
