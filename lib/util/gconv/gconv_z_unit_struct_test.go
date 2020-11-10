package gconv_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sjatsh/beanstalk-go/lib/os/gtime"
	"github.com/sjatsh/beanstalk-go/lib/test/gtest"
	"github.com/sjatsh/beanstalk-go/lib/util/gconv"
)

func Test_Struct_Basic1(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type User struct {
			Uid      int
			Name     string
			Site_Url string
			NickName string
			Pass1    string `gconv:"password1"`
			Pass2    string `gconv:"password2"`
		}
		user := new(User)
		params1 := map[string]interface{}{
			"uid":       1,
			"Name":      "john",
			"siteurl":   "https://goframe.org",
			"nick_name": "johng",
			"PASS1":     "123",
			"PASS2":     "456",
		}
		if err := gconv.Struct(params1, user); err != nil {
			t.Error(err)
		}
		t.Assert(user, &User{
			Uid:      1,
			Name:     "john",
			Site_Url: "https://goframe.org",
			NickName: "johng",
			Pass1:    "123",
			Pass2:    "456",
		})

		user = new(User)
		params2 := map[string]interface{}{
			"uid":       2,
			"name":      "smith",
			"site-url":  "https://goframe.org",
			"nick name": "johng",
			"password1": "111",
			"password2": "222",
		}
		if err := gconv.Struct(params2, user); err != nil {
			t.Error(err)
		}
		t.Assert(user, &User{
			Uid:      2,
			Name:     "smith",
			Site_Url: "https://goframe.org",
			NickName: "johng",
			Pass1:    "111",
			Pass2:    "222",
		})
	})
}

func Test_Struct_Basic2(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type User struct {
			Uid     int
			Name    string
			SiteUrl string
			Pass1   string
			Pass2   string
		}
		user := new(User)
		params := map[string]interface{}{
			"uid":      1,
			"Name":     "john",
			"site_url": "https://goframe.org",
			"PASS1":    "123",
			"PASS2":    "456",
		}
		if err := gconv.Struct(params, user); err != nil {
			t.Error(err)
		}
		t.Assert(user, &User{
			Uid:     1,
			Name:    "john",
			SiteUrl: "https://goframe.org",
			Pass1:   "123",
			Pass2:   "456",
		})
	})
}

func Test_Struct_Attr_Pointer(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type User struct {
			Uid  *int
			Name *string
		}
		user := new(User)
		params := map[string]interface{}{
			"uid":  "1",
			"Name": "john",
		}
		if err := gconv.Struct(params, user); err != nil {
			t.Error(err)
		}
		t.Assert(user.Uid, 1)
		t.Assert(*user.Name, "john")
	})
}

func Test_Struct_Attr_Slice1(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type User struct {
			Scores []int
		}
		scores := []interface{}{99, 100, 60, 140}
		user := new(User)
		if err := gconv.Struct(map[string]interface{}{"Scores": scores}, user); err != nil {
			t.Error(err)
		} else {
			t.Assert(user, &User{
				Scores: []int{99, 100, 60, 140},
			})
		}
	})
}

// It does not support this kind of converting yet.
// func Test_Struct_Attr_Slice2(t *testing.T) {
//	gtest.C(t, func(t *gtest.T) {
//		type User struct {
//			Scores [][]int
//		}
//		scores := []interface{}{[]interface{}{99, 100, 60, 140}}
//		user := new(User)
//		if err := gconv.Struct(g.Map{"Scores": scores}, user); err != nil {
//			t.Error(err)
//		} else {
//			t.Assert(user, &User{
//				Scores: [][]int{{99, 100, 60, 140}},
//			})
//		}
//	})
// }

func Test_Struct_Attr_Struct(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type Score struct {
			Name   string
			Result int
		}
		type User struct {
			Scores Score
		}

		user := new(User)
		scores := map[string]interface{}{
			"Scores": map[string]interface{}{
				"Name":   "john",
				"Result": 100,
			},
		}

		// 嵌套struct转换
		if err := gconv.Struct(scores, user); err != nil {
			t.Error(err)
		} else {
			t.Assert(user, &User{
				Scores: Score{
					Name:   "john",
					Result: 100,
				},
			})
		}
	})
}

