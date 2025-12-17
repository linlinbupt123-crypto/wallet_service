# Wallet Service

A simple HD wallet service supporting Ethereum (ETH) and Bitcoin (BTC) testnets.  
This service allows creating HD wallets, deriving addresses, sending transactions, and querying balances.
---
## Features

- Create an HD wallet for a user
- Derive main and additional addresses
- Send transactions (ETH/BTC)
- Query wallet balances
- Compatible with Ethereum Sepolia testnet


去这个创建 testNetwork 的 API
https://docs.metamask.io/developer-tools/faucet/

到Faucet网站领取测试币
https://cloud.google.com/application/web3/faucet/ethereum/sepolia

after send transaction you could search your exchange here by the tx hash
https://sepolia.etherscan.io

local env config
# zsh
export GVM_ROOT="$HOME/.gvm"
[[ -s "$GVM_ROOT/scripts/gvm" ]] && source "$GVM_ROOT/scripts/gvm"

# 默认使用 Go 1.24.9
gvm use go1.24.9 --default
vscode configure: 将vscode配置成golang的ide
1) 安装 VS Code 的 code 命令
打开 VS Code GUI。
按下 Cmd+Shift+P 打开命令面板。
输入并选择：
Shell Command: Install 'code' command in PATH

完成后，关闭终端，重新打开，然后测试：
code --version
如果能显示版本号，说明命令已安装成功。


2) 用终端启动 VS Code
这样 VS Code 会继承你的终端 PATH，包括 GVM 的 Go：
在终端输入这个命令 : code .
打开vscode ide以后，open project

然后在 VS Code 里：
打开命令面板 Cmd+Shift+P
运行 Go: Locate Configured Go Tools，应该就能找到 Go 和 gopls 了。

useful tools

MySQL UI
BeekeeperStudio

MongoDB
UI: https://www.mongodb.com/try/download/compass
