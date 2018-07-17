package helpers

import (
	"log"
	"os"
	"gopkg.in/mgo.v2/bson"
	"encoding/base64"
	"strconv"
	"time"
	"reflect"
)

var logger = log.New(os.Stderr, "app: ", log.LstdFlags) //| log.Llongfile)

func CheckErr(err error, msg string) {
	if err != nil {
		logger.Println(msg, err)
	}
}

func ObjectIdFromString(encodedid string) (bson.ObjectId, error) {
	data, err := base64.URLEncoding.DecodeString(encodedid)
	if err != nil {
		log.Println("Error decoding object ID:", err)
		return bson.NewObjectId(), err
	}
	return bson.ObjectId(data), err
}

func RenderPercent(m float64) string {
	if m < 1.0 {
		return strconv.FormatFloat(m*100, 'f', 2, 64) + "%"
	}
	return strconv.FormatFloat(m, 'f', 2, 64) + "%"
}

func ConvertDate(value string) reflect.Value {
	if date, err := time.Parse("01/02/2006", value); err == nil {
		return reflect.ValueOf(date)
	}
	return reflect.Value{}
}

func RenderTimestamp(value time.Time) string {
	if !value.IsZero() {
		location, err := time.LoadLocation("America/New_York")
		local := value.UTC()
		if err == nil {
			local = local.In(location)
		}
		layout := "01/02/2006 03:04:05 PM"
		return local.Format(layout)
	}
	return ""
}

func RenderTimeDetails(value time.Time) string {
	if !value.IsZero() {
		location, err := time.LoadLocation("America/New_York")
		local := value.UTC()
		if err == nil {
			local = local.In(location)
		}
		layout := "3:04 PM on Jan 2"
		return local.Format(layout)
	}
	return ""
}

func GetIDEncoded(o bson.ObjectId) string {
	return base64.URLEncoding.EncodeToString([]byte(o))
}