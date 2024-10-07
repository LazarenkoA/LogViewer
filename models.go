package main

import (
	"context"
	"sync"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

//go:generate mockgen -source=$GOFILE -destination=./mock/mock.go

type IPages interface {
	tview.Primitive

	AddPage(name string, item tview.Primitive, resize, visible bool) *tview.Pages
	RemovePage(name string) *tview.Pages
}

type ITable interface {
	tview.Primitive

	SetSelectable(rows, columns bool) *tview.Table
	SetCell(row, column int, cell *tview.TableCell) *tview.Table
	ScrollToEnd() *tview.Table
	GetCell(row, column int) *tview.TableCell
	Clear() *tview.Table
	ScrollToBeginning() *tview.Table
	GetSelection() (row, column int)
	SetSelectedFunc(handler func(row, column int)) *tview.Table
	SetMouseCapture(capture func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse)) *tview.Box
	Select(row, column int) *tview.Table
	GetColumnCount() int
	GetRowCount() int
	GetBackgroundColor() tcell.Color
}

type tableView struct {
	sync.RWMutex

	app         *tview.Application
	table       ITable
	pages       IPages
	line        map[string]*tline
	in          chan string
	ctx         context.Context
	sortColumn  int
	csvFileName string
	lang        string
}
