package main

import (
	"encoding/csv"
	"reflect"
	"strings"
	"testing"
)

func TestReadPlayerHeaderAndRows(t *testing.T) {
	input := "Shirt; Name ;Id\n10;Hong Myung-Bo; 91 \n9; Kitazawa ;141\n"
	cr := csv.NewReader(strings.NewReader(input))
	cr.Comma = ';'

	idIdx, nameIdx := readPlayerHeader(cr)
	if idIdx != 2 || nameIdx != 1 {
		t.Fatalf("header indexes = (%d, %d), want (2, 1)", idIdx, nameIdx)
	}

	got := readPlayerRows(cr, idIdx, nameIdx)
	want := []Player{
		{ID: "91", Name: "Hong Myung-Bo"},
		{ID: "141", Name: "Kitazawa"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("rows = %+v, want %+v", got, want)
	}
}
