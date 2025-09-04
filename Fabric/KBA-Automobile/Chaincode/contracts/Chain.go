package contracts

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"crypto/tls"
)

type SmartContract struct {
	contractapi.Contract
}

type Seller struct {
	SellerID         string  `json:"sellerId"`
	NFTID            string  `json:"nftId"`
	IPFSHash         string  `json:"ipfsHash"`
	DecryptionKey    string  `json:"decryptionKey"`
	BasePrice        float64 `json:"basePrice"`
	MedicalSpecialty string  `json:"medicalSpecialty"`
	OriginalFile     string  `json:"originalFile"`
}

type Buyer struct {
	NFTID    string  `json:"nftId"`
	BuyerID  string  `json:"buyerId"`
	Bid      float64 `json:"bid"`
	Category string  `json:"category"`
}

// Vickrey auction result (second-price with reserve/BasePrice).
type AuctionResult struct {
	NFTID            string  `json:"nftId"`
	SellerID         string  `json:"sellerId"`
	WinningBuyerID   string  `json:"winningBuyerId"`
	WinningBid       float64 `json:"winningBid"`       // highest bid submitted by winner
	PricePaid        float64 `json:"pricePaid"`        // clearing price: max(BasePrice, second-highest bid)
	SecondHighestBid float64 `json:"secondHighestBid"` // 0 if only one valid bid
	BasePrice        float64 `json:"basePrice"`
	NumBids          int     `json:"numBids"`
	MedicalSpecialty string  `json:"medicalSpecialty"`
}

type AuctionStatistics struct {
	TotalNFTsAvailable    int     `json:"totalNFTsAvailable"`
	SuccessfulMatches     int     `json:"successfulMatches"`
	SuccessRate           float64 `json:"successRate"`
	NFTsWithBuyers        int     `json:"nftsWithBuyers"`
	NFTsWithoutBuyers     int     `json:"nftsWithoutBuyers"`
	TotalRevenueGenerated float64 `json:"totalRevenueGenerated"` // sum of PricePaid
	AverageSalePrice      float64 `json:"averageSalePrice"`      // avg of PricePaid
	TotalBidsReceived     int     `json:"totalBidsReceived"`
	AverageBidsPerNFT     float64 `json:"averageBidsPerNFT"`
}

type AuctionResponse struct {
	Results    []AuctionResult   `json:"results"`
	Statistics AuctionStatistics `json:"statistics"`
	Message    string            `json:"message"`
}

