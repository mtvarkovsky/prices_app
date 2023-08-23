package testutils

import (
	"encoding/csv"
	"fmt"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"os"
	"time"
)

func GenerateTestData(lines int, targetDir string) {
	now := time.Now()
	fileName := fmt.Sprintf("%s/%d_test_prices.csv", targetDir, now.UnixNano())
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("can't create file %s: (%s)", fileName, err.Error())
	}
	r := rand.New(rand.NewSource(now.UnixNano()))
	writer := csv.NewWriter(file)
	for i := 0; i < lines; i++ {
		id, err := uuid.NewUUID()
		if err != nil {
			log.Fatalf("can't create id %s: (%s)", id, err.Error())
		}

		priceInt := r.Intn(20) * 111
		priceFract := r.Intn(20) * 111111
		price := fmt.Sprintf("%d.%d", priceInt, priceFract)

		expirationDate := time.Now().AddDate(0, 0, r.Intn(5))

		line := []string{id.String(), price, expirationDate.Format("2006-01-02 15:04:05 -0700 MST")}

		err = writer.Write(line)
		if err != nil {
			log.Fatalf("can't write line=%s to file %s: (%s)", line, fileName, err.Error())
		}
	}
	writer.Flush()
	if err := file.Close(); err != nil {
		log.Fatalf("can't close file %s: (%s)", fileName, err.Error())
	}
}
