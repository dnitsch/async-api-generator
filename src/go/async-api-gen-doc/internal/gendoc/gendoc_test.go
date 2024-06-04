package gendoc_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/dnitsch/async-api-generator/internal/gendoc"
	"github.com/dnitsch/async-api-generator/internal/token"
	log "github.com/dnitsch/simplelog"
)

func Test_GenDoc_marshalled_successfully_from_string(t *testing.T) {
	ttests := map[string]struct {
		input  string
		expect gendoc.GenDoc
	}{
		"when using full property names": {
			`parent=domain-foo~bar-assigned id=BizContextAreaEvent category=message type=example subscribers=bazquxdemand,bazquxfoo,bazquxbar`,
			gendoc.GenDoc{Id: "BizContextAreaEvent", Parent: "domain-foo~bar-assigned", CategoryType: gendoc.MessageBlock, ContentType: gendoc.Example},
		},
		"when using  shorthand property names": {
			`parent=domain-foo~bar-assigned id=BizContextAreaEvent c=message type=example sbs=bazquxdemand,bazquxfoo,bazquxbar`,
			gendoc.GenDoc{Id: "BizContextAreaEvent", Parent: "domain-foo~bar-assigned", CategoryType: gendoc.MessageBlock, ContentType: gendoc.Example},
		},
		"when setting serviceId": {
			`parent=domain-foo~bar-assigned id=BizContextAreaEvent serviceId=bazquxsample c=message type=example sbs=bazquxdemand,bazquxfoo,bazquxbar producers=bazquxsample`,
			gendoc.GenDoc{Id: "BizContextAreaEvent",
				Parent:       "domain-foo~bar-assigned",
				CategoryType: gendoc.MessageBlock,
				ContentType:  gendoc.Example,
				ServiceId:    "bazquxsample",
			},
		},
		"when setting channelId": {
			`parent=domain-foo~bar-assigned id=BizContextAreaEvent serviceId=bazquxsample channelId=bazquxsample c=message type=example sbs=bazquxdemand,bazquxfoo,bazquxbar producers=bazquxsample`,
			gendoc.GenDoc{Id: "BizContextAreaEvent",
				Parent:       "domain-foo~bar-assigned",
				CategoryType: gendoc.MessageBlock,
				ContentType:  gendoc.Example,
				ChannelId:    "bazquxsample",
				ServiceId:    "bazquxsample",
			},
		},
		"when including closing comments": {
			`parent=domain-foo~bar-assigned id=BizContextAreaEvent serviceId=bazquxsample channelId=bazquxsample c=message type=example sbs=bazquxdemand,bazquxfoo,bazquxbar producers=bazquxsample -->`,
			gendoc.GenDoc{Id: "BizContextAreaEvent",
				Parent:       "domain-foo~bar-assigned",
				CategoryType: gendoc.MessageBlock,
				ContentType:  gendoc.Example,
				ChannelId:    "bazquxsample",
				ServiceId:    "bazquxsample",
			},
		},
	}
	log := log.New(&bytes.Buffer{}, log.DebugLvl)
	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {

			got, err := gendoc.New(tt.input, log)
			if err != nil {
				t.Fatal(err)
			}
			testGenDoc(t, got, tt.expect)
		})
	}
}

func Test_Unmarshal_from_token_should_succeed(t *testing.T) {
	ttests := map[string]struct {
		input  token.Token
		expect gendoc.GenDoc
	}{
		"when using correct annotation": {
			token.Token{MetaAnnotation: `parent=domain-foo~bar-assigned id=BizContextAreaEvent serviceId=bazquxsample channelId=bazquxsample c=message type=example sbs=bazquxdemand,bazquxfoo,bazquxbar producers=bazquxsample`},
			gendoc.GenDoc{Id: "BizContextAreaEvent",
				Parent:       "domain-foo~bar-assigned",
				CategoryType: gendoc.MessageBlock,
				ContentType:  gendoc.Example,
				ChannelId:    "bazquxsample",
				ServiceId:    "bazquxsample",
			},
		},
	}
	log := log.New(&bytes.Buffer{}, log.DebugLvl)

	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			got, err := gendoc.NewFromToken(tt.input, log)
			if err != nil {
				t.Fatal(err)
			}
			testGenDoc(t, got, tt.expect)
		})
	}
}

func testGenDoc(t *testing.T, got gendoc.GenDoc, expect gendoc.GenDoc) {

	// should fail when fields are extended or changed
	val := reflect.ValueOf(got)
	if val.NumField() != 12 {
		t.Fatalf("field was added to the GenDoc struct but tests were not updated, got number of fields: %d", val.NumField())
	}

	if got.Id != expect.Id {
		t.Errorf("Id error - got: %v, expected: %v", got.Id, expect.Id)
	}
	if got.Parent != expect.Parent {
		t.Errorf("Parent error - got: %v, expected: %v", got.Parent, expect.Parent)
	}
	if got.ServiceId != expect.ServiceId {
		t.Errorf("ServiceId error - got: %v, expected: %v", got.ServiceId, expect.ServiceId)
	}
	if got.ChannelId != expect.ChannelId {
		t.Errorf("ChannelId error - got: %v, expected: %v", got.ChannelId, expect.ChannelId)
	}
}

func Test_Unmarshal_failure(t *testing.T) {
	ttests := map[string]struct {
		input string
		want  error
	}{
		"invalid key/pair without equals": {"ignored=val :notvalidKey missingEquals", gendoc.ErrUnparseableTag},
		"invalid key/pair no value":       {"notvalidKeyPair=", gendoc.ErrZeroLengthKeyOrValue},
		"invalid category specified":      {"ignored=val id=bar category=nonexistant", gendoc.ErrIncorrectCategory},
		"invalid type specified":          {"parent=foo ignored=val type=nonexistant", gendoc.ErrIncorrectType},
	}

	for name, tt := range ttests {
		t.Run(name, func(t *testing.T) {
			_, err := gendoc.New(tt.input, log.New(&bytes.Buffer{}, log.DebugLvl))
			if err == nil {
				t.Fatal(err)
			}
			if !errors.Is(err, tt.want) {
				t.Errorf("error does not include required message\n\ngot: %s\nwant: %s", err.Error(), tt.want.Error())
			}
		})
	}
}
