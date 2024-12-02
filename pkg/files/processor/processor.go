//go:generate mockgen -source processor.go -destination repository_mock.go -package processor PricesRepo

package processor

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"prices/pkg/config"
	"prices/pkg/files"
	"prices/pkg/models"
	"sync"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

type (
	FileQueue interface {
		Data() (<-chan files.File, error)
	}

	PricesRepo interface {
		CreateMany(ctx context.Context, prices []*models.Price) error
		ImportFile(ctx context.Context, filePath string) error
	}

	V1 struct {
		ctx     context.Context
		wg      *sync.WaitGroup
		wgRead  *sync.WaitGroup
		wgWrite *sync.WaitGroup
		config  *config.FileProcessor
		data    chan []*models.Price
		files   FileQueue
		repo    PricesRepo
		logger  *zap.Logger
		stop    <-chan bool
	}
)

func NewProcessor(
	ctx context.Context,
	wg *sync.WaitGroup,
	config *config.FileProcessor,
	files FileQueue,
	repo PricesRepo,
	logger *zap.Logger,
	stop <-chan bool,
) *V1 {
	log := logger.Named("FileProcessor")
	wgRead := &sync.WaitGroup{}
	wgWrite := &sync.WaitGroup{}
	p := &V1{
		wg:      wg,
		wgRead:  wgRead,
		wgWrite: wgWrite,
		ctx:     ctx,
		config:  config,
		data:    make(chan []*models.Price, config.DataBatchQueueSize),
		files:   files,
		repo:    repo,
		logger:  log,
		stop:    stop,
	}
	return p
}

func (p *V1) Process() {
	p.wg.Add(1)
	if p.config.ImportByLines {
		p.ProcessLines()
	} else {
		p.ProcessFiles()
	}
	<-p.stop
	p.logger.Sugar().Info("stop processing files")
	p.wgRead.Wait()
	close(p.data)
	p.wgWrite.Wait()
	p.wg.Done()
}

func (p *V1) ProcessLines() {
	p.logger.Sugar().Info("start processing files by lines")
	p.wgRead.Add(1)
	go p.readFilesByLines()
	for i := 0; i < p.config.WorkersCount; i++ {
		p.wgWrite.Add(1)
		go p.saveLines()
	}
}

func (p *V1) saveLines() {
	defer p.wgWrite.Done()
	p.logger.Sugar().Info("start processing worker")
	for prices := range p.data {
		err := p.repo.CreateMany(p.ctx, prices)
		if err != nil {
			p.logger.Sugar().Errorf("worker unable to process data item: (%s)", err.Error())
		}
	}
	p.logger.Sugar().Info("stop processing worker")
}

func (p *V1) readFilesByLines() {
	defer p.wgRead.Done()
	p.logger.Sugar().Info("start reading files")
	data, err := p.files.Data()
	if err != nil {
		p.logger.Sugar().Infof("can't to get file queue data: (%s)", err.Error())
		return
	}
	for file := range data {
		p.wgRead.Add(1)
		go p.readFileByLines(file)
	}
	p.logger.Sugar().Info("stop reading files")
}

func (p *V1) readFileByLines(file files.File) {
	defer p.wgRead.Done()
	p.logger.Sugar().Infof("start reading file=%s", file)
	f, err := os.Open(file.Path)
	if err != nil {
		p.logger.Sugar().Errorf("can't open file=%s: (%s)", file, err.Error())
	}
	reader := csv.NewReader(f)
	var prices []*models.Price
	for {
		line, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				p.logger.Sugar().Infof("done reading file=%s", file)
				_ = f.Close()
				break
			}
			p.logger.Sugar().Errorf("can't read file=%s data: (%s)", file, err.Error())
			return
		}
		if price := p.toPrice(file.Path, line); price != nil {
			prices = append(prices, price)
			if len(prices) == p.config.DataBatchSize {
				p.logger.Sugar().Infof("send file=%s data batch to processing", file)
				p.data <- prices
				prices = nil
			}
		}
	}
}

func (p *V1) toPrice(path string, line []string) *models.Price {
	price := &models.Price{}
	if len(line) > 3 {
		p.logger.Sugar().Errorf("bad file=%s format, only 3 columns expected", path)
		return nil
	}

	id, priceData, expirationDate := line[0], line[1], line[2]

	price.ID = id
	parsedPrice, err := p.parsePriceData(priceData)
	if err != nil {
		p.logger.Sugar().Errorf("bad file=%s data, cant parse price: (%s)", path, err.Error())
		return nil
	}
	price.Price = *parsedPrice
	expDate, err := p.parseExpirationDate(expirationDate)
	if err != nil {
		p.logger.Sugar().Errorf("bad file=%s data, cant parse expirationDate: (%s)", path, err.Error())
		return nil
	}
	price.ExpirationDate = *expDate
	return price
}

func (p *V1) parsePriceData(priceData string) (*decimal.Decimal, error) {
	price, err := decimal.NewFromString(priceData)
	if err != nil {
		return nil, fmt.Errorf("can't parse price data=%s as a floating point number", priceData)
	}
	return &price, nil
}

func (p *V1) parseExpirationDate(expirationDate string) (*time.Time, error) {
	expDate, err := time.Parse("2006-01-02 15:04:05 -0700 MST", expirationDate)
	if err != nil {
		return nil, fmt.Errorf("can't parse expiration date=%s as a timestamp", expirationDate)
	}
	return &expDate, nil
}

func (p *V1) ProcessFiles() {
	p.logger.Sugar().Info("start processing files")
	for i := 0; i < p.config.WorkersCount; i++ {
		p.wgWrite.Add(1)
		go p.saveFiles()
	}
}

func (p *V1) saveFiles() {
	p.logger.Info("start save files worker")
	defer p.wgWrite.Done()
	data, err := p.files.Data()
	if err != nil {
		p.logger.Sugar().Infof("can't to get file queue data: (%s)", err.Error())
		return
	}
	for file := range data {
		p.logger.Sugar().Infof("save file=%s to storage", file)
		err := p.repo.ImportFile(p.ctx, file.Path)
		if err != nil {
			p.logger.Sugar().Errorf("worker unable to process file=%s: (%s)", file, err.Error())
		}
	}
	p.logger.Info("stop save files worker")
}
