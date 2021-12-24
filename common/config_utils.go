package common

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/yedf/dtm/dtmcli/dtmimp"
)

func loadFromEnv(prefix string, conf interface{}) {
	rv := reflect.ValueOf(conf)
	dtmimp.PanicIf(rv.Kind() != reflect.Ptr || rv.IsNil(),
		fmt.Errorf("should be a valid pointer, but %s found", reflect.TypeOf(conf).Name()))
	loadFromEnvInner(prefix, rv.Elem(), "")
}

func loadFromEnvInner(prefix string, conf reflect.Value, defaultValue string) {
	kind := conf.Kind()
	switch kind {
	case reflect.Struct:
		t := conf.Type()
		for i := 0; i < t.NumField(); i++ {
			tag := t.Field(i).Tag
			loadFromEnvInner(prefix+"_"+tag.Get("yaml"), conf.Field(i), tag.Get("default"))
		}
	case reflect.String:
		str := os.Getenv(toUnderscoreUpper(prefix))
		if str == "" {
			str = defaultValue
		}
		conf.Set(reflect.ValueOf(str))
	case reflect.Int64:
		str := os.Getenv(toUnderscoreUpper(prefix))
		if str == "" {
			str = defaultValue
		}
		if str == "" {
			str = "0"
		}
		conf.Set(reflect.ValueOf(int64(dtmimp.MustAtoi(str))))
	default:
		panic(fmt.Errorf("unsupported type: %s", conf.Type().Name()))
	}
}

func toUnderscoreUpper(key string) string {
	key = strings.Trim(key, "_")
	matchFirstCap := regexp.MustCompile("([a-z])([A-Z]+)")
	s2 := matchFirstCap.ReplaceAllString(key, "${1}_${2}")
	// logger.Debugf("loading from env: %s", strings.ToUpper(s2))
	return strings.ToUpper(s2)
}
