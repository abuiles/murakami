package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func getData(url string, dest *gin.H) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, dest)
	if err != nil {
		return err
	}

	return nil
}

func getComparison() gin.H {
	var wg sync.WaitGroup

	transferWise := gin.H{}
	stellarDEX := gin.H{}

	wg.Add(1)
	go func() {
		log.Println("fetching data from Transferwise")
		getData(
			"https://api.transferwise.com/v3/comparisons?sourceCurrency=USD&targetCurrency=BRL&sendAmount=15",
			&transferWise,
		)
		log.Println("data fetched from Transferwise")
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		log.Println("fetching data from Stellar DEX")
		getData(
			"https://horizon.stellar.org/paths/strict-send?&source_amount=15&source_asset_type=credit_alphanum4&source_asset_code=USD&source_asset_issuer=GDUKMGUGDZQK6YHYA5Z6AY2G4XDSZPSZ3SW5UN3ARVMO6QSRDWP5YLEX&destination_assets=BRL:GDVKY2GU2DRXWTBEYJJWSFXIGBZV6AZNBVVSUHEPZI54LIS6BA7DVVSP",
			&stellarDEX,
		)
		log.Println("data fetched from Stellar DEX")
		wg.Done()
	}()

	wg.Wait()

	response := gin.H{
		"transferWise": transferWise,
		"stellarDEX":   stellarDEX,
	}

	return response
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET"},
	}))

	api := r.Group("/api/v1")

	api.GET("/comparisons", func(c *gin.Context) {
		c.JSON(200, getComparison())
	})

	return r
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	r := setupRouter()

	s := &http.Server{
		Addr:        ":" + port,
		Handler:     r,
		ReadTimeout: 5 * time.Second,
	}

	s.ListenAndServe()
}
