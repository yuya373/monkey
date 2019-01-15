package ast

import (
	"bytes"
	"github.com/yuya373/monkey/token"
	"strings"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

func (p *Program) TokenLiteral() string {
	if (len(p.Statements)) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (s *ReturnStatement) statementNode()       {}
func (s *ReturnStatement) TokenLiteral() string { return s.Token.Literal }
func (s *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(s.TokenLiteral() + " ")
	if s.ReturnValue != nil {
		out.WriteString(s.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

type ExpressionStatement struct {
	Token      token.Token
	Expression Expression
}

func (s *ExpressionStatement) statementNode()       {}
func (s *ExpressionStatement) TokenLiteral() string { return s.Token.Literal }
func (s *ExpressionStatement) String() string {
	var out bytes.Buffer
	if s.Expression != nil {
		out.WriteString(s.Expression.String())
	}
	return out.String()
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (s *IntegerLiteral) expressionNode()      {}
func (s *IntegerLiteral) TokenLiteral() string { return s.Token.Literal }
func (s *IntegerLiteral) String() string       { return s.Token.Literal }

type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (exp *PrefixExpression) expressionNode()      {}
func (exp *PrefixExpression) TokenLiteral() string { return exp.Token.Literal }
func (exp *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(exp.Operator)
	out.WriteString(exp.Right.String())
	out.WriteString(")")
	return out.String()
}

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (exp *InfixExpression) expressionNode()      {}
func (exp *InfixExpression) TokenLiteral() string { return exp.Token.Literal }
func (exp *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(exp.Left.String())
	out.WriteString(" " + exp.Operator + " ")
	out.WriteString(exp.Right.String())
	out.WriteString(")")

	return out.String()
}

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (s *BlockStatement) statementNode()       {}
func (s *BlockStatement) TokenLiteral() string { return s.Token.Literal }
func (s *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range s.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type IfExpression struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (s *IfExpression) expressionNode()      {}
func (s *IfExpression) TokenLiteral() string { return s.Token.Literal }
func (s *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString("(")
	out.WriteString(s.Condition.String())
	out.WriteString(")")
	out.WriteString("{")
	out.WriteString(s.Consequence.String())
	out.WriteString("}")

	if s.Alternative != nil {
		out.WriteString("else")
		out.WriteString("{")
		out.WriteString(s.Alternative.String())
		out.WriteString("}")
	}

	return out.String()
}

type FunctionLiteral struct {
	Token      token.Token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (l *FunctionLiteral) expressionNode()      {}
func (l *FunctionLiteral) TokenLiteral() string { return l.Token.Literal }
func (l *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range l.Parameters {
		params = append(params, p.String())
	}

	out.WriteString(l.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	out.WriteString("{")
	out.WriteString(l.Body.String())
	out.WriteString("}")

	return out.String()
}

type CallExpression struct {
	Token     token.Token
	Function  Expression // Identifier or FunctionLiteral
	Arguments []Expression
}

func (exp *CallExpression) expressionNode()      {}
func (exp *CallExpression) TokenLiteral() string { return exp.Token.Literal }
func (exp *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range exp.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(exp.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (l *StringLiteral) expressionNode()      {}
func (l *StringLiteral) TokenLiteral() string { return l.Token.Literal }
func (l *StringLiteral) String() string       { return l.Token.Literal }

type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (l *ArrayLiteral) expressionNode()      {}
func (l *ArrayLiteral) TokenLiteral() string { return l.Token.Literal }
func (l *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range l.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (exp *IndexExpression) expressionNode()      {}
func (exp *IndexExpression) TokenLiteral() string { return exp.Token.Literal }
func (exp *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(exp.Left.String())
	out.WriteString("[")
	out.WriteString(exp.Index.String())
	out.WriteString("]")
	out.WriteString(")")

	return out.String()
}
