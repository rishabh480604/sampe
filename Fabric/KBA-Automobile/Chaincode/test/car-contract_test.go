package chaincodetest

import (
	"encoding/json"
	"fmt"
	"testing"

	"kbaauto/contracts"

	"kbaauto/test/mocks"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/stretchr/testify/require"
)

//go:generate counterfeiter -o mocks/transaction.go -fake-name TransactionContext ./car-contract_test.go transactionContext
type transactionContext interface {
	contractapi.TransactionContextInterface
}

//go:generate counterfeiter -o mocks/chaincodestub.go -fake-name ChaincodeStub ./car-contract_test.go chaincodeStub
type chaincodeStub interface {
	shim.ChaincodeStubInterface
}

//go:generate counterfeiter -o mocks/clientIdentity.go -fake-name ClientIdentity ./car-contract_test.go clientIdentity
type clientIdentity interface {
	cid.ClientIdentity
}

const orgMsp string = "manufacturer-auto-com"

func prepMocks(orgMSP string) (*mocks.TransactionContext, *mocks.ChaincodeStub) {
	// create an instance of chaincode stub
	chaincodeStub := &mocks.ChaincodeStub{}

	// create an instance of transaction context
	transactionContext := &mocks.TransactionContext{}

	// Associate chaincode stub with transaction context
	transactionContext.GetStubReturns(chaincodeStub)

	// create an instance of client identity
	clientIdentity := &mocks.ClientIdentity{}

	// Associate chaincode stub with transaction context
	transactionContext.GetClientIdentityReturns(clientIdentity)

	// Equip the client identity with corresponding msp id required
	clientIdentity.GetMSPIDReturns(orgMSP, nil)

	// Return transaction context and chaincode stub
	return transactionContext, chaincodeStub
}

func TestCreateCar(t *testing.T) {
	// Set up mocks
	transactionContext, chaincodeStub := prepMocks(orgMsp)

	// Create CarContract instance
	carAsset := contracts.CarContract{}

	// Configure mock behavior
	chaincodeStub.GetStateReturns(nil, nil)
	chaincodeStub.PutStateReturns(nil)

	// Call CreateCar with valid input
	result, err := carAsset.CreateCar(transactionContext, "car1", "Honda", "Civic", "Blue", "Factory-01", "2024-01-01")

	// Assert successful creation
	require.NoError(t, err)
	require.Equal(t, "successfully added car car1", result)

	// Assert car already exist
	chaincodeStub.GetStateReturns([]byte{}, nil)
	_, err = carAsset.CreateCar(transactionContext, "car1", "", "", "", "", "")
	require.EqualError(t, err, "the car, car1 already exists")

	// Assert reading error
	chaincodeStub.GetStateReturns(nil, fmt.Errorf("some error"))
	_, err = carAsset.CreateCar(transactionContext, "car1", "", "", "", "", "")
	require.EqualError(t, err, "failed to read from world state: some error")
}

func TestReadCar(t *testing.T) {
	// Set up mocks
	transactionContext, chaincodeStub := prepMocks(orgMsp)

	// Create CarContract instance
	carAsset := contracts.CarContract{}

	// Create pointer to Car struct
	expectedAsset := &contracts.Car{CarId: "car1"}
	bytes, err := json.Marshal(expectedAsset)
	require.NoError(t, err)


	// Configure mock behavior
	chaincodeStub.GetStateReturns(bytes, nil)

	// Assert successful reading
	car, err := carAsset.ReadCar(transactionContext, "car1")
	require.NoError(t, err)
	require.Equal(t, expectedAsset, car)

	// Assert reading error
	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve car"))
	_, err = carAsset.ReadCar(transactionContext, "")
	require.EqualError(t, err, "failed to read from world state: unable to retrieve car")

	// Assert car doesn't exist
	chaincodeStub.GetStateReturns(nil, nil)
	car, err = carAsset.ReadCar(transactionContext, "car1")
	require.EqualError(t, err, "the car car1 does not exist")
	require.Nil(t, car)
}

func TestDeleteCar(t *testing.T) {
	// Set up mocks
	transactionContext, chaincodeStub := prepMocks(orgMsp)

	// Create CarContract instance
	carAsset := contracts.CarContract{}

	// Create pointer to Car struct
	car := &contracts.Car{CarId: "car1"}
	bytes, err := json.Marshal(car)
	require.NoError(t, err)

	// Configure mock behavior
	chaincodeStub.GetStateReturns(bytes, nil)
	chaincodeStub.DelStateReturns(nil)

	// Assert successful removal of car
	result, err := carAsset.DeleteCar(transactionContext, "car1")
	require.NoError(t, err)
	require.Equal(t, "car with id car1 is deleted from the world state.", result)

	// Assert car doesn't exist
	chaincodeStub.GetStateReturns(nil, nil)
	_, err = carAsset.DeleteCar(transactionContext, "car1")
	require.EqualError(t, err, "the car, car1 does not exist")

	// Assert reading error
	chaincodeStub.GetStateReturns(nil, fmt.Errorf("unable to retrieve car"))
	_, err = carAsset.DeleteCar(transactionContext, "")
	require.EqualError(t, err, "failed to read from world state: unable to retrieve car")
}
