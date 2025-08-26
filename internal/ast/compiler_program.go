// langur/ast/compiler_program.go

package ast

import (
	"fmt"
	"langur/common"
	"langur/opcode"
	"langur/str"
)

const moduleorder = "module/imports/modes/declarations"
const nonmoduleorder = "imports/expressions"

func (c *Compiler) compileProgram(node *Program, executeModule bool) (
	pkg opcode.InsPackage, err error) {

	var bSlc opcode.InsPackage

	importsDone := false

	for i, s := range node.Statements {
		switch n := s.(type) {
		case *ModuleNode:
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

		case *ImportNode:
			if importsDone {
				err = c.makeErr(node, fmt.Sprintf("Instructions out of required order; expected %s", nonmoduleorder))
				return
			}
			bSlc, err = c.compileNodeWithPopIfExprStmt(s)
			if err != nil {
				return
			}
			pkg = pkg.Append(bSlc)

		case nil:
			err = c.makeErr(node, "Unexpected nil node")
			return

		default:
			// not a module or import node
			importsDone = true
			bSlc, err = c.compileNodeWithPopIfExprStmt(s)
			if err != nil {
				return
			}
			pkg = pkg.Append(bSlc)
		}
	}
	return
}

func (c *Compiler) compileModule(nodes []Node, execute bool) (
	pkg opcode.InsPackage, err error) {

	var modes []*ModeNode
	var modeNames []string
	var declarations []*LineDeclarationNode
	var imports []*ImportNode
	var bytes opcode.InsPackage

	importsDone := false
	modesDone := false

	for _, s := range nodes {
		switch node := s.(type) {
		case *ImportNode:
			if importsDone {
				err = c.makeErr(node, fmt.Sprintf("Instructions out of required order; expected %s", moduleorder))
				return
			}
			imports = append(imports, node)

		case *ModeNode:
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

		case *ExpressionStatementNode:
			decl, ok := node.Expression.(*LineDeclarationNode)
			if !ok {
				err = c.makeErr(node, "Expected declarations only; cannot use other expressions in module context")
				return
			}
			importsDone = true
			modesDone = true

			// if possible, split up multi-variable assignments (or "flatten")
			var flatten []*LineDeclarationNode
			flatten, err = FlattenDeclaration(decl)
			if err != nil {
				return
			}

			declarations = append(declarations, flatten...)

		case nil:
			err = c.makeErr(node, "Unexpected nil node")
			return

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
		pkg = pkg.Append(bytes)
	}

	// then compile mode statements
	for _, mode := range modes {
		bytes, err = c.compileNodeWithPopIfExprStmt(mode)
		if err != nil {
			return
		}
		pkg = pkg.Append(bytes)
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
		pkg = pkg.Append(bytes)
	}

	if execute {
		bytes, err = c.compileNodeWithPopIfExprStmt(ExecuteMain)
		if err != nil {
			return
		}
		pkg = pkg.Append(bytes)
	}

	if c.impureEffects && !c.moduleDeclaredImpureEffects {
		err = c.makeErr(nodes[0], "Module contains impure effects and is not declared impure; use module* (with asterisk)")
	}

	return
}

func (c *Compiler) fixModuleDeclarations(declarations []*LineDeclarationNode) (
	decl []*LineDeclarationNode, err error) {

	L := len(declarations)
	decl = make([]*LineDeclarationNode, L)

	for i := range declarations {
		if declarations[i].Mutable {
			// disallow module var declarations for now
			// Within a function, the variables would be closures.
			// Since we can't mutate them anyway, allowing mutable declarations is confusing.
			return nil, c.makeErr(declarations[i], "Cannot use var declarations in module context in this implementation/version")
		}

		// fix function declarations that won't pass as system functions
		a, ok := declarations[i].Assignment.(*AssignmentNode)
		if ok {
			if len(a.Identifiers) == 1 {
				id, ok := a.Identifiers[0].(*IdentNode)
				if ok {
					switch id.Name {
					case common.MainFnName:
						// set to system to make it work
						declarations[i].Assignment.(*AssignmentNode).Identifiers[0].(*IdentNode).System = true

						// verify not mutable and is a function
						if declarations[i].Mutable {
							return decl, c.makeErr(declarations[i], fmt.Sprintf("%s must be immutable declaration", id.Name))
						}
						_, isFunction := a.Values[0].(*FunctionNode)
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
