package api

import (
	"fmt"
	"reflect"
	"testing"

	api "gitlab.cern.ch/lb-experts/goermis/api"
)

func TestContainsCname(t *testing.T) {
	ExistingCname := "cname1"
	NonExistingCname := "cname2"
	cnames := []api.Cname{
		{Cname: "cname1"},
		{Cname: "cname-dfs"},
		{Cname: "cname-test"},
	}
	fmt.Println("Now we will test if the existing cname can be found")
	if !api.ContainsCname(cnames, ExistingCname) {
		t.Errorf("Could not find cname %v even though it exists", ExistingCname)

	}
	fmt.Println("Now we will test it with a non-existing cname")
	if api.ContainsCname(cnames, NonExistingCname) {
		t.Errorf("We found the cname %v even though it should not exist", NonExistingCname)

	}
}

func TestExplode(t *testing.T) {
	type data struct {
		encode   string
		data     []string
		expected []string
	}
	tests := []data{
		{"application/json", []string{"test1", "test2", "test3"}, []string{"test1", "test2", "test3"}},
		{"application/x-www-form-urlencoded", []string{"test1,test2,test3"}, []string{"test1", "test2", "test3"}},
		{"random", []string{"test1", "test2", "test3"}, []string{}},
		{"random", []string{"test1,test2,test3"}, []string{}},
	}
	for _, test := range tests {
		if !reflect.DeepEqual(api.Explode(test.encode, test.data), test.expected) {
			t.Errorf("Failed test in TestExplode. Expected: %v\n Received: %v\n", test.expected, api.Explode(test.encode, test.data))
		}
	}

}

func TestContainsAlarm(t *testing.T) {
	ExistingAlarm := api.Alarm{
		Name:      "minimum",
		Recipient: "lb-experts@cern.ch",
		Parameter: 1,
	}
	NonExistingAlarm := api.Alarm{
		Name:      "minimum",
		Recipient: "it-dep@cern.ch",
		Parameter: 10,
	}
	alarms := []api.Alarm{
		{
			Name:      "minimum",
			Recipient: "lb-experts@cern.ch",
			Parameter: 1,
		},
		{
			Name:      "maximum",
			Recipient: "lb-experts@cern.ch",
			Parameter: 1,
		}, {
			Name:      "minimum",
			Recipient: "ermis-experts@cern.ch",
			Parameter: 1,
		},
		{
			Name:      "minimum",
			Recipient: "lb-experts@cern.ch",
			Parameter: 2,
		},
	}
	fmt.Println("Now we will test if the alarm can be found")
	if !api.ContainsAlarm(alarms, ExistingAlarm) {
		t.Errorf("Could not find alarm %v even though it exists", ExistingAlarm)

	}
	fmt.Println("Now we will test for the non existing alarm")
	if  api.ContainsAlarm(alarms, NonExistingAlarm) {
		t.Errorf("Could find alarm %v even though it doesn't exist", NonExistingAlarm)

	}
}

func TestContainsNode(t *testing.T) {

}
