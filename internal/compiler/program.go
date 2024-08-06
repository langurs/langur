// langur/compiler/program.go

package compiler

import (
	"fmt"
	"langur/ast"
	"langur/opcode"
	"langur/str"
)

const moduleorder = "module/imports/modes/declarations"
const nonmoduleorder = "imports/expressions"

func (c *Compiler) compileProgram(node *ast.Program, executeModule bool) (
	ins opcode.Instructions, err error) {

	var bSlc []byte

	importsDone := false

	for i, s := range node.Statements {
		switch n := s.(type) {
		case *ast.ModuleNode:
			if i == 0 {
				if n.Name != "" {
					err = c.makeErr(node, "No name expected on module")
					return
				}

				c.moduleDeclaredImpureEffects = n.ImpureEffects

				// A module has a more defined structure that must be followed.
				return c.compileModule(node.Statements[1:], executeModule)

			} else {
				// not first node; an error
				err = c.makeErr(node, "Module must be first part of code to compile as module")
				return
			}

		case *ast.ImportNode:
			if importsDone {
				err = c.makeErr(node, fmt.Sprintf("Instructions out of required order; expected %s", nonmoduleorder))
				return
			}
			bSlc, err = c.compileNodeWithPopIfExprStmt(s)
			if err != nil {
				return
			}
			ins = append(ins, bSlc...)

		default:
			// not a module or import node
			importsDone = true
			bSlc, err = c.compileNodeWithPopIfExprStmt(s)
			if err != nil {
				return
			}
			ins = append(ins, bSlc...)
		}
	}
	return
}

func (c *Compiler) compileModule(nodes []ast.Node, execute bool) (
	ins opcode.Instructions, err error) {

	var modes []*ast.ModeNode
	var modeNames []string
	var declarations []*ast.LineDeclarationNode
	var imports []*ast.ImportNode
	var bytes opcode.Instructions

	importsDone := false
	modesDone := false

	for _, s := range nodes {
		switch node := s.(type) {
		case *ast.ImportNode:
			if importsDone {
				err = c.makeErr(node, fmt.Sprintf("Instructions out of required order; expected %s", moduleorder))
				return
			}
			imports = append(imports, node)

		case *ast.ModeNode:
			if modesDone {
				err = c.makeErr(node, fmt.Sprintf("Instructions out of required order; expected %s", moduleorder))
				return
			}
			importsDone = true

			if str.IsInSlice(node.Name, modeNames) {
				err = c.makeErr(node, fmt.Sprintf("Repeat of mode setting for %s", node.Name))
				return
			}
			modes = append(modes, node)
			modeNames = append(modeNames, node.Name)

		case *ast.ExpressionStatementNode:
			decl, ok := node.Expression.(*ast.LineDeclarationNode)
			if !ok {
				err = c.makeErr(node, "Expected declarations only; cannot use other expressions in module context")
				return
			}
			importsDone = true
			modesDone = true

			// if possible, split up multi-variable assignments (or "flatten")
			var flatten []*ast.LineDeclarationNode
			flatten, err = ast.FlattenDeclaration(decl)
			if err != nil {
				return
			}

			declarations = append(declarations, flatten...)

		default:
			err = c.makeErr(node, fmt.Sprintf("Expected imports/modes/declarations, not %T", node))
			return
		}
	}

	// first compile import statements
	for _, importstmt := range imports {
		bytes, err = c.compileNodeWithPopIfExprStmt(importstmt)
		if err != nil {
			return
		}
		ins = append(ins, bytes...)
	}

	// then compile mode statements
	for _, mode := range modes {
		bytes, err = c.compileNodeWithPopIfExprStmt(mode)
		if err != nil {
			return
		}
		ins = append(ins, bytes...)
	}

	// last of all, compile declarations
	declarations, err = c.fixModuleDeclarations(declarations)
	if err != nil {
		return
	}
	for _, decl := range declarations {
		bytes, err = c.compileNodeWithPopIfExprStmt(decl)
		if err != nil {
			return
		}
		ins = append(ins, bytes...)
	}

	if execute {
		bytes, err = c.compileNodeWithPopIfExprStmt(ast.ExecuteMain)
		if err != nil {
			return
		}
		ins = append(ins, bytes...)
	}

	if c.impureEffects && !c.moduleDeclaredImpureEffects {
		err = c.makeErr(nodes[0], "Module contains impure effects and is not declared impure; use module* (with asterisk)")
	}

	return
}

func (c *Compiler) fixModuleDeclarations(declarations []*ast.LineDeclarationNode) (
	decl []*ast.LineDeclarationNode, err error) {

	L := len(declarations)
	decl = make([]*ast.LineDeclarationNode, L)

	for i := range declarations {
		if declarations[i].Mutable {
			// disallow module var declarations for now
			// Within a function, the variables would be closures.
			// Since we can't mutate them anyway, allowing mutable declarations is confusing.
			return nil, c.makeErr(declarations[i], "Cannot use var declarations in module context")
		}

		// fix function declarations that won't pass as system functions
		a, ok := declarations[i].Assignment.(*ast.AssignmentNode)
		if ok {
			if len(a.Identifiers) == 1 {
				id, ok := a.Identifiers[0].(*ast.IdentNode)
				if ok {
					switch id.Name {
					case "_main":
						// set to system to make it work
						declarations[i].Assignment.(*ast.AssignmentNode).Identifiers[0].(*ast.IdentNode).System = true

						// verify not mutable and is a function
						if declarations[i].Mutable {
							return decl, c.makeErr(declarations[i], fmt.Sprintf("%s must be immutable declaration", id.Name))
						}
						_, isFunction := a.Values[0].(*ast.FunctionNode)
						if !isFunction {
							return decl, c.makeErr(declarations[i], fmt.Sprintf("%s must be a function", id.Name))
						}
					}
				}
			}
		}

		// reverse order to put dependent functions at top
		decl[L-i-1] = declarations[i]
	}

	return decl, nil
}
