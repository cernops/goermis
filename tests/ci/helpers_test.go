package ci

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/asaskevich/govalidator"
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
	type test struct {
		caseID   int
		encode   string
		input    []string
		expected []string
	}
	testCases := []test{
		{1, "application/json", []string{"test1", "test2", "test3"}, []string{"test1", "test2", "test3"}},
		{2, "application/x-www-form-urlencoded", []string{"test1,test2,test3"}, []string{"test1", "test2", "test3"}},
		{3, "random", []string{"test1", "test2", "test3"}, []string{}},
		{4, "random", []string{"test1,test2,test3"}, []string{}},
	}
	for _, tc := range testCases {
		output := api.Explode(tc.encode, tc.input)
		if !reflect.DeepEqual(output, tc.expected) {
			t.Errorf("Failed test in TestExplode.\nFAILED CASE ID:%v\nINPUT:\n%v\nEXPECTED: \n%v\nRECEIVED: \n%v\n", tc.caseID, tc.input, tc.expected, output)
		}
	}

}

func TestContainsAlarm(t *testing.T) {
	type test struct {
		caseID   int
		input    api.Alarm
		expected bool
	}

	testCases := []test{
		{
			caseID: 1,
			input: api.Alarm{
				Name:      "minimum",
				Recipient: "lb-experts@cern.ch",
				Parameter: 1,
			},
			expected: true,
		},
		{
			caseID: 2,
			input: api.Alarm{
				Name:      "minimum",
				Recipient: "it-dep@cern.ch",
				Parameter: 10,
			},
			expected: false,
		},
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
		},
	}
	fmt.Println("Now we will test if the alarm can be found")
	for _, tc := range testCases {
		output := api.ContainsAlarm(alarms, tc.input)
		if output != tc.expected {
			t.Errorf("Failed in TestContainsAlarm\nFAILED CASE ID:%v\nI\n%v\nEXPECTED:\n%v\nBut RECEIVED:\n%v\n", tc.caseID, tc.input, tc.expected, output)
		}
	}
}

func TestContainsNode(t *testing.T) {
	type test struct {
		caseID            int
		input             api.Relation
		expectedName      bool
		expectedBlacklist bool
	}

	testCases := []test{
		{caseID: 1,
			input: api.Relation{
				Blacklist: false,
				Node: &api.Node{
					NodeName: "test1.cern.ch",
				},
			},
			expectedName:      true,
			expectedBlacklist: true,
		},

		{caseID: 2,
			input: api.Relation{
				Blacklist: true,
				Node: &api.Node{
					NodeName: "test1.cern.ch",
				},
			},
			expectedName:      true,
			expectedBlacklist: false,
		},
		{caseID: 3,
			input: api.Relation{
				Blacklist: true,
				Node: &api.Node{
					NodeName: "test12.cern.ch",
				},
			},
			expectedName:      false,
			expectedBlacklist: false,
		},
	}

	relations := []api.Relation{
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
		if output1, output2 := api.ContainsNode(relations, tc.input); output1 != tc.expectedName || output2 != tc.expectedBlacklist {
			t.Errorf("We did not receive what we expected for %v\nFAILED CASE ID:%v\nWE RECEIVED:\n %v, %v\nWE EXPECTED:\n %v,\n %v\n", tc.input.Node.NodeName, tc.caseID, output1, output2, tc.expectedName, tc.expectedBlacklist)
		}
	}

}

func TestFindNodeID(t *testing.T) {
	type test struct {
		caseID   int
		input    string
		expected int
	}
	testCases := []test{
		//legit node
		{caseID: 1,
			input:    "testnode.cern.ch", //testnode is declared for alias seed.cern.ch
			expected: 118},
		//non existing node
		{caseID: 2,
			input:    "nonexistent.cern.ch",
			expected: 0},
	}
	//Slice of relations that will be searched
	relations := []api.Relation{
		{
			NodeID: 118,
			Node: &api.Node{
				NodeName: "testnode.cern.ch",
			},
		},
	}

	for _, tc := range testCases {
		output := api.FindNodeID(tc.input, relations)
		if output != tc.expected {
			t.Errorf("Failed to find the correct alias ID.\nFAILED CASE ID:%v\nINPUT:\n%v\nEXPECTED:\n%v\nRECEIVED:\n%v\n", tc.caseID, tc.input, tc.expected, output)
		}
	}
}

