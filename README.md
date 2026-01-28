# Wallet Service
A simple wallet service supporting Ethereum (ETH) testnets.  
This service allows creating HD wallets, deriving addresses, sending transactions, and querying balances.
---
## Features

- Create an HD wallet for a user
- Derive main and additional addresses
- Send transactions (ETH/BTC)
- Query wallet balances
- Compatible with Ethereum Sepolia testnet
- Import local hardhat private to the wallet

This system adopts a three-level model: 
- User → Wallet → Address.
- The Wallet layer is responsible for key management, while the Address layer represents on-chain identities and handles blockchain interactions.
- This design provides a unified abstraction for both HD wallets and imported private-key wallets.

apply test ETH from Faucet
- https://cloud.google.com/application/web3/faucet/ethereum/sepolia

after send transaction you could search your exchange here by the tx hash
- https://sepolia.etherscan.io

所有 ETH 地址：
- 输入：接受任意合法 hex
- 内部：统一 EIP-55 checksum
- 存储：checksum
- 输出：checksum