package contracts

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// CarContract contract for managing CRUD for Car
type CarContract struct {
	contractapi.Contract
}
type PaginatedQueryResult struct {
	Records             []*Car `json:"records"`
	FetchedRecordsCount int32  `json:"fetchedRecordsCount"`
	Bookmark            string `json:"bookmark"`
}

type HistoryQueryResult struct {
	Record    *Car   `json:"record"`
	TxId      string `json:"txId"`
	Timestamp string `json:"timestamp"`
	IsDelete  bool   `json:"isDelete"`
}

type Car struct {
	AssetType         string `json:"assetType"`
	CarId             string `json:"carId"`
	Color             string `json:"color"`
	DateOfManufacture string `json:"dateOfManufacture"`
	Make              string `json:"make"`
	Model             string `json:"model"`
	OwnedBy           string `json:"ownedBy"`
	Status            string `json:"status"`
}

// CarExists returns true when asset with given ID exists in world state
func (c *CarContract) CarExists(ctx contractapi.TransactionContextInterface, carID string) (bool, error) {
	data, err := ctx.GetStub().GetState(carID)

	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return data != nil, nil
}

// CreateCar creates a new instance of Car
func (c *CarContract) CreateCar(ctx contractapi.TransactionContextInterface, carID string, make string, model string, color string, manufacturerName string, dateOfManufacture string) (string, error) {
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()

	if err != nil {
		return "", fmt.Errorf("could not fetch client identity. %s", err)
	}

	if clientOrgID == "ManufacturerMSP" {
	// if clientOrgID == "Org1MSP" {
		//if clientOrgID == "manufacturer-auto-com" {

		exists, err := c.CarExists(ctx, carID)
		if err != nil {
			return "", fmt.Errorf("could not fetch the details from world state.%s", err)
		} else if exists {
			return "", fmt.Errorf("the car, %s already exists", carID)
		}

		car := Car{
			AssetType:         "car",
			CarId:             carID,
			Color:             color,
			DateOfManufacture: dateOfManufacture,
			Make:              make,
			Model:             model,
			OwnedBy:           manufacturerName,
			Status:            "In Factory",
		}
		fmt.Println("Create car data ======= ", car)
		bytes, _ := json.Marshal(car)

		err = ctx.GetStub().PutState(carID, bytes)
		if err != nil {
			return "", fmt.Errorf("could not create car. %s", err)
		} else {
			return fmt.Sprintf("successfully added car %v", carID), nil
		}

	} else {
		return "", fmt.Errorf("user under following MSPID: %v can't perform this action", clientOrgID)
	}
}

// ReadCar retrieves an instance of Car from the world state
func (c *CarContract) ReadCar(ctx contractapi.TransactionContextInterface, carID string) (*Car, error) {

	bytes, err := ctx.GetStub().GetState(carID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if bytes == nil {
		return nil, fmt.Errorf("the car %s does not exist", carID)
	}

	var car Car

	err = json.Unmarshal(bytes, &car)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal world state data to type Car")
	}

	return &car, nil
}

// DeleteCar removes the instance of Car from the world state
func (c *CarContract) DeleteCar(ctx contractapi.TransactionContextInterface, carID string) (string, error) {

	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("could not fetch client identity. %s", err)
	}
	if clientOrgID == "ManufacturerMSP" {
	// if clientOrgID == "Org1MSP" {
		//if clientOrgID == "manufacturer-auto-com" {

		exists, err := c.CarExists(ctx, carID)
		if err != nil {
			return "", fmt.Errorf("%s", err)
		} else if !exists {
			return "", fmt.Errorf("the car, %s does not exist", carID)
		}

		err = ctx.GetStub().DelState(carID)
		if err != nil {
			return "", err
		} else {
			return fmt.Sprintf("car with id %v is deleted from the world state.", carID), nil
		}

	} else {
		return "", fmt.Errorf("user under following MSPID: %v can't perform this action", clientOrgID)
	}
}

func (c *CarContract) GetCarsByRange(ctx contractapi.TransactionContextInterface, startKey, endKey string) ([]*Car, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, fmt.Errorf("could not fetch the  data by range. %s", err)
	}
	defer resultsIterator.Close()

	return carResultIteratorFunction(resultsIterator)
}

func (c *CarContract) GetAllCars(ctx contractapi.TransactionContextInterface) ([]*Car, error) {

	queryString := `{"selector":{"assetType":"car"}, "sort":[{ "color": "desc"}]}`

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("could not fetch the query result. %s", err)
	}
	defer resultsIterator.Close()
	return carResultIteratorFunction(resultsIterator)
}

func carResultIteratorFunction(resultsIterator shim.StateQueryIteratorInterface) ([]*Car, error) {
	var cars []*Car
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("could not fetch the details of the result iterator. %s", err)
		}
		var car Car
		err = json.Unmarshal(queryResult.Value, &car)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal the data. %s", err)
		}
		cars = append(cars, &car)
	}

	return cars, nil
}

