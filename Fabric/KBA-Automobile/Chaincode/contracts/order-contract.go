package contracts

// first write collection file

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// OrderContract contract for managing CRUD for Order
type OrderContract struct {
	contractapi.Contract
}

type Order struct {
	AssetType  string `json:"assetType"`
	Color      string `json:"color"`
	DealerName string `json:"dealerName"`
	Make       string `json:"make"`
	Model      string `json:"model"`
	OrderID    string `json:"orderID"`
}

const collectionName string = "OrderCollection"

// OrderExists returns true when asset with given ID exists in private data collection
func (o *OrderContract) OrderExists(ctx contractapi.TransactionContextInterface, orderID string) (bool, error) {

	data, err := ctx.GetStub().GetPrivateDataHash(collectionName, orderID)

	if err != nil {
		return false, fmt.Errorf("could not fetch the private data hash. %s", err)
	}

	return data != nil, nil
}

// CreateOrder creates a new instance of Order
func (o *OrderContract) CreateOrder(ctx contractapi.TransactionContextInterface, orderID string) (string, error) {

	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("could not fetch client identity. %s", err)
	}

	if clientOrgID == "DealerMSP" {
	// if clientOrgID == "Org2MSP" {
		//if clientOrgID == "dealer-auto-com" {
		exists, err := o.OrderExists(ctx, orderID)
		if err != nil {
			return "", fmt.Errorf("could not read from world state. %s", err)
		} else if exists {
			return "", fmt.Errorf("the asset %s already exists", orderID)
		}

		var order Order

		transientData, err := ctx.GetStub().GetTransient()
		if err != nil {
			return "", fmt.Errorf("could not fetch transient data. %s", err)
		}

		if len(transientData) == 0 {
			return "", fmt.Errorf("please provide the private data of make, model, color, dealerName")
		}

		make, exists := transientData["make"]
		if !exists {
			return "", fmt.Errorf("the make was not specified in transient data. Please try again")
		}
		order.Make = string(make)

		model, exists := transientData["model"]
		if !exists {
			return "", fmt.Errorf("the model was not specified in transient data. Please try again")
		}
		order.Model = string(model)

		color, exists := transientData["color"]
		if !exists {
			return "", fmt.Errorf("the color was not specified in transient data. Please try again")
		}
		order.Color = string(color)

		dealerName, exists := transientData["dealerName"]
		if !exists {
			return "", fmt.Errorf("the dealer was not specified in transient data. Please try again")
		}
		order.DealerName = string(dealerName)

		order.AssetType = "Order"
		order.OrderID = orderID

		bytes, _ := json.Marshal(order)
		err = ctx.GetStub().PutPrivateData(collectionName, orderID, bytes)
		if err != nil {
			return "", fmt.Errorf("could not able to write the data")
		}
		return fmt.Sprintf("order with id %v added successfully", orderID), nil
	} else {
		return fmt.Sprintf("order cannot be created by organisation with MSPID %v ", clientOrgID), nil
	}
}

// ReadOrder retrieves an instance of Order from the private data collection
func (o *OrderContract) ReadOrder(ctx contractapi.TransactionContextInterface, orderID string) (*Order, error) {
	exists, err := o.OrderExists(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("the asset %s does not exist", orderID)
	}

	bytes, err := ctx.GetStub().GetPrivateData(collectionName, orderID)
	if err != nil {
		return nil, fmt.Errorf("could not get the private data. %s", err)
	}
	var order Order

	err = json.Unmarshal(bytes, &order)

	if err != nil {
		return nil, fmt.Errorf("could not unmarshal private data collection data to type Order")
	}

	return &order, nil

}

// DeleteOrder deletes an instance of Order from the private data collection
func (o *OrderContract) DeleteOrder(ctx contractapi.TransactionContextInterface, orderID string) error {
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("could not read the client identity. %s", err)
	}

	if clientOrgID == "DealerMSP" {
	// if clientOrgID == "Org2MSP" {
		//if clientOrgID == "dealer-auto-com" {

		exists, err := o.OrderExists(ctx, orderID)

		if err != nil {
			return fmt.Errorf("could not read from world state. %s", err)
		} else if !exists {
			return fmt.Errorf("the asset %s does not exist", orderID)
		}

		return ctx.GetStub().DelPrivateData(collectionName, orderID)
	} else {
		return fmt.Errorf("organisation with %v cannot delete the order", clientOrgID)
	}
}

func (o *OrderContract) GetAllOrders(ctx contractapi.TransactionContextInterface) ([]*Order, error) {
	queryString := `{"selector":{"assetType":"Order"}}`
	resultsIterator, err := ctx.GetStub().GetPrivateDataQueryResult(collectionName, queryString)
	if err != nil {
		return nil, fmt.Errorf("could not fetch the query result. %s", err)
	}
	defer resultsIterator.Close()
	return OrderResultIteratorFunction(resultsIterator)
}

func (o *OrderContract) GetOrdersByRange(ctx contractapi.TransactionContextInterface, startKey string, endKey string) ([]*Order, error) {
	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(collectionName, startKey, endKey)

	if err != nil {
		return nil, fmt.Errorf("could not fetch the private data by range. %s", err)
	}
	defer resultsIterator.Close()

	return OrderResultIteratorFunction(resultsIterator)

}

// iterator function

func OrderResultIteratorFunction(resultsIterator shim.StateQueryIteratorInterface) ([]*Order, error) {
	var orders []*Order
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("could not fetch the details of result iterator. %s", err)
		}
		var order Order
		err = json.Unmarshal(queryResult.Value, &order)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal the data. %s", err)
		}
		orders = append(orders, &order)
	}

	return orders, nil
}
