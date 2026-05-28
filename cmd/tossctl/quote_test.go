package main

import (
	"reflect"
	"testing"
)

func TestParseBatchSymbols(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want []string
	}{
		{
			name: "single arg",
			args: []string{"삼성전자"},
			want: []string{"삼성전자"},
		},
		{
			name: "space-separated args",
			args: []string{"삼성전자", "KB금융"},
			want: []string{"삼성전자", "KB금융"},
		},
		{
			name: "comma-separated single arg",
			args: []string{"삼성전자,KB금융"},
			want: []string{"삼성전자", "KB금융"},
		},
		{
			name: "mixed comma + space",
			args: []string{"삼성전자,KB금융", "현대차"},
			want: []string{"삼성전자", "KB금융", "현대차"},
		},
		{
			name: "trims whitespace around commas",
			args: []string{"삼성전자 , KB금융 ,  현대차"},
			want: []string{"삼성전자", "KB금융", "현대차"},
		},
		{
			name: "drops empty tokens",
			args: []string{"삼성전자,,KB금융,"},
			want: []string{"삼성전자", "KB금융"},
		},
		{
			name: "all-empty stays empty",
			args: []string{",", " , ", ""},
			want: nil,
		},
	}
	for _, c := range cases {
		got := parseBatchSymbols(c.args)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%s: parseBatchSymbols(%#v) = %#v, want %#v", c.name, c.args, got, c.want)
		}
	}
}