func Test_Struct_Attr_Struct_Ptr(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type Score struct {
			Name   string
			Result int
		}
		type User struct {
			Scores *Score
		}

		user := new(User)
		scores := map[string]interface{}{
			"Scores": map[string]interface{}{
				"Name":   "john",
				"Result": 100,
			},
		}

		// 嵌套struct转换
		if err := gconv.Struct(scores, user); err != nil {
			t.Error(err)
		} else {
			t.Assert(user.Scores, &Score{
				Name:   "john",
				Result: 100,
			})
		}
	})
}

func Test_Struct_Attr_Struct_Slice1(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type Score struct {
			Name   string
			Result int
		}
		type User struct {
			Scores []Score
		}

		user := new(User)
		scores := map[string]interface{}{
			"Scores": map[string]interface{}{
				"Name":   "john",
				"Result": 100,
			},
		}

		if err := gconv.Struct(scores, user); err != nil {
			t.Error(err)
		} else {
			t.Assert(user.Scores, []Score{
				{
					Name:   "john",
					Result: 100,
				},
			})
		}
	})
}

func Test_Struct_Attr_Struct_Slice2(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type Score struct {
			Name   string
			Result int
		}
		type User struct {
			Scores []Score
		}

		user := new(User)
		scores := map[string]interface{}{
			"Scores": []interface{}{
				map[string]interface{}{
					"Name":   "john",
					"Result": 100,
				},
				map[string]interface{}{
					"Name":   "smith",
					"Result": 60,
				},
			},
		}

		if err := gconv.Struct(scores, user); err != nil {
			t.Error(err)
		} else {
			t.Assert(user.Scores, []Score{
				{
					Name:   "john",
					Result: 100,
				},
				{
					Name:   "smith",
					Result: 60,
				},
			})
		}
	})
}

func Test_Struct_Attr_Struct_Slice_Ptr(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type Score struct {
			Name   string
			Result int
		}
		type User struct {
			Scores []*Score
		}

		user := new(User)
		scores := map[string]interface{}{
			"Scores": []interface{}{
				map[string]interface{}{
					"Name":   "john",
					"Result": 100,
				},
				map[string]interface{}{
					"Name":   "smith",
					"Result": 60,
				},
			},
		}

		// 嵌套struct转换，属性为slice类型，数值为slice map类型
		if err := gconv.Struct(scores, user); err != nil {
			t.Error(err)
		} else {
			t.Assert(len(user.Scores), 2)
			t.Assert(user.Scores[0], &Score{
				Name:   "john",
				Result: 100,
			})
			t.Assert(user.Scores[1], &Score{
				Name:   "smith",
				Result: 60,
			})
		}
	})
}

func Test_Struct_Attr_CustomType1(t *testing.T) {
	type MyInt int
	type User struct {
		Id   MyInt
		Name string
	}
	gtest.C(t, func(t *gtest.T) {
		user := new(User)
		err := gconv.Struct(map[string]interface{}{"id": 1, "name": "john"}, user)
		t.Assert(err, nil)
		t.Assert(user.Id, 1)
		t.Assert(user.Name, "john")
	})
}

func Test_Struct_Attr_CustomType2(t *testing.T) {
	type MyInt int
	type User struct {
		Id   []MyInt
		Name string
	}
	gtest.C(t, func(t *gtest.T) {
		user := new(User)
		err := gconv.Struct(map[string]interface{}{"id": []int{1, 2}, "name": "john"}, user)
		t.Assert(err, nil)
		t.Assert(user.Id, []int{1, 2})
		t.Assert(user.Name, "john")
	})
}

func Test_Struct_PrivateAttribute(t *testing.T) {
	type User struct {
		Id   int
		name string
	}
	gtest.C(t, func(t *gtest.T) {
		user := new(User)
		err := gconv.Struct(map[string]interface{}{"id": 1, "name": "john"}, user)
		t.Assert(err, nil)
		t.Assert(user.Id, 1)
		t.Assert(user.name, "")
	})
}

func Test_StructDeep1(t *testing.T) {
	type Base struct {
		Age int
	}
	type User struct {
		Id   int
		Name string
		Base
	}
	gtest.C(t, func(t *gtest.T) {
		user := new(User)
		params := map[string]interface{}{
			"id":   1,
			"name": "john",
			"age":  18,
		}
		err := gconv.Struct(params, user)
		t.Assert(err, nil)
		t.Assert(user.Id, params["id"])
		t.Assert(user.Name, params["name"])
		t.Assert(user.Age, 0)
	})

	gtest.C(t, func(t *gtest.T) {
		user := new(User)
		params := map[string]interface{}{
			"id":   1,
			"name": "john",
			"age":  18,
		}
		err := gconv.StructDeep(params, user)
		t.Assert(err, nil)
		t.Assert(user.Id, params["id"])
		t.Assert(user.Name, params["name"])
		t.Assert(user.Age, params["age"])
	})
}

