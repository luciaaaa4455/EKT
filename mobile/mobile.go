package mobile

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/EducationEKT/EKT/core/types"
	"github.com/EducationEKT/EKT/core/userevent"
	"github.com/EducationEKT/EKT/crypto"
	"github.com/EducationEKT/EKT/param"
	"github.com/EducationEKT/EKT/util"

	"github.com/EducationEKT/xserver/x_http/x_resp"
)

func init() {
	DelegateNode = param.MainNet
}

const (
	InvalidParam  = `{"status": -400, "resp": {}}`
	InternalError = `{"status": -500, "resp": {}}`
	NullResp      = `{"status": 0, "resp": {}}`
)

var DelegateNode []types.Peer

type GoMobileParam struct {
	Method string                 `json:"method"`
	Param  map[string]interface{} `json:"param"`
}

type GoMobileResp struct {
	Status int                    `json:"status"`
	Resp   map[string]interface{} `json:"resp"`
}

func Call(arg string) string {
	var param GoMobileParam
	err := json.Unmarshal([]byte(arg), &param)
	if err != nil {
		return InvalidParam
	}
	resp := call(param)
	if resp == "" {
		return InternalError
	}
	return resp
}

func call(param GoMobileParam) string {
	defer func() {
		recover()
	}()
	switch param.Method {
	case "CreateAccount":
		return createAccount()
	case "SignMsg":
		return signMsg(param)
	case "Sha3_256":
		return sha3_256(param)

	case "SendTransaction":
		return sendTransaction(param)
	}
	return NullResp
}

func sendTransaction(param GoMobileParam) string {
	tx, exist := param.Param["tx"]
	if !exist {
		return InvalidParam
	}
	transaction, ok := tx.(userevent.Transaction)
	if !ok {
		return InvalidParam
	}

	private, exist := param.Param["private"]
	if !exist {
		return InvalidParam
	}

	key, ok := private.(string)
	if !ok {
		return InvalidParam
	}

	privateKey, err := hex.DecodeString(key)
	if err != nil {
		return InvalidParam
	}

	pubKey, err := crypto.PubKey(privateKey)
	if err != nil {
		return InvalidParam
	}

	address := types.FromPubKeyToAddress(pubKey)

	transaction.From = address
	userevent.SignTransaction(&transaction, privateKey)
	success := _sendTx(transaction)
	return buildResp(0, map[string]interface{}{
		"success": success,
		"txId":    transaction.TransactionId(),
	})
}

func _sendTx(tx userevent.Transaction) bool {
	for _, node := range DelegateNode {
		url := fmt.Sprintf(`http://%s:%d/transaction/api/newTransaction`, node.Address, node.Port)
		resp, err := util.HttpPost(url, tx.Bytes())
		if err == nil {
			return true
		}
		var r x_resp.XRespBody
		err = json.Unmarshal(resp, &r)
		if err == nil && r.Status == 0 {
			return true
		}
	}
	return false
}

func sha3_256(param GoMobileParam) string {
	msg, exist := param.Param["msg"]
	if !exist {
		return InvalidParam
	}
	message, ok := msg.(string)
	if !ok {
		return InvalidParam
	}
	result := hex.EncodeToString(crypto.Sha3_256([]byte(message)))
	return buildResp(0, map[string]interface{}{
		"result": result,
	})
}

func signMsg(param GoMobileParam) string {
	privateKey, exist := param.Param["private"]
	if !exist {
		return InvalidParam
	}
	pk, ok := privateKey.(string)
	if !ok {
		return InvalidParam
	}
	msg, exist := param.Param["msg"]
	if !exist {
		return InvalidParam
	}
	message, ok := msg.(string)
	if !ok {
		return InvalidParam
	}
	result := SignMsg(message, pk)
	return buildResp(0, map[string]interface{}{
		"result": result,
	})
}

func createAccount() string {
	pub, private := crypto.GenerateKeyPair()
	address := PubKey2Address(hex.EncodeToString(pub))
	m := map[string]interface{}{
		"private": hex.EncodeToString(private),
		"address": address,
	}
	return buildResp(0, m)
}

func buildResp(status int, resp map[string]interface{}) string {
	response := GoMobileResp{
		Status: status,
		Resp:   resp,
	}
	data, _ := json.Marshal(response)
	return string(data)
}
