package ci

import (
	"fmt"
	"reflect"
	"testing"

	"gitlab.cern.ch/lb-experts/goermis/api"
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
	if api.ContainsAlarm(alarms, NonExistingAlarm) {
		t.Errorf("Could find alarm %v even though it doesn't exist", NonExistingAlarm)

	}
}

func TestContainsNode(t *testing.T) {
	type test struct {
		input             *api.Relation
		expectedName      bool
		expectedBlacklist bool
	}

	testCases := []test{
		{
			input: &api.Relation{
				Blacklist: false,
				Node: &api.Node{
					NodeName: "test1.cern.ch",
				},
			},
			expectedName:      true,
			expectedBlacklist: true,
		},

		{
			input: &api.Relation{
				Blacklist: true,
				Node: &api.Node{
					NodeName: "test1.cern.ch",
				},
			},
			expectedName:      true,
			expectedBlacklist: false,
		},
		{
			input: &api.Relation{
				Blacklist: true,
				Node: &api.Node{
					NodeName: "test12.cern.ch",
				},
			},
			expectedName:      false,
			expectedBlacklist: false,
		},
	}

	relations := []*api.Relation{
		{
			Blacklist: false,
			Node: &api.Node{
				NodeName: "test1.cern.ch",
			},
		},
		{
			Blacklist: true,
			Node: &api.Node{
				NodeName: "testme.cern.ch",
			},
		},
		{
			Blacklist: false,
			Node: &api.Node{
				NodeName: "test56.cern.ch",
			},
		},
	}

	for _, tc := range testCases {
		if name, bl := api.ContainsNode(relations, tc.input); name != tc.expectedName || bl != tc.expectedBlacklist {
			t.Errorf("We did not receive what we expected for %v\nWe received: %v, %v\nWe expected: %v, %v", tc.input.Node.NodeName, name, bl, tc.expectedName, tc.expectedBlacklist)
		}
	}

}

/*For testing FindNodeID and FindAliasID, there is an alias
manually created beforehand */
func TestFindAliasID(t *testing.T) {
	type alias struct {
		input    string
		expected int
	}
	aliases := []alias{
		//legit alias
		{input: "seed.cern.ch", //seed entry , which is used also for tests
			expected: 84},
		//non existing alias
		{input: "idontexist.cern.ch",
			expected: 0},
	}

	for _, alias := range aliases {
		output := api.FindAliasID(alias.input)
		if output != alias.expected {
			t.Errorf("Failed to find the correct alias ID.\nExpected:%v\nReceived:%v\n", alias.expected, output)
		}
	}
}

func TestFindNodeID(t *testing.T) {
	type node struct {
		input    string
		expected int
	}
	nodes := []node{
		//legit node
		{input: "testnode.cern.ch", //testnode is declared for alias seed.cern.ch
			expected: 118},
		//non existing node
		{input: "nonexistent.cern.ch",
			expected: 0},
	}

	for _, node := range nodes {
		output := api.FindNodeID(node.input)
		if output != node.expected {
			t.Errorf("Failed to find the correct alias ID.\nExpected:%v\nReceived:%v\n", node.expected, output)
		}
	}
}

func TestDeleteEmpty(t *testing.T) {
	type slice struct {
		input    []string
		expected []string
	}
	slices := []slice{
		{
			input:    []string{"a", "b", "c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{
			input:    []string{"a", "", "", "d"},
			expected: []string{"a", "d"},
		},
	}
	for _, slice := range slices {
		output := api.DeleteEmpty(slice.input)
		if !reflect.DeepEqual(output, slice.expected) {
			t.Errorf("Failed test for DeleteEmpty\nExpected:%v\nReceived:%v\n", slice.expected, output)

		}

	}
}