func Test_StructDeep2(t *testing.T) {
	type Ids struct {
		Id  int
		Uid int
	}
	type Base struct {
		Ids
		Time string
	}
	type User struct {
		Base
		Name string
	}
	params := map[string]interface{}{
		"id":   1,
		"uid":  10,
		"name": "john",
	}
	gtest.C(t, func(t *gtest.T) {
		user := new(User)
		err := gconv.Struct(params, user)
		t.Assert(err, nil)
		t.Assert(user.Id, 0)
		t.Assert(user.Uid, 0)
		t.Assert(user.Name, "john")
	})

	gtest.C(t, func(t *gtest.T) {
		user := new(User)
		err := gconv.StructDeep(params, user)
		t.Assert(err, nil)
		t.Assert(user.Id, 1)
		t.Assert(user.Uid, 10)
		t.Assert(user.Name, "john")
	})
	gtest.C(t, func(t *gtest.T) {
		user := (*User)(nil)
		err := gconv.StructDeep(params, &user)
		t.Assert(err, nil)
		t.Assert(user.Id, 1)
		t.Assert(user.Uid, 10)
		t.Assert(user.Name, "john")
	})
}

func Test_StructDeep3(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type Ids struct {
			Id  int `json:"id"`
			Uid int `json:"uid"`
		}
		type Base struct {
			Ids
			CreateTime string `json:"create_time"`
		}
		type User struct {
			Base
			Passport string `json:"passport"`
			Password string `json:"password"`
			Nickname string `json:"nickname"`
		}
		data := map[string]interface{}{
			"id":          100,
			"uid":         101,
			"passport":    "t1",
			"password":    "123456",
			"nickname":    "T1",
			"create_time": "2019",
		}
		user := new(User)
		err := gconv.StructDeep(data, user)
		t.Assert(err, nil)
		t.Assert(user.Id, 100)
		t.Assert(user.Uid, 101)
		t.Assert(user.Nickname, "T1")
		t.Assert(user.CreateTime, "2019")
	})
}

// https://github.com/gogf/gf/issues/775
func Test_StructDeep4(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type Sub2 struct {
			SubName string
		}
		type sub1 struct {
			Sub2
			Name string
		}
		type Test struct {
			Sub sub1 `json:"sub"`
		}

		data := `{
    "sub": {
		"map":{"k":"v"},
        "Name": "name",
        "SubName": "subname"
    }}`

		expect := Test{
			Sub: sub1{
				Name: "name",
				Sub2: Sub2{
					SubName: "subname",
				},
			},
		}
		tx := new(Test)
		if err := gconv.StructDeep(data, &tx); err != nil {
			panic(err)
		}
		t.Assert(tx, expect)
	})
}

func Test_StructDeep5(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type Base struct {
			Pass1 string `params:"password1"`
			Pass2 string `params:"password2"`
		}
		type UserWithBase1 struct {
			Id   int
			Name string
			Base
		}
		type UserWithBase2 struct {
			Id   int
			Name string
			Pass Base
		}

		data := map[string]interface{}{
			"id":        1,
			"name":      "john",
			"password1": "123",
			"password2": "456",
		}
		var err error
		user1 := new(UserWithBase1)
		user2 := new(UserWithBase2)
		err = gconv.StructDeep(data, user1)
		t.Assert(err, nil)
		t.Assert(user1, &UserWithBase1{1, "john", Base{"123", "456"}})

		err = gconv.StructDeep(data, user2)
		t.Assert(err, nil)
		t.Assert(user2, &UserWithBase2{1, "john", Base{"123", "456"}})
	})
}

