package utils

/*
BIP-44 路径解析这个路径中的每一个斜杠 / 分隔的部分都代表了密钥树结构中的一个层级（Level）。
路径的结构遵循以下五个层级：m / purpose' / coin_type' / account' / change / address_index
层级	路径中的值	名称(BIP-44)		  描述
1		m         Master Key     	   代表主私钥（从助记词派生），是所有后续密钥的起点。
2       44'		  Purpose        	   目的。44' 是 BIP-44 标准的编号。它告诉软件这个钱包路径是按照 BIP-44 规范生成的。撇号 (' ) 表示这个层级的密钥是硬化派生 (Hardened Derivation) 的。
3       60'       Coin           	   Type币种类型。这个数字代表特定的加密货币。

	例如：
		• $0' = Bitcoin (BTC)
		• $1' = Testnet (测试网络)
		• $60' = Ethereum (ETH)

4 		0'		  Account账户 			允许用户将他们的资金分到不同的“账户”中（类似于银行账户）。通常从 $0'$ 开始编号。
5		0		  Change找零/内部。      用 0 表示外部链（External Chain），用于接收地址。用1表示内部链（Internal Chain/Change Chain），用于找零地址（主要用于 UTXO 模型，如比特币）。没有撇号表示常规派生 (Normal Derivation)。
6.      0.        Address Index地址索引。这是特定账户中派生出的第 N 个地址。从 0 开始计数 ($0, 1, 2, ...$)。没有撇号表示常规派生。
*/
const (
	ETH_DERIVATION_PATH_PREFIX = "m/44'/60'/0'/0/"
)
