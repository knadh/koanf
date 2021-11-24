package structs

import (
	"reflect"
	"testing"
)

type parentStruct struct {
	Name   string      `koanf:"name"`
	ID     int         `koanf:"id"`
	Child1 childStruct `koanf:"child1"`
}
type childStruct struct {
	Name        string            `koanf:"name"`
	Type        string            `koanf:"type"`
	Empty       map[string]string `koanf:"empty"`
	Grandchild1 grandchildStruct  `koanf:"grandchild1"`
}
type grandchildStruct struct {
	Ids []int `koanf:"ids"`
	On  bool  `koanf:"on"`
}
type testStruct struct {
	Type    string            `koanf:"type"`
	Empty   map[string]string `koanf:"empty"`
	Parent1 parentStruct      `koanf:"parent1"`
}

type testStructWithDelim struct {
	Endpoint string `koanf:"conf_endpoint"`
	Username string `koanf:"conf_creds.username"`
	Password string `koanf:"conf_creds.password"`
}

func TestStructs_Read(t *testing.T) {
	type fields struct {
		s     interface{}
		tag   string
		delim string
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "read",
			fields: fields{
				s: testStruct{
					Type:  "json",
					Empty: make(map[string]string),
					Parent1: parentStruct{
						Name: "parent1",
						ID:   1234,
						Child1: childStruct{
							Name:  "child1",
							Type:  "json",
							Empty: make(map[string]string),
							Grandchild1: grandchildStruct{
								Ids: []int{1, 2, 3},
								On:  true,
							},
						},
					},
				},
				tag: "koanf",
			},
			want: map[string]interface{}{
				"type":  "json",
				"empty": map[string]string{},
				"parent1": map[string]interface{}{
					"child1": map[string]interface{}{
						"empty": map[string]string{},
						"type":  "json",
						"name":  "child1",
						"grandchild1": map[string]interface{}{
							"on":  true,
							"ids": []int{1, 2, 3},
						},
					},
					"name": "parent1",
					"id":   1234,
				},
			},
			wantErr: false,
		},
		{
			name: "read delim struct",
			fields: fields{
				s: testStructWithDelim{
					Endpoint: "test_endpoint",
					Username: "test_username",
					Password: "test_password",
				},
				tag:   "koanf",
				delim: ".",
			},
			want: map[string]interface{}{
				"conf_creds": map[string]interface{}{
					"password": "test_password",
					"username": "test_username",
				},
				"conf_endpoint": "test_endpoint",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Structs{
				s:     tt.fields.s,
				tag:   tt.fields.tag,
				delim: tt.fields.delim,
			}
			got, err := s.Read()
			if (err != nil) != tt.wantErr {
				t.Errorf("Structs.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Structs.Read() = \n%#v\nwant \n%#v\n", got, tt.want)
			}
		})
	}
}
