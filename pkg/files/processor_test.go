package files

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"os"
	"prices/pkg/config"
	"prices/pkg/models"
	"prices/pkg/repository"
	"prices/pkg/testutils"
	"sync"
	"testing"
	"time"
)

func newTestLineProcessor(t *testing.T) (*processor, chan bool) {
	dir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)

	files := NewFileQueueInMem(0)

	wg := &sync.WaitGroup{}
	log := zap.NewNop()

	cfg := &config.FileProcessor{
		FilesDir:           dir,
		WorkersCount:       2,
		DataBatchSize:      1,
		DataBatchQueueSize: 1,
		ImportByLines:      true,
	}

	ctx := context.Background()

	ctrl := gomock.NewController(t)

	repo := repository.NewMockPrices(ctrl)

	stop := make(chan bool)

	prcssr := NewProcessor(ctx, wg, cfg, files, repo, log, stop)

	p := prcssr.(*processor)
	return p, stop
}

func newTestFileProcessor(t *testing.T) (*processor, chan bool) {
	dir, err := os.MkdirTemp("", "")
	assert.NoError(t, err)

	files := NewFileQueueInMem(0)

	wg := &sync.WaitGroup{}
	log := zap.NewNop()

	cfg := &config.FileProcessor{
		FilesDir:           dir,
		WorkersCount:       2,
		DataBatchSize:      1,
		DataBatchQueueSize: 1,
		ImportByLines:      false,
	}

	ctx := context.Background()

	ctrl := gomock.NewController(t)

	repo := repository.NewMockPrices(ctrl)

	stop := make(chan bool)

	prcssr := NewProcessor(ctx, wg, cfg, files, repo, log, stop)

	p := prcssr.(*processor)
	return p, stop
}

func TestProcessor_ParsePriceData(t *testing.T) {
	prcssr, _ := newTestFileProcessor(t)

	priceData := "3333.3333"

	expected, err := decimal.NewFromString(priceData)
	assert.NoError(t, err)

	res, err := prcssr.parsePriceData(priceData)
	assert.NoError(t, err)

	assert.Equal(t, &expected, res)
}

func TestProcessor_ParseExpirationDate(t *testing.T) {
	prcssr, _ := newTestFileProcessor(t)

	expirationDate := "2023-08-25 10:42:33 +0200 CEST"

	expected, err := time.Parse("2006-01-02 15:04:05 -0700 MST", expirationDate)
	assert.NoError(t, err)

	res, err := prcssr.parseExpirationDate(expirationDate)
	assert.NoError(t, err)

	assert.Equal(t, &expected, res)
}

func TestProcessor_ToPrice(t *testing.T) {
	prcssr, _ := newTestFileProcessor(t)

	line := []string{
		"d65d3cba-40c7-11ee-afc6-a45e60d0762b",
		"666.2111109",
		"2023-08-23 10:42:33 +0200 CEST",
	}
	path := "test"

	price, err := decimal.NewFromString(line[1])
	assert.NoError(t, err)

	expirationDate, err := time.Parse("2006-01-02 15:04:05 -0700 MST", line[2])
	assert.NoError(t, err)

	expected := &models.Price{
		ID:             line[0],
		Price:          price,
		ExpirationDate: expirationDate,
	}

	res := prcssr.toPrice(path, line)
	assert.Equal(t, *expected, *res)
}

func TestProcessor_ReadFileByLines(t *testing.T) {
	prcssr, _ := newTestLineProcessor(t)
	dir := prcssr.config.FilesDir

	data := prcssr.data

	testutils.GenerateTestData(2, dir)

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	entry := entries[0]

	file := File{
		Path: fmt.Sprintf("%s/%s", dir, entry.Name()),
		Name: entry.Name(),
	}

	f, err := os.Open(file.Path)
	assert.NoError(t, err)

	reader := csv.NewReader(f)
	lines, err := reader.ReadAll()
	assert.NoError(t, err)

	f.Close()

	prcssr.wgRead.Add(1)
	go prcssr.readFileByLines(file)

	line1 := prcssr.toPrice(file.Path, lines[0])
	assert.Equal(t, []*models.Price{line1}, <-data)

	line2 := prcssr.toPrice(file.Path, lines[1])
	assert.Equal(t, []*models.Price{line2}, <-data)
}