func (s *SmartContract) MapBuyersToSellers(ctx contractapi.TransactionContextInterface, sellersURL string, buyersURL string) (*AuctionResponse, error) {
	sellersData, err := downloadTSV(sellersURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download sellers data: %s", err)
	}

	buyersData, err := downloadTSV(buyersURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download buyers data: %s", err)
	}

	sellers, err := parseSellers(sellersData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sellers data: %s", err)
	}

	buyersByNFT, err := parseBuyers(buyersData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse buyers data: %s", err)
	}

	// Sort NFT IDs to ensure deterministic iteration
	var nftIDs []string
	for nftId := range sellers {
		nftIDs = append(nftIDs, nftId)
	}
	sort.Strings(nftIDs)

	results := []AuctionResult{}

	// Initialize statistics tracking variables
	totalNFTs := len(sellers)
	successfulMatches := 0
	nftsWithBuyers := 0
	nftsWithoutBuyers := 0
	totalRevenue := 0.0
	totalBids := 0

	// Count total bids
	for _, buyers := range buyersByNFT {
		totalBids += len(buyers)
	}

	// Iterate in sorted order for deterministic behavior
	for _, nftId := range nftIDs {
		seller := sellers[nftId]
		buyers, exists := buyersByNFT[nftId]

		if !exists || len(buyers) == 0 {
			nftsWithoutBuyers++
			continue
		}

		nftsWithBuyers++

		// Sort buyers by bid desc, break ties by BuyerID asc for determinism
		sort.Slice(buyers, func(i, j int) bool {
			if buyers[i].Bid == buyers[j].Bid {
				return buyers[i].BuyerID < buyers[j].BuyerID
			}
			return buyers[i].Bid > buyers[j].Bid
		})

		// Highest bid (candidate winner)
		winner := buyers[0]
		winningBid := winner.Bid

		// Reserve check: if the highest bid doesn't meet BasePrice, no sale
		if winningBid < seller.BasePrice {
			// winner fails to meet reserve; NFT remains unsold
			continue
		}

		// Compute second-highest bid if present
		secondHighest := 0.0
		if len(buyers) > 1 {
			secondHighest = buyers[1].Bid
		}

		// Vickrey price with reserve: pay max(BasePrice, second-highest bid)
		pricePaid := seller.BasePrice
		if secondHighest > pricePaid {
			pricePaid = secondHighest
		}

		// Record successful match
		successfulMatches++
		totalRevenue += pricePaid

		result := AuctionResult{
			NFTID:            nftId,
			SellerID:         seller.SellerID,
			WinningBuyerID:   winner.BuyerID,
			WinningBid:       winningBid,
			PricePaid:        pricePaid,
			SecondHighestBid: secondHighest,
			BasePrice:        seller.BasePrice,
			NumBids:          len(buyers),
			MedicalSpecialty: seller.MedicalSpecialty,
		}
		results = append(results, result)

		resultJSON, _ := json.Marshal(result)
		if err := ctx.GetStub().PutState(nftId, resultJSON); err != nil {
			return nil, fmt.Errorf("failed to put state for %s: %s", nftId, err)
		}
	}

	// Calculate statistics
	var successRate float64
	if totalNFTs > 0 {
		successRate = (float64(successfulMatches) / float64(totalNFTs)) * 100
	}

	var averageSalePrice float64
	if successfulMatches > 0 {
		averageSalePrice = totalRevenue / float64(successfulMatches)
	}

	var averageBidsPerNFT float64
	if totalNFTs > 0 {
		averageBidsPerNFT = float64(totalBids) / float64(totalNFTs)
	}

	statistics := AuctionStatistics{
		TotalNFTsAvailable:    totalNFTs,
		SuccessfulMatches:     successfulMatches,
		SuccessRate:           successRate,
		NFTsWithBuyers:        nftsWithBuyers,
		NFTsWithoutBuyers:     nftsWithoutBuyers,
		TotalRevenueGenerated: totalRevenue,
		AverageSalePrice:      averageSalePrice,
		TotalBidsReceived:     totalBids,
		AverageBidsPerNFT:     averageBidsPerNFT,
	}

	// Store statistics in blockchain state
	statisticsJSON, _ := json.Marshal(statistics)
	if err := ctx.GetStub().PutState("auction_statistics", statisticsJSON); err != nil {
		return nil, fmt.Errorf("failed to store auction statistics: %s", err)
	}

	// Create response message
	message := fmt.Sprintf(
		"Vickrey auction completed! %d of %d NFTs sold. Total revenue (pricePaid) = %.2f.",
		successfulMatches, totalNFTs, totalRevenue,
	)

	response := &AuctionResponse{
		Results:    results,
		Statistics: statistics,
		Message:    message,
	}

	return response, nil
}

func downloadTSV(url string) ([][]string, error) {
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{Transport: tr}

    resp, err := client.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    reader := csv.NewReader(resp.Body)
    reader.Comma = '\t'
    reader.FieldsPerRecord = -1

    data, err := reader.ReadAll()
    if err != nil {
        return nil, err
    }

    if len(data) <= 1 {
        return nil, fmt.Errorf("no data rows found (only header)")
    }

    return data[1:], nil
}

func parseSellers(data [][]string) (map[string]Seller, error) {
	sellers := make(map[string]Seller)
	for i, row := range data {
		if len(row) < 7 {
			return nil, fmt.Errorf("invalid seller data at row %d: expected 7 columns, got %d", i+2, len(row))
		}

		basePrice, err := strconv.ParseFloat(row[4], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid base price at row %d: %s", i+2, err)
		}

		seller := Seller{
			SellerID:         row[0],
			NFTID:            row[1],
			IPFSHash:         row[2],
			DecryptionKey:    row[3],
			BasePrice:        basePrice,
			MedicalSpecialty: row[5],
			OriginalFile:     row[6],
		}
		sellers[seller.NFTID] = seller
	}
	return sellers, nil
}

func parseBuyers(data [][]string) (map[string][]Buyer, error) {
	buyers := make(map[string][]Buyer)
	for i, row := range data {
		if len(row) < 4 {
			return nil, fmt.Errorf("invalid buyer data at row %d: expected 4 columns, got %d", i+2, len(row))
		}

		bid, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid bid at row %d: %s", i+2, err)
		}

		buyer := Buyer{
			NFTID:    row[0],
			BuyerID:  row[1],
			Bid:      bid,
			Category: row[3],
		}
		buyers[buyer.NFTID] = append(buyers[buyer.NFTID], buyer)
	}
	return buyers, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating contracts chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting contracts chaincode: %s", err.Error())
	}
}
