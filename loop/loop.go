package loop

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/DrewRepaskyOlive/loop-sheets/search"
	"github.com/DrewRepaskyOlive/loop-sheets/sheets"

	ldk "github.com/open-olive/loop-development-kit/ldk/go"
)

const (
	loopName        = "loop-sheets"
	searchIndexName = "loop-sheets-index"
	whisperLabel    = "Spreadsheet Search"
)

type Loop struct {
	ctx             context.Context
	cancel          context.CancelFunc
	logger          *ldk.Logger
	sidekick        ldk.Sidekick
	searcher        *search.Searcher
	myDocumentsPath string
	csvPaths        []string
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

	l.logger.Debug("results count matched in sheets", len(results))

	for _, result := range results {
		label := fmt.Sprintf("%s matched %q", whisperLabel, searchCriteria)

		markdown := ""
		for key, value := range result {
			markdown += fmt.Sprintf("**%s**: %s\n  \n", key, value)
		}
		l.SendWhisper(label, markdown)
	}
}

// TODO search criteria should be dynamically read, not hardcoded
const initWhisper = `# I found these spreadsheets and will match their contents when you search within Olive Helps: 
  
%s  

For instance, try searching on SIDE-1752
`

const initWhisperNoFiles = "Please add .csv files to `%s` to make their contents searchable within OliveHelps"

func (l *Loop) LoopStart(sidekick ldk.Sidekick) error {
	l.logger.Trace("starting " + loopName)
	l.ctx, l.cancel = context.WithCancel(context.Background())
	l.sidekick = sidekick

	var err error
	l.myDocumentsPath, err = sheets.GetMyDocumentsPath()
	if err != nil {
		return err
	}

	l.csvPaths, err = sheets.GetFilesByExtension(l.myDocumentsPath, ".csv")
	if err != nil {
		return err
	}

	if len(l.csvPaths) == 0 {
		message := fmt.Sprintf(initWhisperNoFiles, l.myDocumentsPath)
		l.SendWhisper(whisperLabel, message)
		return nil
	}

	l.buildSearchIndex(l.csvPaths)

	err = sidekick.UI().ListenGlobalSearch(l.ctx, l.listener)
	if err != nil {
		return err
	}

	err = sidekick.UI().ListenSearchbar(l.ctx, l.listener)
	if err != nil {
		return err
	}

	err = sidekick.Clipboard().Listen(l.ctx, l.listener)
	if err != nil {
		return err
	}

	csvList := ""
	for i, file := range l.csvPaths {
		csvList += fmt.Sprintf("%d. %s\n  ", i+1, filepath.Base(file))
	}

	markdown := fmt.Sprintf(initWhisper, csvList)
	l.SendWhisper(whisperLabel, markdown)
	return nil
}

func (l *Loop) buildSearchIndex(csvFiles []string) error {
	searcher, err := search.New(l.logger, searchIndexName)
	if err != nil {
		return err
	}
	l.searcher = searcher

	var sheetContents []*sheets.SheetContent

	for _, csvFile := range csvFiles {
		sheetContent, err := sheets.ReadCSV(csvFile)
		if err != nil {
			return err
		}
		l.logger.Info("parsed sheet", "filePath", csvFile, "headers", sheetContent.Headers, "rowCount", len(sheetContent.Content))
		sheetContents = append(sheetContents, sheetContent)
	}

	l.searcher.IndexSheets(sheetContents)

	return nil
}

func (l *Loop) LoopStop() error {
	l.logger.Trace("stopping " + loopName)
	l.cancel()

	if l.searcher != nil {
		err := l.searcher.BleveIndex.Close()
		if err != nil {
			return fmt.Errorf("failed to close BleveIndex gracefully: %w", err)
		}
	}
	return nil
}