func Test_Struct_Time(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type User struct {
			CreateTime time.Time
		}
		now := time.Now()
		user := new(User)
		gconv.Struct(map[string]interface{}{
			"create_time": now,
		}, user)
		t.Assert(user.CreateTime.UTC().String(), now.UTC().String())
	})

	gtest.C(t, func(t *gtest.T) {
		type User struct {
			CreateTime *time.Time
		}
		now := time.Now()
		user := new(User)
		gconv.Struct(map[string]interface{}{
			"create_time": &now,
		}, user)
		t.Assert(user.CreateTime.UTC().String(), now.UTC().String())
	})

	gtest.C(t, func(t *gtest.T) {
		type User struct {
			CreateTime *gtime.Time
		}
		now := time.Now()
		user := new(User)
		gconv.Struct(map[string]interface{}{
			"create_time": &now,
		}, user)
		t.Assert(user.CreateTime.Time.UTC().String(), now.UTC().String())
	})

	gtest.C(t, func(t *gtest.T) {
		type User struct {
			CreateTime gtime.Time
		}
		now := time.Now()
		user := new(User)
		gconv.Struct(map[string]interface{}{
			"create_time": &now,
		}, user)
		t.Assert(user.CreateTime.Time.UTC().String(), now.UTC().String())
	})

	gtest.C(t, func(t *gtest.T) {
		type User struct {
			CreateTime gtime.Time
		}
		now := time.Now()
		user := new(User)
		gconv.Struct(map[string]interface{}{
			"create_time": now,
		}, user)
		t.Assert(user.CreateTime.Time.UTC().String(), now.UTC().String())
	})
}

// Auto create struct when given pointer.
func Test_Struct_Create(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type User struct {
			Uid  int
			Name string
		}
		user := (*User)(nil)
		params := map[string]interface{}{
			"uid":  1,
			"Name": "john",
		}
		err := gconv.Struct(params, &user)
		t.Assert(err, nil)
		t.Assert(user.Uid, 1)
		t.Assert(user.Name, "john")
	})

	gtest.C(t, func(t *gtest.T) {
		type User struct {
			Uid  int
			Name string
		}
		user := (*User)(nil)
		params := map[string]interface{}{
			"uid":  1,
			"Name": "john",
		}
		err := gconv.Struct(params, user)
		t.AssertNE(err, nil)
		t.Assert(user, nil)
	})
}

func Test_Struct_Interface(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type User struct {
			Uid  interface{}
			Name interface{}
		}
		user := (*User)(nil)
		params := map[string]interface{}{
			"uid":  1,
			"Name": nil,
		}
		err := gconv.Struct(params, &user)
		t.Assert(err, nil)
		t.Assert(user.Uid, 1)
		t.Assert(user.Name, nil)
	})
}

func Test_Struct_NilAttribute(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type Item struct {
			Title string `json:"title"`
			Key   string `json:"key"`
		}

		type M struct {
			Id    string                 `json:"id"`
			Me    map[string]interface{} `json:"me"`
			Txt   string                 `json:"txt"`
			Items []*Item                `json:"items"`
		}
		m := new(M)
		err := gconv.Struct(map[string]interface{}{
			"id": "88888",
			"me": map[string]interface{}{
				"name": "mikey",
				"day":  "20009",
			},
			"txt":   "hello",
			"items": nil,
		}, m)
		t.Assert(err, nil)
		t.AssertNE(m.Me, nil)
		t.Assert(m.Me["day"], "20009")
		t.Assert(m.Items, nil)
	})
}

