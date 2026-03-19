package mm

type walletType string

const (
	CUSTOMER_WALLET walletType = "CUSTOMER"
	MERCHANT_WALLET walletType = "MERCHANT"
)

type accountType string

const (
	ACC_DEFAULT  accountType = "DEFAULT"
	ACC_PAYMENT  accountType = "PAYMENT"
	ACC_INCOMING accountType = "INCOMING"
)

type wallet struct {
	id         int32
	userId     string
	walletType walletType
}

type account struct {
	id          int32
	paise       int64
	accountType accountType
	walletId    int32
}

type transaction struct {
	id                       int32
	pid                      string
	srcUserId                string
	dstUserId                string
	srcWalletId              int32
	dstWalletId              int32
	srcAccountId             int32
	dstAccountId             int32
	srcAccountType           accountType
	dstAccountType           accountType
	finalDstMerchantWalletId int32
	amount                   int64
}
