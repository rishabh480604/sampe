# sampe
test purpose

# Folder structure
```bash
KBA-Automobile/
├── Automobile-network/        # Hyperledger Fabric network files
├── Chaincode/                 # Your chaincode smart contracts
├── caliper-workplace/         # Caliper benchmark workspace
│   ├── networks/              # Network configuration files
│   └── benchmarks/            # Benchmark scenarios
```

# Deploy Chaincode


This repository contains the setup and benchmark scripts for the KBA-Automobile Hyperledger Fabric network using Hyperledger Caliper.

---

## Prerequisites

Ensure the following tools and dependencies are installed on your system:

- **cURL**
- **Node.js** (with npm)
- **Hyperledger Fabric binaries**
- **Hyperledger Caliper CLI**

---


## Setup Hyperledger Fabric

Download and setup Hyperledger Fabric binaries:

```bash
# Download Fabric samples, binaries, and Docker images
curl -sSL https://bit.ly/2ysbOFE | bash -s

# Add Fabric binaries to PATH
export PATH=${PWD}/fabric-samples/bin:$PATH

# Verify installation
peer version
configtxlator version

# (Optional) Download specific versions
curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.5.0 1.5.6
export PATH=${PWD}/bin:$PATH
```

## Start Network
- Navigate to KBA-Automobile/Automobile-workspace
- Place your chaincode in the Chaincode folder.

```bash
cd KBA-Automobile/Automobile-network

# Start the Fabric network
./startNetwork.sh
```

# Configure Caliper
- Navigate to KBA-Automobile/caliper-workspace
- Verify the network configuration file networkConfig.yaml inside the network folder and update paths if necessary.
- Modify the workload JS files in the workload folder according to your contract queries.

## Installing Dependencies

```bash

# Install Caliper CLI dependency
npm install --only=prod @hyperledger/caliper-cli@0.5.0

# Bind Caliper to Fabric
npx caliper bind --caliper-bind-sut fabric:2.4
```

## npx caliper launch manager \
  --caliper-workspace ./ \
  --caliper-networkconfig networks/networkConfig.yaml \
  --caliper-benchconfig benchmarks/myAssetBenchmark.yaml \
  --caliper-flow-only-test \
  --caliper-fabric-gateway-enabled




