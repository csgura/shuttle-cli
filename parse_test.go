package main_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/csgura/fp/should"
	"github.com/csgura/fp/slice"
	"mvdan.cc/sh/v3/syntax"
)

func sprintword(v *syntax.Word) string {
	sb := &strings.Builder{}
	syntax.NewPrinter().Print(sb, v)
	return sb.String()
}

func sprintnode(v syntax.Node) string {
	sb := &strings.Builder{}
	syntax.NewPrinter().Print(sb, v)
	return sb.String()
}

func TestSyntax(t *testing.T) {
	p := syntax.NewParser()
	s, err := p.Parse(strings.NewReader("xtitle 'DB2 (172.10.3.10)' && SSHPASS='alti.123' sshpass -e ssh -oStrictHostKeyChecking=no altibase@172.10.3.10"), "hello")
	should.BeNil(t, err)

	//syntax.DebugPrint(os.Stdout, s)
	syntax.Walk(s, func(n syntax.Node) bool {
		switch x := n.(type) {
		case *syntax.CallExpr:
			fmt.Printf("command %s, args : %v\n", x.Args[0].Lit(), slice.Map(x.Args[1:], sprintword))
		case *syntax.Assign:
			fmt.Printf("%s=%s\n", x.Name.Value, sprintword(x.Value))
		default:
			//fmt.Printf("node type = %T, value = %s\n", n, sprintnode(n))
		}
		return true
	})
}
