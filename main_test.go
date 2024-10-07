package main

import (
	mock_main "LogViewer/mock"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func Test_SaveToCSV(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	table := mock_main.NewMockITable(c)
	table.EXPECT().GetColumnCount().Return(3).AnyTimes()
	table.EXPECT().GetRowCount().Return(3).AnyTimes()
	table.EXPECT().GetCell(gomock.Any(), gomock.Any()).Return(&tview.TableCell{Text: "test"}).AnyTimes()

	newView := new(tableView).Construct(context.Background())
	newView.lang = "ru"
	newView.table = table

	t.Run("empty csvFileName", func(t *testing.T) {
		newView.exportToCSV()
		_, err := os.Stat(newView.csvFileName + ".csv")
		assert.True(t, errors.Is(err, os.ErrNotExist))
	})
	t.Run("pass", func(t *testing.T) {
		newView.csvFileName = "test"
		newView.exportToCSV()

		time.Sleep(time.Second)

		f, err := os.Open(newView.csvFileName)
		assert.False(t, errors.Is(err, os.ErrNotExist))

		data, _ := io.ReadAll(f)
		assert.Equal(t, strings.ReplaceAll(string(data), "\n", ""), `test,testtest,testtest,test`)
		f.Close()
	})
}
