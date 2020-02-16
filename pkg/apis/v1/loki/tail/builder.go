package tail

import (
	"encoding/json"
	"io/ioutil"
	"github.com/klovercloud-dev/klovercloud-loki-client/config"
	"github.com/klovercloud-dev/klovercloud-loki-client/pkg/apis/common"
	"log"
	"net/http"
	"strconv"
)

const (
	URL = "loki/api/v1/tail?"
)


type QueryResponse struct {
	Streams     []Streams `json:"streams"`
	DroppedEntries     []DroppedEntry `json:"dropped_entries"`
}

type DroppedEntry struct {
	Labels     LabelSet  `json:"labels"`
	Timestamp int64  `json:"timestamp"`
}

type  Streams struct {
	Stream     LabelSet  `json:"stream"`
	Values     []interface{} `json:"values"`
}

type LabelSet map[string]string


type Builder interface {
	Get() Builder
	Post() Builder
	Query() Builder
	Init() Builder
	Label(levelName string, value string) Builder
	Limit(int) Builder
	Sum() Builder
	Rate(int) Builder
	Start(int64) Builder
	Delay(int) Builder
	Contains(string) Builder
	NotContains(string) Builder
	Matches(expression string) Builder
	NotMatches(expression string) Builder
	CountOverTime(minutes int) Builder
	TopK(k int64) Builder
	Build() Builder
	Fire() QueryResponse

}

type builder struct {
	method string
	url    string
	labels map[string]string
	start  int64
	limit  int
	delay  int
	body   interface{}
	query  string
	pipe   string
}

func (qb *builder) Fire() QueryResponse {
	client := &http.Client{}
	req, err := http.NewRequest(qb.method, qb.url, nil)
	req.SetBasicAuth(config.Username, config.Password)
	req.Header.Add("Content-Type","application/json")
	log.Println("Requesting:",qb.url)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	response :=QueryResponse{}
	json.Unmarshal([]byte(body), &response)
	return response
}

func (qb *builder) TopK(k int64) Builder {
	temp:=""
	if(qb.query==""){
		temp=createKeyValuePairs(qb.labels)+qb.pipe
	}else{
		temp=qb.query
	}
	qb.query="topk("+strconv.Itoa(int(k))+","+temp+")"
	return qb
}


func (qb *builder) CountOverTime(minutes int) Builder {
	temp:=""
	if(qb.query==""){
		temp=createKeyValuePairs(qb.labels)+qb.pipe
	}else{
		temp=qb.query
	}
	qb.query="count_over_time("+temp+"["+strconv.Itoa(int(minutes))+"m])"
	return qb
}

func (qb *builder) Contains(str string) Builder {
	qb.pipe=qb.pipe+"|="+"\""+str+"\""
	return qb
}

func (qb *builder) NotContains(str string) Builder {
	qb.pipe=qb.pipe+"!="+"\""+str+"\""
	return qb
}

func (qb *builder) Matches(expression string) Builder {
	qb.pipe=qb.pipe+"|~"+"\""+expression+"\""
	return qb
}

func (qb *builder) NotMatches(expression string) Builder {
	qb.pipe=qb.pipe+"!~"+"\""+expression+"\""
	return qb
}

func (qb *builder) Delay(seconds int) Builder {
	qb.delay=seconds
	return qb
}

func (qb *builder) Start(start int64) Builder {
	qb.start=start
	return qb
}



func (qb *builder) Sum() Builder {
	temp:=""
	if(qb.query==""){
		temp=createKeyValuePairs(qb.labels)
	}else{
		temp=qb.query
	}
	qb.query="sum("+temp+")"
	return qb
}

func (qb *builder) Rate(minutes int) Builder {
	temp:=""
	if(qb.query==""){
		temp=createKeyValuePairs(qb.labels)+qb.pipe
	}else{
		temp=qb.query
	}
	qb.query="rate("+temp+"["+strconv.Itoa(int(minutes))+"m])"
	return qb
}

func (qb *builder) Init() Builder {
	qb.labels = make(map[string]string)
	return qb
}


func (qb *builder) Limit(limit int) Builder {
	qb.limit = limit
	return qb
}

func (qb *builder) Label(levelName string, value string) Builder {
	qb.labels[levelName] = value
	return qb
}

func (qb *builder) Query() Builder {
	qb.url = qb.url + "query="
	return qb
}

func (qb *builder) Get() Builder {
	qb.method = common.GET
	qb.url = config.LokiWSUrl + URL
	return qb
}

func (qb *builder) Post() Builder {
	qb.method = common.POST
	return qb
}

func (qb *builder) Build() Builder {
	str:=""
	if(qb.query==""){
		str= qb.url + "query=" + createKeyValuePairs(qb.labels)
	}else{
		str=qb.url + "query=" +qb.query
	}

	if qb.limit != 0 {
		str = str + "&limit=" + strconv.Itoa(int(qb.limit))
	}

	if qb.start != 0 {
		str = str + "&start=" + strconv.FormatInt(qb.start, 10)
	}

	if qb.delay != 0 {
		str = str + "&delay_for=" + strconv.Itoa(qb.delay)
	}

	qb.url = str
	return qb
}

func NewUrlBuilder() Builder {
	return &builder{}
}

func createKeyValuePairs(m map[string]string) string {
	count := 0
	str := "{"
	for key, value := range m {
		count++
		if count < len(m) {
			str = str + key + "=" + "\"" + value + "\"" + ","
		} else {
			str = str + key + "=" + "\"" + value + "\"" + "}"
		}
	}
	return str
}

