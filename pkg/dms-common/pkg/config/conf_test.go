package config

import (
	"strconv"
	"testing"
)

const confFile = `
[default]
host = example.com
port = 43
compression = on
active = false

[service-1]
port = 443
`

//url = http://%(host)s/something

type stringtest struct {
	section string
	option  string
	answer  string
}

type inttest struct {
	section string
	option  string
	answer  int
}

type booltest struct {
	section string
	option  string
	answer  bool
}

var testSet = []interface{}{
	stringtest{"", "host", "example.com"},
	inttest{"default", "port", 43},
	booltest{"default", "compression", true},
	booltest{"default", "active", false},
	inttest{"service-1", "port", 443},
	//stringtest{"service-1", "url", "http://example.com/something"},
}

func TestBuild(t *testing.T) {
	c, err := ReadConfigBytes([]byte(confFile))
	if err != nil {
		t.Error(err)
	}

	for _, element := range testSet {
		switch e := element.(type) {
		case stringtest:
			ans, err := c.GetString(e.section, e.option)
			if err != nil {
				t.Error("c.GetString(\"" + e.section + "\",\"" + e.option + "\") returned error: " + err.Error())
			} else if ans != e.answer {
				t.Error("c.GetString(\"" + e.section + "\",\"" + e.option + "\") returned incorrect answer: " + ans)
			}
		case inttest:
			ans, err := c.GetInt(e.section, e.option)
			if err != nil {
				t.Error("c.GetInt(\"" + e.section + "\",\"" + e.option + "\") returned error: " + err.Error())
			} else if ans != e.answer {
				t.Error("c.GetInt(\"" + e.section + "\",\"" + e.option + "\") returned incorrect answer: " + strconv.Itoa(ans))
			}
		case booltest:
			ans, err := c.GetBool(e.section, e.option)
			if err != nil {
				t.Error("c.GetBool(\"" + e.section + "\",\"" + e.option + "\") returned error: " + err.Error())
			} else if ans != e.answer {
				t.Error("c.GetBool(\"" + e.section + "\",\"" + e.option + "\") returned incorrect answer")
			}
		}
	}
}

var (
	byteConfigForTest = `[default1]
host = something.com
port = 443
active = true  
compression = off

[service-2]
port = 444
skip_lock
fake_key = value1
host = something.com
fake_key = value

[service-1]#comment1
compression = on
`
)

func TestRemoveOption(t *testing.T) {
	t.Logf("conf is like before :\n%v\n", byteConfigForTest)
	conf, err := ReadConfigBytes([]byte(byteConfigForTest))
	assertNoError(t, err)
	success := conf.RemoveOption("service-2", "fake_key")
	if !success {
		t.Fatalf("failed to remove fake_key")
	}
	options, err := conf.GetOptions("service-2")
	assertNoError(t, err)
	if 3 != len(options) {
		t.Fatalf("expect 3 but get: %v", len(options))
	}
	for _, option := range options {
		if option == "fake_key" {
			t.Fatalf("option: fake_key should be removed")
		}
	}
}

func TestAddOption(t *testing.T) {
	t.Logf("conf is like before :\n%v\n", byteConfigForTest)
	conf, err := ReadConfigBytes([]byte(byteConfigForTest))
	assertNoError(t, err)
	isNotOverwritten := conf.AddOption("service-2", "fake_key", "fakeValue")
	if isNotOverwritten {
		t.Fatalf("add existed should be overwritten")
	}
	t.Logf("conf is like after :\n%v\n", conf.String())
	options, err := conf.GetOptions("service-2")
	assertNoError(t, err)
	if 5 != len(options) {
		t.Fatalf("expect 5 but get: %v", len(options))
	}

	kvDataMapList, isExisted := conf.GetKVDatas("service-2")
	if !isExisted {
		t.Fatalf("should get section service-2")
	}
	for _, kvDataMap := range kvDataMapList {
		_, isExisted := kvDataMap["fake_key"]
		if isExisted && kvDataMap["fake_key"] != "fakeValue" {
			t.Fatalf("expect: fakeValue but get: %v", kvDataMap["fake_key"])
		}
	}

	isNotOverwritten = conf.AddOption("service-2", "fake_key1", "fakeValue1")
	if !isNotOverwritten {
		t.Fatalf("add existed should not be overwritten")
	}

	value, err := conf.GetString("service-2", "fake_key1")
	if err != nil {
		t.Fatalf("failed to get key fake_key1")
	}
	if value != "fakeValue1" {
		t.Fatalf("expect: fakeValue1 but get: %v", value)
	}
}

var (
	byteConfigGetStringTest = `[default1]
host = something.com
host1 = "something.com"
host2 = 'something.com"
host3 = 'something.com'
host4 = "'something.com"
`
)

func TestGetString(t *testing.T) {
	t.Logf("conf is like before :\n%v\n", byteConfigGetStringTest)
	conf, err := ReadConfigBytes([]byte(byteConfigGetStringTest))
	assertNoError(t, err)
	options, err := conf.GetOptions("default1")
	assertNoError(t, err)
	for _, option := range options {
		value, err := conf.GetString("default1", option)
		assertNoError(t, err)
		if option == "host4" {
			if value != "'something.com" {
				t.Fatalf("expect: something.com but get: %v", value)
			}
			continue
		}

		if value != "something.com" {
			t.Fatalf("expect: something.com but get: %v", value)
		}
	}
}
