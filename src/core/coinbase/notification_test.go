package coinbase_test

import (
	"fmt"
	"testing"

	. "github.com/Emyrk/LendingBot/src/core/coinbase"
)

var _ = fmt.Println

func TestCreatePayment(t *testing.T) {
	data, err := CreatePayment("hello@gmail.com")
	if err != nil {
		t.Error(err)
	}

	fmt.Println(string(data))
}

var exResp = `
{  
   "id":"bd43bcfa-84bd-5f67-a1a9-7994697dedfe",
   "type":"wallet:orders:paid",
   "data":{  
      "id":"5830f01f-61a2-5614-bee3-b651f1d83b83",
      "code":"8D3QK9SL",
      "type":"order",
      "name":"Test",
      "description":null,
      "amount":{  
         "amount":"0.00010000",
         "currency":"BTC"
      },
      "receipt_url":"https://www.coinbase.com/orders/8521e308d27eb340c0a93d4c765067ef/receipt",
      "resource":"order",
      "resource_path":"/v2/orders/5830f01f-61a2-5614-bee3-b651f1d83b83",
      "status":"paid",
      "overpaid":false,
      "bitcoin_amount":{  
         "amount":"0.00012000",
         "currency":"BTC"
      },
      "total_amount_received":{  
         "amount":"0.00012000",
         "currency":"BTC"
      },
      "payout_amount":null,
      "bitcoin_address":"1Pjt3SD3voBrUXK8N3jgn8hFduLYzTuDA6",
      "refund_address":"1AwdvNE7kSKt1vi1zAgcRTcVG5koLFA9vz",
      "bitcoin_uri":"bitcoin:1Pjt3SD3voBrUXK8N3jgn8hFduLYzTuDA6?amount=0.00012\u0026r=https://www.coinbase.com/r/5994669425bf3c01c108ed35",
      "notifications_url":null,
      "paid_at":"2017-08-16T15:37:21Z",
      "mispaid_at":null,
      "expires_at":"2017-08-16T15:51:52Z",
      "metadata":{  
         "custom":"TEST_ORDER1"
      },
      "created_at":"2017-08-16T15:36:52Z",
      "updated_at":"2017-08-16T15:37:21Z",
      "customer_info":{  
         "name":null,
         "email":"stevenmasley@gmail.com",
         "phone_number":null
      },
      "transaction":{  
         "id":"7e2c3fe9-e111-558c-a2cd-69885acb1d24",
         "resource":"transaction",
         "resource_path":"/v2/accounts/b4dabe96-6188-5640-81e5-5d5ba53b4154/transactions/7e2c3fe9-e111-558c-a2cd-69885acb1d24"
      },
      "mispayments":[  

      ],
      "refunds":[  

      ]
   },
   "user":{  
      "id":"79a7ca2f-25c9-5b86-8517-2e87fb14d10d",
      "resource":"user",
      "resource_path":"/v2/users/79a7ca2f-25c9-5b86-8517-2e87fb14d10d"
   },
   "account":{  
      "id":"b4dabe96-6188-5640-81e5-5d5ba53b4154",
      "resource":"account",
      "resource_path":"/v2/accounts/b4dabe96-6188-5640-81e5-5d5ba53b4154"
   },
   "delivery_attempts":0,
   "created_at":"2017-08-16T15:37:21Z",
   "resource":"notification",
   "resource_path":"/v2/notifications/bd43bcfa-84bd-5f67-a1a9-7994697dedfe",
   "additional_data":{  

   }
}`