func (c *CarContract) GetCarHistory(ctx contractapi.TransactionContextInterface, carID string) ([]*HistoryQueryResult, error) {

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(carID)
	if err != nil {
		return nil, fmt.Errorf("could not get the data. %s", err)
	}
	defer resultsIterator.Close()

	var records []*HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("could not get the value of resultsIterator. %s", err)
		}

		var car Car
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &car)
			if err != nil {
				return nil, err
			}
		} else {
			car = Car{
				CarId: carID,
			}
		}

		timestamp := response.Timestamp.AsTime()

		formattedTime := timestamp.Format(time.RFC1123)

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: formattedTime,
			Record:    &car,
			IsDelete:  response.IsDelete,
		}
		records = append(records, &record)
	}

	return records, nil
}

func (c *CarContract) GetCarsWithPagination(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string) (*PaginatedQueryResult, error) {

	queryString := `{"selector":{"assetType":"car"}}`

	resultsIterator, responseMetadata, err := ctx.GetStub().GetQueryResultWithPagination(queryString, pageSize, bookmark)
	if err != nil {
		return nil, fmt.Errorf("could not get the car records. %s", err)
	}
	defer resultsIterator.Close()

	cars, err := carResultIteratorFunction(resultsIterator)
	if err != nil {
		return nil, fmt.Errorf("could not return the car records %s", err)
	}

	return &PaginatedQueryResult{
		Records:             cars,
		FetchedRecordsCount: responseMetadata.FetchedRecordsCount,
		Bookmark:            responseMetadata.Bookmark,
	}, nil
}

// GetMatchingOrders get matching orders for car from the orders
func (c *CarContract) GetMatchingOrders(ctx contractapi.TransactionContextInterface, carID string) ([]*Order, error) {

	car, err := c.ReadCar(ctx, carID)
	if err != nil {
		return nil, fmt.Errorf("error reading car %v", err)
	}
	queryString := fmt.Sprintf(`{"selector":{"assetType":"Order","make":"%s", "model": "%s", "color":"%s"}}`, car.Make, car.Model, car.Color)
	resultsIterator, err := ctx.GetStub().GetPrivateDataQueryResult(collectionName, queryString)

	if err != nil {
		return nil, fmt.Errorf("could not get the data. %s", err)
	}
	defer resultsIterator.Close()

	return OrderResultIteratorFunction(resultsIterator)

}

// MatchOrder matches car with matching order
func (c *CarContract) MatchOrder(ctx contractapi.TransactionContextInterface, carID string, orderID string) (string, error) {

	bytes, err := ctx.GetStub().GetPrivateData(collectionName, orderID)
	if err != nil {
		return "", fmt.Errorf("could not get the private data: %s", err)
	}

	var order Order

	err = json.Unmarshal(bytes, &order)

	if err != nil {
		return "", fmt.Errorf("could not unmarshal the data. %s", err)
	}

	car, err := c.ReadCar(ctx, carID)
	if err != nil {
		return "", fmt.Errorf("could not read the data. %s", err)
	}

	if car.Make == order.Make && car.Color == order.Color && car.Model == order.Model {
		car.OwnedBy = order.DealerName
		car.Status = "assigned to a dealer"

		bytes, _ := json.Marshal(car)

		ctx.GetStub().DelPrivateData(collectionName, orderID)

		err = ctx.GetStub().PutState(carID, bytes)

		if err != nil {
			return "", fmt.Errorf("could not add the data %s", err)
		} else {
			return fmt.Sprintf("Deleted order %v and Assigned %v to %v", orderID, car.CarId, order.DealerName), nil
		}
	} else {
		return "", fmt.Errorf("order is not matching")
	}
}

// RegisterCar register car to the buyer
func (c *CarContract) RegisterCar(ctx contractapi.TransactionContextInterface, carID string, ownerName string, registrationNumber string) (string, error) {
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("could not get the MSPID %s", err)
	}

	if clientOrgID == "MvdMSP" {
	// if clientOrgID == "Org3MSP" {
		//if clientOrgID == "mvd-auto-com" {

		car, _ := c.ReadCar(ctx, carID)
		car.Status = fmt.Sprintf("Registered to  %v with plate number %v", ownerName, registrationNumber)
		car.OwnedBy = ownerName

		bytes, _ := json.Marshal(car)
		err = ctx.GetStub().PutState(carID, bytes)
		if err != nil {
			return "", fmt.Errorf("could not add the updated car details %s", err)
		} else {
			return fmt.Sprintf("Car %v successfully registered to %v", carID, ownerName), nil
		}

	} else {
		return "", fmt.Errorf("user under following MSPID: %v cannot able to perform this action", clientOrgID)
	}

}
