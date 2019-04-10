package gofiql

import (
	"testing"
)

func TestParse(t *testing.T) {

	var test = []struct {
		expression string
		result string
	}{
		{"!(sel1==arg1)", "NOT(EQ(sel1,arg1))"},
		{"sel1=out=(arg1,arg2)", "OUT(sel1,(arg1,arg2))"},
		{";,;,((sel1==arg1))", "EQ(sel1,arg1)"},
		{",,sel1==*arg1*", "EQ(sel1,*arg1*)"},
		{"sel1=lt=arg1,sel2=le=arg2,sel3=gt=arg3,sel4=ge=arg4,sel5=in=(1,2,3)", "OR(LT(sel1,arg1),LE(sel2,arg2),GT(sel3,arg3),GE(sel4,arg4),IN(sel5,(1,2,3)))"},
		{"sel1==arg1;sel2=ne=arg3", "AND(EQ(sel1,arg1),NEQ(sel2,arg3))"},
		{"(sel1==arg1,sel2=lt=arg2);sel3=gt=arg3", "AND(OR(EQ(sel1,arg1),LT(sel2,arg2)),GT(sel3,arg3))"},
		{"sel1==arg1,sel2==arg2;sel3==arg3", "OR(EQ(sel1,arg1),AND(EQ(sel2,arg2),EQ(sel3,arg3)))"},
	}

	for _, row := range test {
		node, err := Parse(row.expression)
		if (err != nil) {
			t.Errorf("Parse (%s) got error: %+v", row.expression, err)
			continue
		}
		if node.String() != row.result {
			t.Errorf("Expected %s, but actual is %s", row.result, node.String())
			continue
		}
	}
}

func TestParseAndConvert(t *testing.T) {

	var test = []struct {
		expression string
		result string
	}{
		{"!(sel1==arg1)", "NOT (sel1 = arg1)"},
		{"sel1=out=(arg1,arg2)", "sel1 NOT IN (arg1,arg2)"},
		{";,;,((sel1==arg1))", "sel1 = arg1"},
		{",,sel1==*arg1*", "sel1 LIKE %arg1%"},
		{"sel1==arg1;sel2=ne=arg3", "sel1 = arg1 AND sel2 <> arg3"},
		{"(sel1==arg1,sel2=le=arg2);(sel3=ge=arg3)", "(sel1 = arg1 OR sel2 <= arg2) AND sel3 >= arg3"},
		{"sel1==*arg1*,sel2=ne=arg2*", "sel1 LIKE %arg1% OR sel2 NOT LIKE arg2%"},
		{"(sel1==arg1,sel2=lt=arg2);sel3=gt=arg3", "(sel1 = arg1 OR sel2 < arg2) AND sel3 > arg3"},
		{"(sel1==arg1;sel2=ne=arg2);(sel3=le=arg3,sel4=out=(1,2,3),sel5=ge=arg3)", "(sel1 = arg1 AND sel2 <> arg2) AND (sel3 <= arg3 OR sel4 NOT IN (1,2,3) OR sel5 >= arg3)"},
	}

	for _, row := range test {
		str, err := ParseAndConvert(row.expression)
		if (err != nil) {
			t.Errorf("Parse (%s) got error: %+v", row.expression, err)
			continue
		}
		if str != row.result {
			t.Errorf("Expected %s, but actual is %s", row.result, str)
			continue
		}
	}
}