func Test_Struct_Complex(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type ApplyReportDetail struct {
			ApplyScore        string `json:"apply_score"`
			ApplyCredibility  string `json:"apply_credibility"`
			QueryOrgCount     string `json:"apply_query_org_count"`
			QueryFinanceCount string `json:"apply_query_finance_count"`
			QueryCashCount    string `json:"apply_query_cash_count"`
			QuerySumCount     string `json:"apply_query_sum_count"`
			LatestQueryTime   string `json:"apply_latest_query_time"`
			LatestOneMonth    string `json:"apply_latest_one_month"`
			LatestThreeMonth  string `json:"apply_latest_three_month"`
			LatestSixMonth    string `json:"apply_latest_six_month"`
		}
		type BehaviorReportDetail struct {
			LoansScore         string `json:"behavior_report_detailloans_score"`
			LoansCredibility   string `json:"behavior_report_detailloans_credibility"`
			LoansCount         string `json:"behavior_report_detailloans_count"`
			LoansSettleCount   string `json:"behavior_report_detailloans_settle_count"`
			LoansOverdueCount  string `json:"behavior_report_detailloans_overdue_count"`
			LoansOrgCount      string `json:"behavior_report_detailloans_org_count"`
			ConsfinOrgCount    string `json:"behavior_report_detailconsfin_org_count"`
			LoansCashCount     string `json:"behavior_report_detailloans_cash_count"`
			LatestOneMonth     string `json:"behavior_report_detaillatest_one_month"`
			LatestThreeMonth   string `json:"behavior_report_detaillatest_three_month"`
			LatestSixMonth     string `json:"behavior_report_detaillatest_six_month"`
			HistorySucFee      string `json:"behavior_report_detailhistory_suc_fee"`
			HistoryFailFee     string `json:"behavior_report_detailhistory_fail_fee"`
			LatestOneMonthSuc  string `json:"behavior_report_detaillatest_one_month_suc"`
			LatestOneMonthFail string `json:"behavior_report_detaillatest_one_month_fail"`
			LoansLongTime      string `json:"behavior_report_detailloans_long_time"`
			LoansLatestTime    string `json:"behavior_report_detailloans_latest_time"`
		}
		type CurrentReportDetail struct {
			LoansCreditLimit    string `json:"current_report_detailloans_credit_limit"`
			LoansCredibility    string `json:"current_report_detailloans_credibility"`
			LoansOrgCount       string `json:"current_report_detailloans_org_count"`
			LoansProductCount   string `json:"current_report_detailloans_product_count"`
			LoansMaxLimit       string `json:"current_report_detailloans_max_limit"`
			LoansAvgLimit       string `json:"current_report_detailloans_avg_limit"`
			ConsfinCreditLimit  string `json:"current_report_detailconsfin_credit_limit"`
			ConsfinCredibility  string `json:"current_report_detailconsfin_credibility"`
			ConsfinOrgCount     string `json:"current_report_detailconsfin_org_count"`
			ConsfinProductCount string `json:"current_report_detailconsfin_product_count"`
			ConsfinMaxLimit     string `json:"current_report_detailconsfin_max_limit"`
			ConsfinAvgLimit     string `json:"current_report_detailconsfin_avg_limit"`
		}
		type ResultDetail struct {
			ApplyReportDetail    ApplyReportDetail    `json:"apply_report_detail"`
			BehaviorReportDetail BehaviorReportDetail `json:"behavior_report_detail"`
			CurrentReportDetail  CurrentReportDetail  `json:"current_report_detail"`
		}

		type Data struct {
			Code         string       `json:"code"`
			Desc         string       `json:"desc"`
			TransID      string       `json:"trans_id"`
			TradeNo      string       `json:"trade_no"`
			Fee          string       `json:"fee"`
			IDNo         string       `json:"id_no"`
			IDName       string       `json:"id_name"`
			Versions     string       `json:"versions"`
			ResultDetail ResultDetail `json:"result_detail"`
		}

		type XinYanModel struct {
			Success   bool        `json:"success"`
			Data      Data        `json:"data"`
			ErrorCode interface{} `json:"errorCode"`
			ErrorMsg  interface{} `json:"errorMsg"`
		}

		var data = `{
    "success": true,
    "data": {
        "code": "0",
        "desc": "查询成功",
        "trans_id": "14910304379231213",
        "trade_no": "201704011507240100057329",
        "fee": "Y",
        "id_no": "0783231bcc39f4957e99907e02ae401c",
        "id_name": "dd67a5943781369ddd7c594e231e9e70",
        "versions": "1.0.0",
        "result_detail":{
            "apply_report_detail": {
                "apply_score": "189",
                "apply_credibility": "84",
                "query_org_count": "7",
                "query_finance_count": "2",
                "query_cash_count": "2",
                "query_sum_count": "13",
                "latest_query_time": "2017-09-03",
                "latest_one_month": "1",
                "latest_three_month": "5",
                "latest_six_month": "12"
            },
            "behavior_report_detail": {
                "loans_score": "199",
                "loans_credibility": "90",
                "loans_count": "300",
                "loans_settle_count": "280",
                "loans_overdue_count": "20",
                "loans_org_count": "5",
                "consfin_org_count": "3",
                "loans_cash_count": "2",
                "latest_one_month": "3",
                "latest_three_month": "20",
                "latest_six_month": "23",
                "history_suc_fee": "30",
                "history_fail_fee": "25",
                "latest_one_month_suc": "5",
                "latest_one_month_fail": "20",
                "loans_long_time": "130",
                "loans_latest_time": "2017-09-16"
            },
            "current_report_detail": {
                "loans_credit_limit": "1400",
                "loans_credibility": "80",
                "loans_org_count": "7",
                "loans_product_count": "8",
                "loans_max_limit": "2000",
                "loans_avg_limit": "1000",
                "consfin_credit_limit": "1500",
                "consfin_credibility": "90",
                "consfin_org_count": "8",
                "consfin_product_count": "5",
                "consfin_max_limit": "5000",
                "consfin_avg_limit": "3000"
            }
        }
    },
    "errorCode": null,
    "errorMsg": null
}`
		m := make(map[string]interface{})
		err := json.Unmarshal([]byte(data), &m)
		t.Assert(err, nil)

		model := new(XinYanModel)
		err = gconv.Struct(m, model)
		t.Assert(err, nil)
		t.Assert(model.ErrorCode, nil)
		t.Assert(model.ErrorMsg, nil)
		t.Assert(model.Success, true)
		t.Assert(model.Data.IDName, "dd67a5943781369ddd7c594e231e9e70")
		t.Assert(model.Data.TradeNo, "201704011507240100057329")
		t.Assert(model.Data.ResultDetail.ApplyReportDetail.ApplyScore, "189")
		t.Assert(model.Data.ResultDetail.BehaviorReportDetail.LoansSettleCount, "280")
		t.Assert(model.Data.ResultDetail.CurrentReportDetail.LoansProductCount, "8")
	})
}

