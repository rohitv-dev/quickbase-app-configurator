package api

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"log"
	"net/http"
	"time"
)

type App struct {
	Created                  time.Time `json:"created"`
	DateFormat               string    `json:"dateFormat"`
	Description              string    `json:"description"`
	HasEveryoneOnTheInternet bool      `json:"hasEveryoneOnTheInternet"`
	ID                       string    `json:"id"`
	Name                     string    `json:"name"`
	SecurityProperties       struct {
		AllowClone          bool `json:"allowClone"`
		AllowExport         bool `json:"allowExport"`
		EnableAppTokens     bool `json:"enableAppTokens"`
		HideFromPublic      bool `json:"hideFromPublic"`
		MustBeRealmApproved bool `json:"mustBeRealmApproved"`
		UseIPFilter         bool `json:"useIPFilter"`
	} `json:"securityProperties"`
	TimeZone           string    `json:"timeZone"`
	Updated            time.Time `json:"updated"`
	DataClassification string    `json:"dataClassification"`
}

type Table struct {
	Name               string    `json:"name"`
	Created            time.Time `json:"created"`
	Updated            time.Time `json:"updated"`
	Alias              string    `json:"alias"`
	Description        string    `json:"description"`
	ID                 string    `json:"id"`
	NextRecordID       int       `json:"nextRecordId"`
	NextFieldID        int       `json:"nextFieldId"`
	DefaultSortFieldID int       `json:"defaultSortFieldId"`
	DefaultSortOrder   string    `json:"defaultSortOrder"`
	KeyFieldID         int       `json:"keyFieldId"`
	SingleRecordName   string    `json:"singleRecordName"`
	PluralRecordName   string    `json:"pluralRecordName"`
	SizeLimit          string    `json:"sizeLimit"`
	SpaceUsed          string    `json:"spaceUsed"`
	SpaceRemaining     string    `json:"spaceRemaining"`
}

type Field struct {
	ID               int    `json:"id"`
	Label            string `json:"label"`
	FieldType        string `json:"fieldType"`
	NoWrap           bool   `json:"noWrap"`
	Bold             bool   `json:"bold"`
	Required         bool   `json:"required"`
	AppearsByDefault bool   `json:"appearsByDefault"`
	FindEnabled      bool   `json:"findEnabled"`
	Unique           bool   `json:"unique"`
	DoesDataCopy     bool   `json:"doesDataCopy"`
	FieldHelp        string `json:"fieldHelp"`
	Audited          bool   `json:"audited"`
	Properties       struct {
		PrimaryKey      bool   `json:"primaryKey"`
		ForeignKey      bool   `json:"foreignKey"`
		NumLines        int    `json:"numLines"`
		MaxLength       int    `json:"maxLength"`
		AppendOnly      bool   `json:"appendOnly"`
		AllowHTML       bool   `json:"allowHTML"`
		AllowMentions   bool   `json:"allowMentions"`
		SortAsGiven     bool   `json:"sortAsGiven"`
		CarryChoices    bool   `json:"carryChoices"`
		AllowNewChoices bool   `json:"allowNewChoices"`
		Formula         string `json:"formula"`
		DefaultValue    string `json:"defaultValue"`
	} `json:"properties"`
}

type GetTablesResponse struct {
	AppId  string
	Tables []Table
}

type GetPageResponse struct {
	XMLName   xml.Name `xml:"qdbapi"`
	ErrorCode string   `xml:"errcode"`
	ErrorText string   `xml:"errtext"`
	PageBody  string   `xml:"pagebody"`
}

type GetPageBody struct {
	XMLName   xml.Name `xml:"qdbapi"`
	UserToken string   `xml:"usertoken"`
	PageID    string   `xml:"pageID"`
}

type ReplacePageBody struct {
	XMLName   xml.Name `xml:"qdbapi"`
	UserToken string   `xml:"usertoken"`
	PageType  string   `xml:"pagetype"`
	PageID    string   `xml:"pageID"`
	PageBody  string   `xml:"pagebody"`
}

type ReplacePageResponse struct {
	XMLName   xml.Name `xml:"qdbapi"`
	Action    string   `xml:"action"`
	ErrorCode string   `xml:"errcode"`
	ErrorText string   `xml:"errtext"`
	PageID    string   `xml:"pageID"`
}

type UpdateFieldBody struct {
	XMLName   xml.Name `xml:"qdbapi"`
	FieldID   string   `xml:"fid"`
	Formula   string   `xml:"formula"`
	UserToken string   `xml:"usertoken"`
}

type UpdateFieldResponse struct {
	XMLName   xml.Name `xml:"qdbapi"`
	FieldID   string   `xml:"fid"`
	FieldName string   `xml:"fname"`
	ErrorCode string   `xml:"errcode"`
	ErrorText string   `xml:"errtext"`
}

