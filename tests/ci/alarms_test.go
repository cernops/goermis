package ci

import (
	"testing"

	"gitlab.cern.ch/lb-experts/goermis/alarms"
)

func TestCheckMinimumAlarm(t *testing.T) {
	type test struct {
		caseID   int
		input1   string
		input2   int
		expected bool
	}
	testCases := []test{
		//Case1: Correct fields
		{caseID: 1,
			input1:   "goermis.cern.ch",
			input2:   1,
			expected: false},
		//Case2: Without domain name, it should not find any
		{caseID: 2,
			input1:   "goermis",
			input2:   1,
			expected: true},
		//Case3: Number of nodes smaller than the threshold(1 node, threshold is 5)
		{caseID: 3,
			input1:   "goermis.cern.ch",
			input2:   5,
			expected: true},
		//Case4: Malformed alias
		{caseID: 4,
			input1:   "@!?>",
			input2:   5,
			expected: true},
		//Case5: 0 threshold
		{caseID: 4,
			input1:   "goermis.cern.ch",
			input2:   0,
			expected: false},
	}
	for _, tc := range testCases {
		output := alarms.CheckMinimumAlarm(tc.input1, tc.input2)
		if output != tc.expected {
			t.Errorf("Failed in TestCheckMinimumAlarm for case ID:%v\nALIAS:%v\nPARAMETER:%v\nEXPECTED:%v\nRECEIVED:%v\n", tc.caseID, tc.input1, tc.input2, tc.expected, output)
		}
	}

}

func TestSendNotification(t *testing.T) {
	alias := "test.cern.ch"
	recipient := "randomemail.cern.ch"
	name := "minimum"
	parameter := 1
	err := alarms.SendNotification(alias, recipient, name, parameter)
	if err != nil {
		t.Errorf("Failed in TestSendNotification with Alias:%v\nRecipient:%v\nName:%v\nParameter:%v\nError:%v\n", alias, recipient, name, parameter, err)
	}
}