func TestProcessor_SaveLines(t *testing.T) {
	prcssr, _ := newTestLineProcessor(t)

	data := prcssr.data
	repo := prcssr.repo.(*repository.MockPrices)

	priceData1, err := decimal.NewFromString("2109.555555")
	assert.NoError(t, err)
	expirationDate1, err := time.Parse("2006-01-02 15:04:05 -0700 MST", "2023-08-23 16:32:48 +0200 CEST")
	assert.NoError(t, err)
	price1 := &models.Price{
		ID:             "999f6c3c-402f-11ee-9150-a45e60d0762b",
		Price:          priceData1,
		ExpirationDate: expirationDate1,
	}

	priceData2, err := decimal.NewFromString("1776.1777776")
	assert.NoError(t, err)
	expirationDate2, err := time.Parse("2006-01-02 15:04:05 -0700 MST", "2023-08-25 16:32:48 +0200 CEST")
	assert.NoError(t, err)
	price2 := &models.Price{
		ID:             "999f6c46-402f-11ee-9150-a45e60d0762b",
		Price:          priceData2,
		ExpirationDate: expirationDate2,
	}

	prices := [][]*models.Price{
		{
			price1,
		},
		{
			price2,
		},
	}

	prcssr.wgWrite.Add(1)
	go prcssr.saveLines()

	repo.EXPECT().CreateMany(prcssr.ctx, prices[0]).Return(nil)
	repo.EXPECT().CreateMany(prcssr.ctx, prices[1]).Return(nil)

	data <- prices[0]
	data <- prices[1]

	close(data)

	prcssr.wgWrite.Wait()
}

func TestProcessor_SaveFiles(t *testing.T) {
	prcssr, _ := newTestFileProcessor(t)

	files := prcssr.files
	repo := prcssr.repo.(*repository.MockPrices)

	file1 := File{Path: "test1.csv"}
	file2 := File{Path: "test2.csv"}

	prcssr.wgWrite.Add(1)
	go prcssr.saveFiles()

	repo.EXPECT().ImportFile(prcssr.ctx, file1.Path).Return(nil)
	repo.EXPECT().ImportFile(prcssr.ctx, file2.Path).Return(nil)

	err := files.Put(file1)
	assert.NoError(t, err)
	err = files.Put(file2)
	assert.NoError(t, err)

	err = files.Close()
	assert.NoError(t, err)

	prcssr.wgWrite.Wait()
}

func TestProcessor_ProcessLines(t *testing.T) {
	prcssr, stop := newTestLineProcessor(t)

	wg := prcssr.wg

	dir := prcssr.config.FilesDir

	files := prcssr.files
	repo := prcssr.repo.(*repository.MockPrices)

	testutils.GenerateTestData(2, dir)

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	entry := entries[0]

	file := File{
		Path: fmt.Sprintf("%s/%s", dir, entry.Name()),
		Name: entry.Name(),
	}

	f, err := os.Open(file.Path)
	assert.NoError(t, err)

	reader := csv.NewReader(f)
	lines, err := reader.ReadAll()
	assert.NoError(t, err)

	f.Close()

	prices := [][]*models.Price{
		{
			prcssr.toPrice(file.Path, lines[0]),
		},
		{
			prcssr.toPrice(file.Path, lines[1]),
		},
	}

	go prcssr.Process()

	repo.EXPECT().CreateMany(prcssr.ctx, prices[0]).Return(nil)
	repo.EXPECT().CreateMany(prcssr.ctx, prices[1]).Return(nil)

	err = files.Put(file)
	assert.NoError(t, err)

	err = files.Close()
	assert.NoError(t, err)
	stop <- true
	wg.Wait()
}

func TestProcessor_ProcessFiles(t *testing.T) {
	prcssr, stop := newTestFileProcessor(t)

	wg := prcssr.wg

	dir := prcssr.config.FilesDir

	files := prcssr.files
	repo := prcssr.repo.(*repository.MockPrices)

	testutils.GenerateTestData(2, dir)

	entries, err := os.ReadDir(dir)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	entry := entries[0]

	file := File{
		Path: fmt.Sprintf("%s/%s", dir, entry.Name()),
		Name: entry.Name(),
	}

	go prcssr.Process()

	repo.EXPECT().ImportFile(prcssr.ctx, file.Path).Return(nil)

	err = files.Put(file)
	assert.NoError(t, err)

	err = files.Close()
	stop <- true
	wg.Wait()
}