func TestDeleteEmpty(t *testing.T) {
	type test struct {
		caseID   int
		input    []string
		expected []string
	}
	testCases := []test{
		{caseID: 1,
			input:    []string{"a", "b", "c", "d"},
			expected: []string{"a", "b", "c", "d"},
		},
		{caseID: 2,
			input:    []string{"a", "", "", "d"},
			expected: []string{"a", "d"},
		},
	}
	for _, tc := range testCases {
		output := api.DeleteEmpty(tc.input)
		if !reflect.DeepEqual(output, tc.expected) {
			t.Errorf("Failed test for DeleteEmpty\nFAILED CASE ID:%v\nINPUT:\n%v\nEXPECTED:\n%v\nRECEIVED:\n%v\n", tc.caseID, tc.input, tc.expected, output)

		}

	}
}

func TestStringInSlice(t *testing.T) {
	type test struct {
		caseID   int
		input1   string
		input2   []string
		expected bool
	}
	testCases := []test{
		{caseID: 1,
			input1:   "string1",
			input2:   []string{"string1", "string2", "string3"},
			expected: true},
		{caseID: 2,
			input1:   "string4",
			input2:   []string{"string1", "string2", "string3"},
			expected: false},
	}
	for _, tc := range testCases {
		output := api.StringInSlice(tc.input1, tc.input2)
		if !output == tc.expected {
			t.Errorf("Failed in StringInSlice\nFAILED CASE ID:%v\nINPUTS:\n%v\n%v\nEXPECTED:\n%v\nRECEIVED:\n%v\n", tc.caseID, tc.input1, tc.input2, tc.expected, output)
		}
	}
}
func TestEqualCnames(t *testing.T) {
	type test struct {
		caseID   int
		input1   []api.Cname
		input2   []api.Cname
		expected bool
	}
	testCases := []test{
		//Case 1: Same elements / Same order
		{caseID: 1,
			input1: []api.Cname{
				{Cname: "cname1"},
				{Cname: "cname2"},
				{Cname: "cname3"},
			},
			input2: []api.Cname{
				{Cname: "cname1"},
				{Cname: "cname2"},
				{Cname: "cname3"},
			},
			expected: true,
		},
		//Case 2: Less elements on input1 / Same order
		{caseID: 2,
			input1: []api.Cname{
				{Cname: "cname1"},
				{Cname: "cname2"},
			},
			input2: []api.Cname{
				{Cname: "cname1"},
				{Cname: "cname2"},
				{Cname: "cname3"},
			},
			expected: false,
		},
		//Case 3: One different element
		{caseID: 3,
			input1: []api.Cname{
				{Cname: "cname1"},
				{Cname: "cname2"},
				{Cname: "cname4"},
			},
			input2: []api.Cname{
				{Cname: "cname1"},
				{Cname: "cname2"},
				{Cname: "cname3"},
			},
			expected: false,
		},
		//Case 4: Same elements, different order
		{caseID: 4,
			input1: []api.Cname{
				{Cname: "cname2"},
				{Cname: "cname1"},
				{Cname: "cname3"},
			},
			input2: []api.Cname{
				{Cname: "cname1"},
				{Cname: "cname2"},
				{Cname: "cname3"},
			},
			expected: true,
		},
		//Case 5: Missing element from input2
		{caseID: 5,
			input1: []api.Cname{
				{Cname: "cname1"},
				{Cname: "cname2"},
				{Cname: "cname4"},
			},
			input2: []api.Cname{
				{Cname: "cname1"},
				{Cname: "cname2"},
			},
			expected: false,
		},
		//Case 6: Empty inputs
		{caseID: 6,
			input1:   []api.Cname{},
			input2:   []api.Cname{},
			expected: true,
		},
	}

	for _, tc := range testCases {
		output := api.EqualCnames(tc.input1, tc.input2)
		if output != tc.expected {
			t.Errorf("Failed for TestEqual\nFAILED CASE ID:%v\nINPUTS:\n%v\n%v\nEXPECTED:\n%v\nRECEIVED:\n%v\n", tc.caseID, tc.input1, tc.input2, tc.expected, output)
		}
	}
}

