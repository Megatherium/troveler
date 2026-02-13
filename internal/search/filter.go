// Package search provides search functionality with filter parsing.
package search

import (
	"fmt"
	"strings"
	"unicode"

	"troveler/db"
)

// Parser parses filter expressions
type Parser struct {
	tokens []token
	pos    int
}

type token struct {
	Type  tokenType
	Value string
}

type tokenType int

const (
	tokenField tokenType = iota
	tokenOperator
	tokenValue
	tokenLParen
	tokenRParen
	tokenNot
)

// ParseFilters parses a query string into a db.Filter AST
func ParseFilters(query string) (*db.Filter, string, error) {
	// Check if there are any field=value patterns
	if !strings.Contains(query, "=") {
		// No filters, return the query as-is
		return nil, query, nil
	}

	p := &Parser{}
	p.tokenize(query)
	if len(p.tokens) == 0 {
		return nil, query, nil
	}

	ast, err := p.parseExpression()
	if err != nil {
		return nil, query, err
	}

	// Extract search term from filters
	searchTerm := p.extractSearchTerm(ast)

	return ast, searchTerm, nil
}

func (p *Parser) tokenize(query string) {
	// Simple tokenizer - scan for patterns in order
	var tokens []token
	i := 0
	quotes := false

	for i < len(query) {
		r := rune(query[i])

		// Skip spaces but add as separators
		if unicode.IsSpace(r) {
			i++

			continue
		}

		// Handle quoted values
		if r == '"' || r == '\'' {
			quotes = !quotes
			i++

			continue
		}

		if quotes {
			i++

			continue
		}

		switch r {
		case '&':
			tokens = append(tokens, token{Type: tokenOperator, Value: "&"})
			i++
		case '|':
			tokens = append(tokens, token{Type: tokenOperator, Value: "|"})
			i++
		case '!':
			tokens = append(tokens, token{Type: tokenNot, Value: "!"})
			i++
		case '(':
			tokens = append(tokens, token{Type: tokenLParen})
			i++
		case ')':
			tokens = append(tokens, token{Type: tokenRParen})
			i++
		case '=':
			tokens = append(tokens, token{Type: tokenOperator, Value: "="})
			i++
		default:
			// Collect field or value
			var value strings.Builder
			for i < len(query) {
				r := rune(query[i])
				if unicode.IsSpace(r) || r == '&' || r == '|' || r == '(' || r == ')' || r == '=' {
					break
				}
				value.WriteRune(r)
				i++
			}

			// Determine if this is a field or value
			if len(tokens) > 0 {
				lastToken := tokens[len(tokens)-1]
				if lastToken.Type == tokenOperator && lastToken.Value == "=" {
					tokens = append(tokens, token{Type: tokenValue, Value: value.String()})
				} else {
					tokens = append(tokens, token{Type: tokenField, Value: value.String()})
				}
			} else {
				tokens = append(tokens, token{Type: tokenField, Value: value.String()})
			}
		}
	}

	p.tokens = tokens
}

func (p *Parser) parseExpression() (*db.Filter, error) {
	return p.parseOr()
}

func (p *Parser) parseOr() (*db.Filter, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.pos < len(p.tokens) && p.tokens[p.pos].Value == "|" {
		p.pos++
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &db.Filter{
			Type:  db.FilterOr,
			Left:  left,
			Right: right,
		}
	}

	return left, nil
}

func (p *Parser) parseAnd() (*db.Filter, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}

	for p.pos < len(p.tokens) && p.tokens[p.pos].Value == "&" {
		p.pos++
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = &db.Filter{
			Type:  db.FilterAnd,
			Left:  left,
			Right: right,
		}
	}

	return left, nil
}

func (p *Parser) parseNot() (*db.Filter, error) {
	if p.pos < len(p.tokens) && p.tokens[p.pos].Type == tokenNot {
		p.pos++
		operand, err := p.parseNot()
		if err != nil {
			return nil, err
		}

		return &db.Filter{
			Type: db.FilterNot,
			Left: operand,
		}, nil
	}

	return p.parseTerm()
}

func (p *Parser) parseTerm() (*db.Filter, error) {
	if p.pos >= len(p.tokens) {
		return nil, fmt.Errorf("unexpected end of input")
	}

	// Handle parentheses
	if p.tokens[p.pos].Type == tokenLParen {
		p.pos++
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if p.pos >= len(p.tokens) || p.tokens[p.pos].Type != tokenRParen {
			return nil, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++

		return expr, nil
	}

	// Handle field=value
	if p.pos+2 < len(p.tokens) &&
		p.tokens[p.pos].Type == tokenField &&
		p.tokens[p.pos+1].Type == tokenOperator &&
		p.tokens[p.pos+1].Value == "=" &&
		p.tokens[p.pos+2].Type == tokenValue {
		field := p.tokens[p.pos].Value
		value := p.tokens[p.pos+2].Value
		p.pos += 3

		return &db.Filter{
			Type:  db.FilterField,
			Field: field,
			Value: value,
		}, nil
	}

	return nil, fmt.Errorf("expected field=value or parenthesis at position %d", p.pos)
}

func (p *Parser) extractSearchTerm(ast *db.Filter) string {
	// Collect values that don't have a field filter (pure search terms)
	var terms []string
	p.collectSearchTerms(ast, &terms)
	return strings.Join(terms, " ")
}

func (p *Parser) collectSearchTerms(f *db.Filter, terms *[]string) { //nolint:unparam
	if f == nil {
		return
	}

	// For AND/OR/NOT, recursively collect from children
	if f.Type == db.FilterAnd || f.Type == db.FilterOr || f.Type == db.FilterNot {
		if f.Left != nil {
			p.collectSearchTerms(f.Left, terms)
		}
		if f.Right != nil {
			p.collectSearchTerms(f.Right, terms)
		}
	}
}