type Quickbase struct {
	AppId     string
	UserToken string
	Realm     string
}

func (q *Quickbase) GetApp() App {
	client := http.Client{}

	req, err := http.NewRequest("GET", "https://api.quickbase.com/v1/apps/"+q.AppId, nil)

	if err != nil {
		log.Fatal(err)
	}

	req.Header = http.Header{
		"QB-Realm-Hostname": {q.Realm},
		"Authorization":     {"QB-USER-TOKEN " + q.UserToken},
	}

	res, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	var app App

	json.NewDecoder(res.Body).Decode(&app)

	return app
}

func (q *Quickbase) GetTables() GetTablesResponse {
	client := http.Client{}

	req, err := http.NewRequest("GET", "https://api.quickbase.com/v1/tables?appId="+q.AppId, nil)

	if err != nil {
		log.Fatal(err)
	}

	req.Header = http.Header{
		"QB-Realm-Hostname": {q.Realm},
		"Authorization":     {"QB-USER-TOKEN " + q.UserToken},
	}

	res, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	var tables []Table

	json.NewDecoder(res.Body).Decode(&tables)

	return GetTablesResponse{
		AppId:  q.AppId,
		Tables: tables,
	}
}

func (q *Quickbase) GetPage(pageId string) GetPageResponse {
	client := http.Client{}

	xmlBody, err := xml.MarshalIndent(GetPageBody{
		UserToken: q.UserToken,
		PageID:    pageId,
	}, " ", "  ")

	if err != nil {
		log.Fatal(err)
	}

	body := bytes.NewReader(xmlBody)

	req, err := http.NewRequest("POST", "https://"+q.Realm+"/db/"+q.AppId, body)

	if err != nil {
		log.Fatal(err)
	}

	req.Header = http.Header{
		"Content-Type":     {"application/xml"},
		"QUICKBASE-ACTION": {"API_GetDBPage"},
	}

	res, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	var pageResponse GetPageResponse

	xml.NewDecoder(res.Body).Decode(&pageResponse)

	if pageResponse.ErrorCode != "0" {
		log.Fatal(pageResponse.ErrorText)
	}

	return pageResponse
}

func (q *Quickbase) ReplacePage(pageId string, pageBody string) ReplacePageResponse {
	client := http.Client{}

	xmlBody, err := xml.MarshalIndent(ReplacePageBody{
		UserToken: q.UserToken,
		PageType:  "1",
		PageID:    pageId,
		PageBody:  pageBody,
	}, " ", "  ")

	if err != nil {
		log.Fatal(err)
	}

	body := bytes.NewReader(xmlBody)

	req, err := http.NewRequest("POST", "https://"+q.Realm+"/db/"+q.AppId, body)

	if err != nil {
		log.Fatal(err)
	}

	req.Header = http.Header{
		"Content-Type":     {"application/xml"},
		"QUICKBASE-ACTION": {"API_AddReplaceDBPage"},
	}

	res, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	var response ReplacePageResponse

	xml.NewDecoder(res.Body).Decode(&response)

	if response.ErrorCode != "0" {
		log.Fatal(response.ErrorText)
	}

	return response
}

func (q *Quickbase) GetFields(tableId string) []Field {
	client := http.Client{}

	req, err := http.NewRequest("GET", "https://api.quickbase.com/v1/fields?tableId="+tableId, nil)

	if err != nil {
		log.Fatal(err)
	}

	req.Header = http.Header{
		"QB-Realm-Hostname": {q.Realm},
		"Authorization":     {"QB-USER-TOKEN " + q.UserToken},
	}

	res, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	var fields []Field

	json.NewDecoder(res.Body).Decode(&fields)

	return fields
}

func (q *Quickbase) UpdateField(tableId string, fieldId string, formula string) UpdateFieldResponse {
	client := http.Client{}

	xmlBody, err := xml.MarshalIndent(UpdateFieldBody{
		UserToken: q.UserToken,
		FieldID:   fieldId,
		Formula:   "<![CDATA[" + formula + "]]",
	}, " ", "  ")

	if err != nil {
		log.Fatal(err)
	}

	body := bytes.NewReader(xmlBody)

	req, err := http.NewRequest("POST", "https://"+q.Realm+"/db/"+tableId, body)

	if err != nil {
		log.Fatal(err)
	}

	req.Header = http.Header{
		"Content-Type":     {"application/xml"},
		"QUICKBASE-ACTION": {"API_SetFieldProperties"},
	}

	res, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	var response UpdateFieldResponse

	xml.NewDecoder(res.Body).Decode(&response)

	if response.ErrorCode != "0" {
		log.Fatal(response.ErrorText)
	}

	return response
}
