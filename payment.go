// Package payment 实现了第三方支付协议， 比如 微信， 支付宝等
package payment

import (
	"errors"
	"fmt"
	"io"
	"time"
)

var fnNoProviderErr = func(plat PayPlat) error { return fmt.Errorf("Not found provider %s", plat) }

var providerMap = make(map[PayPlat]Provider, 0)

var DoubleSubmitError = errors.New("重复提交")

// Register 注册支付提供程序
func Register(plat PayPlat, provider Provider) {
	providerMap[plat] = provider
}

// Order 提交支付请求
func Order(plat PayPlat, r *OrderRequest) (*OrderResponse, error) {
	if v, ok := providerMap[plat]; ok {
		return v.Order(r)
	}
	return nil, fnNoProviderErr(plat)
}

// HandleNotify 异步通知处理
func HandleNotify(plat PayPlat, r io.Reader, bizFunc NotifyHandleFunc) (interface{}, error) {
	if v, ok := providerMap[plat]; ok {
		return v.NotifyCallback(r, bizFunc), nil
	}
	return nil, fnNoProviderErr(plat)
}

// Provider 支付提供实现
type Provider interface {
	// Order 下单提交支付请求
	Order(order *OrderRequest) (*OrderResponse, error)
	// NotifyCallback 后台异步支付通知处理
	NotifyCallback(r io.Reader, f NotifyHandleFunc) interface{}
}

// PayPlat 第三方支付平台
type PayPlat string

const (
	// PayPlatWechat 微信支付
	PayPlatWechat PayPlat = "wechat"
	// PayPlatAlipay 支付宝支付
	PayPlatAlipay PayPlat = "alipay"
)

// NotifyResult 异步通知结果
type NotifyResult struct {
	Plat            PayPlat   //支付平台
	MerchantOrderNo string    //商户订单号
	TransactionID   string    // 交易ID
	CompletedTime   time.Time //完成时间
	TotalAmount     int64     //支付金额，单位：分
	Currency        string    //币种
	Attach          string    //附加数据
	Wechat          struct {
		OpenID string
	}
	Alipay struct {
		BuyerID, BuyerLoginID string
		NotifyID              string
	}
}

// NotifyHandleFunc 业务回调处理函数
type NotifyHandleFunc func(result *NotifyResult) error

// OrderResponse 下单返回结果
type OrderResponse struct {
	Wechat struct {
		PrepayID string      `json:",omitempty" xml:",omitempty"`
		PayForm  interface{} `json:",omitempty" xml:",omitempty"`
	} `json:",omitempty" xml:",omitempty"`
	Alipay struct {
		PayForm string `json:",omitempty" xml:",omitempty"`
	}
}

// OrderRequest 创单请求
type OrderRequest struct {
	Details        ProductDetails
	Subject        string
	Desc           string
	Attach         string
	MerchanOrderNo string
	Amount         int64
	ClientIP       string
	Tag            string
	TradeType      string
	ProduceID      string
	OpenID         string
	FromWeb        bool
}

// ProductDetails 商品详情集合
type ProductDetails []ProductDetail

// ProductDetail 商品详情
type ProductDetail struct {
	GoodsID   string `json:"goods_id"`                 //必填 32 商品编号
	WXGoodsID string `json:"wxpay_goods_id,omitempty"` //可选 32 微信支付定义的统一商品编号
	GoodsName string `json:"goods_name"`               //必填 256 商品名称
	Category  string `json:"goods_category"`           // 可选 32 商品类目ID
	Body      string `json:"body,omitempty"`           // 可选 1000 商品描述信息
	Num       int    `json:"goods_num"`                //必填 商品数量
	Price     int64  `json:"price"`                    //必填 商品单价，单位为分
}