func Test_Struct_CatchPanic(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		type Score struct {
			Name   string
			Result int
		}
		type User struct {
			Score
		}

		user := new(User)
		scores := map[string]interface{}{
			"Score": 1,
		}
		err := gconv.Struct(scores, user)
		t.AssertNE(err, nil)
	})
}

type MyTime struct {
	time.Time
}

type MyTimeSt struct {
	ServiceDate MyTime
}

func (st *MyTimeSt) UnmarshalValue(v interface{}) error {
	m := gconv.Map(v)
	t, err := gtime.StrToTime(gconv.String(m["ServiceDate"]))
	if err != nil {
		return err
	}
	st.ServiceDate = MyTime{t.Time}
	return nil
}

func Test_Struct_UnmarshalValue(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		st := &MyTimeSt{}
		err := gconv.Struct(map[string]interface{}{"ServiceDate": "2020-10-10 12:00:01"}, st)
		t.Assert(err, nil)
		t.Assert(st.ServiceDate.Time.Format("2006-01-02 15:04:05"), "2020-10-10 12:00:01")
	})
	gtest.C(t, func(t *gtest.T) {
		st := &MyTimeSt{}
		err := gconv.Struct(map[string]interface{}{"ServiceDate": nil}, st)
		t.AssertNE(err, nil)
	})
}

type T struct {
	Name string
}

func (t *T) Test() string {
	return t.Name
}

type TestInterface interface {
	Test() string
}

type TestStruct struct {
	TestInterface
}

func Test_Struct_WithInterfaceAttr(t *testing.T) {
	// Implemented interface attribute.
	gtest.C(t, func(t *gtest.T) {
		v1 := TestStruct{
			TestInterface: &T{"john"},
		}
		v2 := map[string]interface{}{}
		err := gconv.StructDeep(v2, &v1)
		t.Assert(err, nil)
		t.Assert(v1.Test(), "john")
	})
	// Implemented interface attribute.
	gtest.C(t, func(t *gtest.T) {
		v1 := TestStruct{
			TestInterface: &T{"john"},
		}
		v2 := map[string]interface{}{
			"name": "test",
		}
		err := gconv.StructDeep(v2, &v1)
		t.Assert(err, nil)
		t.Assert(v1.Test(), "test")
	})
	// No implemented interface attribute.
	gtest.C(t, func(t *gtest.T) {
		v1 := TestStruct{}
		v2 := map[string]interface{}{
			"name": "test",
		}
		err := gconv.StructDeep(v2, &v1)
		t.Assert(err, nil)
		t.Assert(v1.TestInterface, nil)
	})
}

func Test_Struct_To_Struct(t *testing.T) {
	var TestA struct {
		Id   int       `p:"id"`
		Date time.Time `p:"date"`
	}

	var TestB struct {
		Id   int       `p:"id"`
		Date time.Time `p:"date"`
	}
	TestB.Id = 666
	TestB.Date = time.Now()

	gtest.C(t, func(t *gtest.T) {
		t.Assert(gconv.Struct(TestB, &TestA), nil)
		t.Assert(TestA.Id, TestB.Id)
		t.Assert(TestA.Date, TestB.Date)
	})
}
