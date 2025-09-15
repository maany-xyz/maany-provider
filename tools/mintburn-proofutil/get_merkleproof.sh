NODE=http://localhost:26657
curl "$NODE/abci_query?path=\"/store/x-mintburn/key\"&data=0x0131&prove=true&height=102" > proof.json
go run ./tools/mintburn-proofutil < proof.json > merkle_proof.json
cat merkle_proof.json 