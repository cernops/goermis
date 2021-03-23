package ci

import (
	"reflect"
	"sort"
	"testing"

	"gitlab.cern.ch/lb-experts/goermis/auth"
)

func TestGetPwnAndGetCud(t *testing.T) {
	type test struct {
		input        string
		expectedPwn  []string
		expectedLdap bool
	}
	testCases := []test{
		{input: "ermists",
			expectedPwn:  []string{"aiermis", "ailbd"},
			expectedLdap: false,
		},
		{input: "ermistst",
			expectedPwn:  []string{"aiermis", "ailbd"},
			expectedLdap: true,
		},
	}
	for _, tc := range testCases {
		outputPwn := auth.GetPwn(tc.input)
		outputLdap := auth.CheckCud(tc.input)
		sort.Strings(outputPwn)
		sort.Strings(tc.expectedPwn)
		if !reflect.DeepEqual(outputPwn, tc.expectedPwn) {
			t.Errorf("Failed in the first subtest of TestGetPwnAndGetLdap\nExpected:%v\nReceived:%v\nInput:%v\n", tc.expectedPwn, outputPwn, tc.input)
		}
		if outputLdap != tc.expectedLdap {
			t.Errorf("Failed in the second subtest of TestGetPwnAndGetLdap\nExpected:%v\nReceived:%v\nInput:%v\n", tc.expectedLdap, outputLdap, tc.input)
		}
	}

}
