package hcl

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHCL_Unmarshal(t *testing.T) {

	hclParserWithFlatten := Parser(true)
	hclParserWithoutFlatten := Parser(false)

	testCases := []struct {
		name     string
		input    []byte
		output   map[string]interface{}
		isErr    bool
		function HCL
	}{
		{
			name:     "Empty HCL - With faltten",
			input:    []byte(`{}`),
			function: *hclParserWithFlatten,
			output:   map[string]interface{}{},
		},
		{
			name:     "Empty HCL - Without flatten",
			input:    []byte(`{}`),
			function: *hclParserWithoutFlatten,
			output:   map[string]interface{}{},
		},
		{
			name: "Valid HCL - With faltten",
			input: []byte(`resource "aws_instance" "example" {
				count = 2 # meta-argument first
				ami           = "abc123"
				instance_type = "t2.micro"
				lifecycle { # meta-argument block last
				  create_before_destroy = true
				}
			  }`),
			function: *hclParserWithFlatten,
			output: map[string]interface{}{
				"resource": map[string]interface{}{
					"aws_instance": map[string]interface{}{
						"example": map[string]interface{}{
							"ami":           "abc123",
							"count":         2,
							"instance_type": "t2.micro",
							"lifecycle": map[string]interface{}{
								"create_before_destroy": true,
							},
						},
					},
				},
			},
		},
		{
			name: "Valid HCL - Without faltten",
			input: []byte(`resource "aws_instance" "example" {
				count = 2 # meta-argument first
				ami           = "abc123"
				instance_type = "t2.micro"
				lifecycle { # meta-argument block last
				  create_before_destroy = true
				}
			  }`),
			function: *hclParserWithoutFlatten,
			output: map[string]interface{}{
				"resource": []map[string]interface{}{{
					"aws_instance": []map[string]interface{}{{
						"example": []map[string]interface{}{{
							"ami":           "abc123",
							"count":         2,
							"instance_type": "t2.micro",
							"lifecycle": []map[string]interface{}{{
								"create_before_destroy": true},
							},
						}},
					}},
				}},
			},
		},
		{
			name: "Invalid HCL - With missing parenthesis",
			input: []byte(`resource "aws_instance" "example" {
				ami = "abc123"
				`),
			function: *hclParserWithFlatten,
			isErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.function.Unmarshal(tc.input)
			if tc.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tc.output, out)
			}
		})
	}
}

func TestHCL_Marshal(t *testing.T) {

	hclParserWithFlatten := Parser(true)

	testCases := []struct {
		name     string
		input    map[string]interface{}
		output   []byte
		isErr    bool
		function HCL
	}{
		{
			name:     "Empty HCL",
			input:    map[string]interface{}{},
			output:   []byte(nil),
			function: *hclParserWithFlatten,
		},
		{
			name: "Valid HCL",
			input: map[string]interface{}{
				"resource": map[string]interface{}{
					"aws_instance": map[string]interface{}{
						"example": map[string]interface{}{
							"ami": "abc123"}}}},
			output: []byte(`"resource" "aws_instance" "example" {
  "ami" = "abc123"
}`),
		},
		{
			name: "Complex HCL",
			input: map[string]interface{}{
				"resource": []map[string]interface{}{{
					"aws_instance": []map[string]interface{}{{
						"example": []map[string]interface{}{{
							"ami":           "abc123",
							"count":         2,
							"instance_type": "t2.micro",
							"lifecycle": []map[string]interface{}{{
								"create_before_destroy": true},
							},
						}},
					}},
				}},
			},
			output: []byte(`"resource" = {
  "aws_instance" = {
    "example" = {
      "ami" = "abc123"

      "count" = 2

      "instance_type" = "t2.micro"

      "lifecycle" = {
        "create_before_destroy" = true
      }
    }
  }
}`),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.function.Marshal(tc.input)
			if tc.isErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				fmt.Printf("string(out): %v\n", string(out))
				assert.Equal(t, tc.output, out)
			}
		})
	}
}