func TestValidation(t *testing.T) {
	type test struct {
		caseID   int
		input    api.Alias
		expected bool
	}

	testCases := []test{
		//Case1: Correct fields
		{caseID: 1,
			input: api.Alias{
				ID:               1,
				AliasName:        "seed.cern.ch",
				Behaviour:        "mindless",
				BestHosts:        1,
				External:         "no",
				Metric:           "cmsfrontier",
				PollingInterval:  300,
				Statistics:       "long",
				Clusters:         "none",
				Tenant:           "golang",
				Hostgroup:        "aiermis",
				User:             "kkouros",
				TTL:              60,
				LastModification: time.Now(),
				Cnames: []api.Cname{
					{
						ID:           1,
						CnameAliasID: 1,
						Cname:        "cname1",
					},
				},
				Relations: []api.Relation{
					{
						ID:        1,
						NodeID:    2,
						Blacklist: true,
						AliasID:   1,
						Node: &api.Node{
							ID:               2,
							NodeName:         "testnode.cern.ch",
							Hostgroup:        "aiermis",
							LastModification: time.Now(),
							Load:             0,
							State:            "",
						},
						Alias: nil,
					},
				},
				Alarms: []api.Alarm{
					{
						ID:           1,
						AlarmAliasID: 1,
						Alias:        "seed.cern.ch",
						Name:         "minimum",
						Recipient:    "lb-experts@cern.ch",
						Parameter:    1,
					},
				},
			},
			expected: true,
		},
		//Case 2: Wrong alias name
		{caseID: 2,
			input: api.Alias{
				AliasName: "$%#$", //wrong field
				BestHosts: 1,
				External:  "no",
				Hostgroup: "aiermis",
			},
			expected: false,
		},
		//Case 3: Empty alias name
		{caseID: 3,
			input: api.Alias{
				AliasName: "", //empty field
				BestHosts: 1,
				External:  "no",
				Hostgroup: "aiermis",
			},
			expected: false,
		},
		//Case 4: Wrong Behaviour field
		{caseID: 4,
			input: api.Alias{
				AliasName: "alias.cern.ch",
				Behaviour: "*", //wrong field
				BestHosts: 1,
				External:  "no",
				Hostgroup: "aiermis",
			},
			expected: false,
		},
		//Case 5: Wrong Best Hosts field
		{caseID: 5,
			input: api.Alias{
				AliasName: "alias.cern.ch",
				BestHosts: -300, //wrong field
				External:  "no",
				Hostgroup: "aiermis",
			},
			expected: false,
		},
		//Case 6: Wrong external  field
		{caseID: 6,
			input: api.Alias{
				AliasName: "alias.cern.ch",
				BestHosts: 1,
				External:  "random", //wrong field
				Hostgroup: "aiermis",
			},
			expected: false,
		},
		//Case 7: Empty external
		{caseID: 7,
			input: api.Alias{
				AliasName: "alias.cern.ch",

				BestHosts: 1,
				External:  "", //empty field
				Hostgroup: "aiermis",
			},
			expected: false,
		},
		//Case 8: Wrong metric
		{caseID: 8,
			input: api.Alias{
				AliasName: "alias.cern.ch",
				BestHosts: 1,
				External:  "no",
				Metric:    "random", //wrong field
				Hostgroup: "aiermis",
			},
			expected: false,
		},
		//Case 9: Empty Hostgroup
		{caseID: 9,
			input: api.Alias{
				AliasName: "alias.cern.ch",
				BestHosts: 1,
				External:  "no",
				Hostgroup: "", //empty field
			},
			expected: false,
		},
		//Case 10: With sub-hostgroup
		{caseID: 10,
			input: api.Alias{
				AliasName: "alias.cern.ch",
				BestHosts: 1,
				External:  "no",
				Hostgroup: "aiermis/test/hg", //with sub-hostgroups
			},
			expected: true,
		},

		//Case 11: Malformed Hostgroup
		{caseID: 11,
			input: api.Alias{
				AliasName: "alias.cern.ch",
				BestHosts: 1,
				External:  "no",
				Hostgroup: "ta@", //malformed
			},
			expected: false,
		},

		//Case 12: Malformed Cname
		{caseID: 12,
			input: api.Alias{
				AliasName: "alias.cern.ch",
				BestHosts: 1,
				External:  "no",
				Hostgroup: "aiermis",
				Cnames: []api.Cname{
					{
						ID:           1,
						CnameAliasID: 1,
						Cname:        "%@q", //malformed field
					},
				},
			},
			expected: false,
		},
		//Case 13: malformed alias in alarms
		{caseID: 13,
			input: api.Alias{
				AliasName: "alias.cern.ch",
				BestHosts: 1,
				External:  "no",
				Hostgroup: "aiermis",
				Alarms: []api.Alarm{
					{
						ID:           1,
						AlarmAliasID: 1,
						Alias:        "@*tes!", //bad field
						Name:         "minimum",
						Recipient:    "lb-experts@cern.ch",
						Parameter:    1,
					},
				},
			},
			expected: false,
		},
		//Case 14: Random alarm name
		{caseID: 14,
			input: api.Alias{
				AliasName: "alias.cern.ch",
				BestHosts: 1,
				External:  "no",
				Hostgroup: "aiermis",
				Alarms: []api.Alarm{
					{
						ID:           1,
						AlarmAliasID: 1,
						Alias:        "alias.cern.ch",
						Name:         "random", //not specified field
						Recipient:    "lb-experts@cern.ch",
						Parameter:    1,
					},
				},
			},
			expected: false,
		},
		//Case 15: malformed email in alarms
		{caseID: 15,
			input: api.Alias{
				AliasName: "alias.cern.ch",
				BestHosts: 1,
				External:  "no",
				Hostgroup: "aiermis",
				Alarms: []api.Alarm{
					{
						ID:           1,
						AlarmAliasID: 1,
						Alias:        "alias.cern.ch",
						Name:         "minimum",
						Recipient:    "lb-expercern.ch", //malformed
						Parameter:    1,
					},
				},
			},
			expected: false,
		},

		//Case 16: Negative parameter in alarms
		{caseID: 16,
			input: api.Alias{
				AliasName: "alias.cern.ch",
				BestHosts: 1,
				External:  "no",
				Hostgroup: "aiermis",
				Alarms: []api.Alarm{
					{
						ID:           1,
						AlarmAliasID: 1,
						Alias:        "alias.cern.ch",
						Name:         "minimum",
						Recipient:    "lb-experts@cern.ch",
						Parameter:    -1, //wrong field
					},
				},
			},
			expected: false,
		},
		//Case 17: Malformed node name
		{caseID: 17,
			input: api.Alias{
				AliasName: "seed.cern.ch",
				BestHosts: 1,
				External:  "no",
				Hostgroup: "aiermis",
				Relations: []api.Relation{
					{
						Blacklist: true,
						Node: &api.Node{
							NodeName: "*&%$", //malformed
						},
					},
				},
			},
			expected: false,
		},
		//Case 18: Malformed hostgroup in Node type
		{caseID: 18,
			input: api.Alias{
				AliasName: "seed.cern.ch",
				BestHosts: 1,
				External:  "no",
				Hostgroup: "aiermis",
				Relations: []api.Relation{
					{
						Blacklist: true,
						Node: &api.Node{
							NodeName:  "testnode.cern.ch",
							Hostgroup: "@#!?", //malformed
						},
					},
				},
			},
			expected: false,
		},
	}
	for _, tc := range testCases {
		output, e := govalidator.ValidateStruct(tc.input)
		if output != tc.expected {
			t.Errorf("Failed in TestValidation\nFAILED CASE ID:%v\nINPUT:\n%v\nEXPECTED:\n%v\nRECEIVED:\n%v\n", tc.caseID, tc.input, tc.expected, output)
			t.Error(e)
		}
	}
}
