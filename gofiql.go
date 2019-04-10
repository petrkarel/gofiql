package gofiql

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var operators = []string{"==","=ne=","=lt=","=le=","=gt=","=ge=","=in=","=out="}
var pattern = regexp.MustCompile("(?i)^([A-Za-z0-9-\\._~]+)(" + strings.Join(operators, "|") + ")")

type Node interface {
    Convert() string
	String() string
}

type BinaryLogicalNode struct {
    operands []Node
    operator string
}

type UnaryLogicalNode struct {
    operand Node
}

type ComparisonNode struct {
    selector string
    operator string
    value string
}

func (node *BinaryLogicalNode) Convert() (str string) {
    for i, operand := range node.operands {
        if reflect.TypeOf(operand).AssignableTo(reflect.TypeOf((*BinaryLogicalNode)(nil))) {
	    	str += "(" + operand.Convert() + ")"
		} else {
	    	str += operand.Convert()
		}
		if i < len(node.operands) - 1 {
	    	str += " " + string(node.operator) + " "
		}
    }
    return
}

func (node *BinaryLogicalNode) String() (str string) {
	str = node.operator + "("
	for i, operand := range node.operands {
		str += operand.String()
		if (i < len(node.operands) - 1) {
			str += ","
		}
	}
	str += ")"
	return
}

func (node *UnaryLogicalNode) Convert() (str string) {
    str = "NOT (" + node.operand.Convert() + ")"
	return
}

func (node *UnaryLogicalNode) String() (str string) {
	str = "NOT(" + node.operand.String() + ")"
	return
}

func (node *ComparisonNode) Convert() (str string) {
    switch node.operator {
	case "==":
	    if strings.Contains(node.value, "*") {
		str = node.selector + " LIKE " + strings.Replace(node.value, "*", "%", -1)
	    } else {
		str = node.selector + " = " + strings.Replace(node.value, "*", "%", -1)
	    }
	case "=ne=":
	    if strings.Contains(node.value, "*") {
		str = node.selector + " NOT LIKE " + strings.Replace(node.value, "*", "%", -1)
	    } else {
		str = node.selector + " <> " + strings.Replace(node.value, "*", "%", -1)
	    }
	case "=lt=":
	    str = node.selector + " < " + node.value
	case "=le=":
	    str = node.selector + " <= " + node.value
	case "=gt=":
	    str = node.selector + " > " + node.value
	case "=ge=":
	    str = node.selector + " >= " + node.value
	case "=in=":
	    str = node.selector + " IN " + node.value
	case "=out=":
	    str = node.selector + " NOT IN " + node.value
    }
    return
}

func (node *ComparisonNode) String() (str string) {
	switch node.operator {
	case "==":
		str = "EQ(" + node.selector + "," + node.value + ")"
	case "=ne=":
		str = "NEQ(" + node.selector + "," + node.value + ")"
	case "=lt=":
		str = "LT(" + node.selector + "," + node.value + ")"
	case "=le=":
		str = "LE(" + node.selector + "," + node.value + ")"
	case "=gt=":
		str = "GT(" + node.selector + "," + node.value + ")"
	case "=ge=":
		str = "GE(" + node.selector + "," + node.value + ")"
	case "=in=":
		str = "IN(" + node.selector + "," + node.value + ")"
	case "=out=":
		str = "OUT(" + node.selector + "," + node.value + ")"
	}
	return
}

func ParseAndConvert(expr string) (str string, err error) {
	node, err := Parse(expr)
	if err != nil {
		err = errors.New("Parsing expression '" + expr + "' failed: " + err.Error())
	} else {
		str = node.Convert()
	}
	return
}

func Parse(expr string) (Node, error) {
    var orExp, andExp, notExp bool
    level := 0
    for i, c := range expr {
		if c == '(' {
			level++
		} else if c == ')' {
			level--
			if level < 0 {
			return nil, errors.New("Unexpected closing bracket at position " + strconv.Itoa(i))
			}
		}
		if level == 0 {
			if c == ',' {
			orExp = true
			} else if c == ';' {
			andExp = true
			} else if c == '!' {
			notExp = true
			}
		}
    }
    if level > 0 {
		return nil, errors.New("Unexpected opening bracket in expression " + expr)
    }

    if orExp {
        node := new(BinaryLogicalNode)
		node.operator = "OR"
		for _, frg := range parseFragments(expr, ',') {
			result, err := Parse(frg)
			if err != nil {
				return nil, err
			}
			node.operands = append(node.operands, result)
		}
		if len(node.operands) == 1 {
			return node.operands[0], nil
		}
		return node, nil
    } else if andExp {
		node := new(BinaryLogicalNode)
		node.operator = "AND"
		for _, frg := range parseFragments(expr, ';') {
			result, err := Parse(frg)
			if err != nil {
				return nil, err
			}
			node.operands = append(node.operands, result)
		}
		if len(node.operands) == 1 {
			return node.operands[0], nil
		}
		return node, nil
    } else if notExp {
		node := new(UnaryLogicalNode)
		result, err := Parse(expr[1:])
		if err != nil {
			return nil, err
		}
		node.operand = result
		return node, nil
    } else if expr[0] == '(' {
		return Parse(expr[1:len(expr)-1])
    } else {
		return parseComparisonNode(expr)
    }
}

func parseFragments(str string, op rune) (fragments []string) {
	var last,level int
    for i,c := range str {
		if (c == '(') {
			level++
		} else if (c == ')') {
			level--
		}
		if level == 0 && c == op {
			fragment := str[last:i]
			if len(fragment) > 1 {
				fragments = append(fragments, fragment)
			}
			last = i + 1
		}
		if i == len(str) - 1 {
			fragment := str[last:]
			if len(fragment) > 1 {
				fragments = append(fragments, fragment)
			}
		}
    }
    return
}

func parseComparisonNode(str string) (Node, error) {
    if !pattern.MatchString(str) {
		return nil, errors.New("not a comparison expression: " + str)
    }
    result := pattern.FindAllStringSubmatch(str, -1)
    selector := result[0][1]
    operator := strings.ToLower(result[0][2])
    value := str[len(selector+operator):]
    return &ComparisonNode{selector, operator, value}, nil
}