'use strict';

const { WorkloadModuleBase } = require('@hyperledger/caliper-core');

class MyWorkload extends WorkloadModuleBase {
    constructor() {
        super();
    }

    /**
     * Initialize workload module
     */
    async initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext) {
        await super.initializeWorkloadModule(workerIndex, totalWorkers, roundIndex, roundArguments, sutAdapter, sutContext);
        console.log(`Worker ${this.workerIndex}: Initialization complete`);
    }

    /**
     * Submit auction transaction
     */
    async submitTransaction() {
        const request = {
            contractId: this.roundArguments.contractId,  // should match your deployed chaincode name
            contractFunction: 'MapBuyersToSellers',
            invokerIdentity: 'User1',
            contractArguments: [
                'https://raw.githubusercontent.com/rishabh480604/sampe/refs/heads/main/50_seller_data_tsv.txt', // seller_url
                'https://raw.githubusercontent.com/rishabh480604/sampe/refs/heads/main/190_buyer_data_tsv.txt'  // buyer_url
            ],
            readOnly: false
        };

        console.log(`Worker ${this.workerIndex}: Submitting MapBuyersToSellers transaction`);
        await this.sutAdapter.sendRequests(request);
    }

    /**
     * Cleanup workload module
     */
    async cleanupWorkloadModule() {
        console.log(`Worker ${this.workerIndex}: Cleanup complete`);
    }
}

/**
 * Factory function for Caliper
 */
function createWorkloadModule() {
    return new MyWorkload();
}

module.exports.createWorkloadModule = createWorkloadModule;
