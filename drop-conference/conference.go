package main

import (
	"strings"
	"tad-demo/common"
)

var cfg = common.NewConfig()
var db = common.NewDbClient(cfg.Service.Redis)
var restcommApi = common.NewRestcommApi(cfg.Service.Restcomm, cfg.Auth.User, cfg.Auth.Pass)

type Conference struct {
}

func (conf Conference) GetParticipants() []string {
	phonesSid := make([]string, 0)

	for _, i := range db.LRange(common.DB_KEY_URI, 0, 1000).Val() {
		uri := i[0 : len(i)-5]
		sid := uri[strings.LastIndex(uri, "/")+1 : len(uri)]
		phonesSid = append(phonesSid, sid)
	}
	return phonesSid
}

func (conf Conference) Drop() []string {
	common.Info.Println("Drop conference")
	numbers := make(map[string]bool)
	for _, i := range db.LRange(common.DB_KEY_URI, 0, 1000).Val() {
		uri := i[0 : len(i)-5]
		dropped := restcommApi.CompleteCallByUri(uri)
		if dropped {
			sid := uri[strings.LastIndex(uri, "/")+1 : len(uri)]
			to := db.Get(sid).Val()
			numbers[to] = true
			db.LRem(common.DB_KEY_URI, 0, i)
		} else {
			common.Error.Println("Can't drop call: ", uri)
		}
	}
	conf.NotifyDropChannel()
	return set(numbers)
}

func (conf Conference) NotifyDropChannel() {
	db.Publish(common.CHANNEL_CONF_DROPPED, "true")
}

func (conf Conference) NotifySms(numbers []string) {
	common.Trace.Println("NotifySms start")
	for _, to := range numbers {
		if to == "" {
			common.Info.Println("\t to is EMPTY")
			continue
		}
		common.Info.Println("Notify sms: " + to)
		db.Publish(cfg.Redis.ConfChannel, to)
	}
}

func set(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for k, _ := range values {
		keys = append(keys, k)
	}
	return keys
}
