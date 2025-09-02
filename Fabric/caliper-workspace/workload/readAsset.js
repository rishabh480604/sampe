'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class MyWorkload extends WorkloadModuleBase {
    constructor() {
        super();
    }

    /**
     * Initialize the workload module
     */
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);

        // Pre-create assets
        for (let i = 0; i < this.roundArguments.assets; i++) {
            const assetID = `${this.workerIndex}_${i}`;
            console.log(`Worker ${this.workerIndex}: Creating car ${assetID}`);

            const request = {
                contractId: this.roundArguments.contractId,  // should match your chaincode name
                contractFunction: 'CreateCar',             // matched Go chaincode function
                invokerIdentity: 'User1',
                contractArguments: [
                    assetID,           // carID
                    'Toyota',          // make
                    'Corolla',         // model
                    'Blue',            // color
                    'ManufacturerCo',  // manufacturerName
                    '2025-09-01'       // dateOfManufacture
                ],
                readOnly: false
            };

            await this.sutAdapter.sendRequests(request);
        }
    }

    /**
     * Submit transaction during workload
     */
    async submitTransaction() {
        const randomId = Math.floor(Math.random() * this.roundArguments.assets);
        const carID = `${this.workerIndex}_${randomId}`;

        const myArgs = {
            contractId: this.roundArguments.contractId,
            contractFunction: 'ReadCar', // matched Go chaincode function
            invokerIdentity: 'User1',
            contractArguments: [carID],
            readOnly: true
        };

        await this.sutAdapter.sendRequests(myArgs);
    }

    /**
     * Cleanup workload module (delete assets)
     */
    async cleanupWorkloadModule() {
        for (let i = 0; i < this.roundArguments.assets; i++) {
            const assetID = `${this.workerIndex}_${i}`;
            console.log(`Worker ${this.workerIndex}: Deleting car ${assetID}`);

            const request = {
                contractId: this.roundArguments.contractId,
                contractFunction: 'DeleteCar', // matched Go chaincode function
                invokerIdentity: 'User1',
                contractArguments: [assetID],
                readOnly: false
            };

            await this.sutAdapter.sendRequests(request);
        }
    }
}

/**
 * Factory function for Caliper
 */
function createWorkloadModule() {
    return new MyWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
