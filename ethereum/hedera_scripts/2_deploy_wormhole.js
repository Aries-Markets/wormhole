require("dotenv").config({ path: ".env" });
const {
  AccountId,
  PrivateKey,
  Client,
  ContractId,
  ContractCreateFlow,
  ContractFunctionParameters,
  ContractCallQuery,
} = require("@hashgraph/sdk");
const Web3 = require("web3");
const web3 = new Web3("ws://localhost:8545");

// CONFIG
const initialSigners = JSON.parse(process.env.INIT_SIGNERS);
const chainId = process.env.INIT_CHAIN_ID;
const governanceChainId = process.env.INIT_GOV_CHAIN_ID;
const governanceContract = process.env.INIT_GOV_CONTRACT; // bytes32

// Configure accounts and client
const operatorId = AccountId.fromString(process.env.OPERATOR_ID);
const operatorKey = PrivateKey.fromString(process.env.OPERATOR_PVKEY);

const client = Client.forTestnet().setOperator(operatorId, operatorKey);

// Based on example for ContractCreateFlow from https://docs.hedera.com/guides/docs/sdks/smart-contracts/create-a-smart-contract#methods
async function deploy(contractName, contractBytecode, gas, constructorFunctionParameters) {
  console.log("deploying " + contractName + ", gas: " + gas + ", byteCodeLen: " + contractBytecode.length)

  //Create the transaction
  const contractCreate = new ContractCreateFlow()
      .setGas(gas)
      .setBytecode(contractBytecode)
      .setConstructorParameters(constructorFunctionParameters)

  //Sign the transaction with the client operator key and submit to a Hedera network
  const txResponse = contractCreate.execute(client)

  //Get the receipt of the transaction
  const receipt = (await txResponse).getReceipt(client)

  //Get the new contract ID
  const contractId = (await receipt).contractId
  const contractAddress = "0x" + contractId.toSolidityAddress()

  console.log("deployed " + contractName + ", contractId: " + contractId + ", contractAddress: " + contractAddress)
  return contractAddress
}

async function main() {
  const Setup = require("../build/contracts/Setup.json");
  const SetupAddress = await deploy("Setup", Setup.bytecode, 100000, new ContractFunctionParameters());

  const Implementation = require("../build/contracts/Implementation.json");
  const ImplementationAddress = await deploy("Implementation", Implementation.bytecode, 100000, new ContractFunctionParameters());

  console.log("generating setup initialization data...");
  const setup = new web3.eth.Contract(Setup.abi, SetupAddress);
  const initData = setup.methods
    .setup(
      ImplementationAddress,
      initialSigners,
      chainId,
      governanceChainId,
      governanceContract
    )
    .encodeABI();

  const Wormhole = require("../build/contracts/Wormhole.json");
  const params = new ContractFunctionParameters()
    .addAddress(SetupAddress)
    .addBytes(new Uint8Array(Buffer.from(initData.substring(2), "hex")));

  await deploy("Wormhole", Wormhole.bytecode, 200000, params);
  console.log("Wormhole deploy complete")
}

main();
